package benchmark

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"
)

type Config struct {
	GatewayURL           string
	DirectURL            string
	APIKey               string
	WarmupIterations     int
	MeasuredIterations   int
	Concurrency          int
	Repeat               int
	Command              string
	SourceCommit         string
	CleanTree            bool
	ImageRef             string
	ImageDigest          string
	DependencyLockDigest string
	Producer             string
	CIRunURL             string
	HardwareClass        string
}

type Result struct {
	SchemaVersion    int         `json:"schema_version"`
	RunID            string      `json:"run_id"`
	Project          string      `json:"project"`
	BenchmarkID      string      `json:"benchmark_id"`
	Workload         Workload    `json:"workload"`
	Metrics          []Metric    `json:"metrics"`
	Execution        Execution   `json:"execution"`
	Environment      Environment `json:"environment"`
	Provenance       Provenance  `json:"provenance"`
	ComparabilityKey string      `json:"comparability_key"`
}

type Workload struct {
	Version            string `json:"version"`
	FixtureDigest      string `json:"fixture_digest"`
	ConfigDigest       string `json:"config_digest"`
	WarmupIterations   int    `json:"warmup_iterations"`
	MeasuredIterations int    `json:"measured_iterations"`
	Concurrency        int    `json:"concurrency"`
}

type Metric struct {
	Name      string         `json:"name"`
	Value     float64        `json:"value"`
	Unit      string         `json:"unit"`
	Direction string         `json:"direction"`
	Samples   []float64      `json:"samples"`
	Failures  int            `json:"failures"`
	Summary   map[string]any `json:"summary"`
}

type Execution struct {
	Command         string  `json:"command"`
	StartedAt       string  `json:"started_at"`
	DurationSeconds float64 `json:"duration_seconds"`
	ExitCode        int     `json:"exit_code"`
	Repeat          int     `json:"repeat"`
}

type Environment struct {
	Runtime       string `json:"runtime"`
	Architecture  string `json:"architecture"`
	HardwareClass string `json:"hardware_class"`
}

type Provenance struct {
	SourceCommit         string `json:"source_commit"`
	CleanTree            bool   `json:"clean_tree"`
	ImageRef             string `json:"image_ref"`
	ImageDigest          string `json:"image_digest"`
	DependencyLockDigest string `json:"dependency_lock_digest"`
	Producer             string `json:"producer"`
	CIRunURL             string `json:"ci_run_url,omitempty"`
	ArtifactDigest       string `json:"artifact_digest"`
}

type endpointStats struct {
	p50        float64
	p95        float64
	p99        float64
	throughput float64
	rejects    int
	failures   int
}

func Run(ctx context.Context, cfg Config) (*Result, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	started := time.Now().UTC()
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.Concurrency * 4,
			MaxIdleConnsPerHost: cfg.Concurrency * 2,
			IdleConnTimeout:     30 * time.Second,
		},
	}
	defer client.CloseIdleConnections()

	if _, err := measure(ctx, client, cfg.DirectURL+"/echo", cfg.APIKey, cfg.WarmupIterations, cfg.Concurrency); err != nil {
		return nil, fmt.Errorf("warm direct upstream: %w", err)
	}
	if _, err := measure(ctx, client, cfg.GatewayURL+"/echo", cfg.APIKey, cfg.WarmupIterations, cfg.Concurrency); err != nil {
		return nil, fmt.Errorf("warm gateway: %w", err)
	}

	directRuns := make([]endpointStats, 0, cfg.Repeat)
	gatewayRuns := make([]endpointStats, 0, cfg.Repeat)
	for repetition := 0; repetition < cfg.Repeat; repetition++ {
		direct, err := measure(ctx, client, cfg.DirectURL+"/echo", cfg.APIKey, cfg.MeasuredIterations, cfg.Concurrency)
		if err != nil {
			return nil, fmt.Errorf("direct repetition %d: %w", repetition+1, err)
		}
		gateway, err := measure(ctx, client, cfg.GatewayURL+"/echo", cfg.APIKey, cfg.MeasuredIterations, cfg.Concurrency)
		if err != nil {
			return nil, fmt.Errorf("gateway repetition %d: %w", repetition+1, err)
		}
		directRuns = append(directRuns, direct)
		gatewayRuns = append(gatewayRuns, gateway)
	}

	metrics := buildMetrics(directRuns, gatewayRuns)
	result := &Result{
		SchemaVersion: 2,
		RunID:         newUUID(),
		Project:       "api-gateway-lite",
		BenchmarkID:   "gateway-overhead-v2",
		Workload: Workload{
			Version:            "2.0.0",
			FixtureDigest:      digest([]byte("GET /echo HTTP/1.1\nresponse=200:ok\n")),
			ConfigDigest:       digest([]byte(fmt.Sprintf("warmup=%d;measured=%d;concurrency=%d;repeat=%d", cfg.WarmupIterations, cfg.MeasuredIterations, cfg.Concurrency, cfg.Repeat))),
			WarmupIterations:   cfg.WarmupIterations,
			MeasuredIterations: cfg.MeasuredIterations,
			Concurrency:        cfg.Concurrency,
		},
		Metrics: metrics,
		Execution: Execution{
			Command:         cfg.Command,
			StartedAt:       started.Format(time.RFC3339),
			DurationSeconds: time.Since(started).Seconds(),
			ExitCode:        0,
			Repeat:          cfg.Repeat,
		},
		Environment: Environment{
			Runtime:       runtime.Version(),
			Architecture:  runtime.GOOS + "/" + runtime.GOARCH,
			HardwareClass: cfg.HardwareClass,
		},
		Provenance: Provenance{
			SourceCommit:         cfg.SourceCommit,
			CleanTree:            cfg.CleanTree,
			ImageRef:             cfg.ImageRef,
			ImageDigest:          cfg.ImageDigest,
			DependencyLockDigest: cfg.DependencyLockDigest,
			Producer:             cfg.Producer,
			CIRunURL:             cfg.CIRunURL,
			ArtifactDigest:       "sha256:" + string(make([]byte, 64)),
		},
		ComparabilityKey: fmt.Sprintf("api-gateway-lite:v2:echo:%dx%d:r%d:redis-otel", cfg.MeasuredIterations, cfg.Concurrency, cfg.Repeat),
	}

	canonical, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal benchmark artifact: %w", err)
	}
	result.Provenance.ArtifactDigest = digest(canonical)
	return result, nil
}

func validateConfig(cfg Config) error {
	if cfg.GatewayURL == "" || cfg.DirectURL == "" || cfg.APIKey == "" {
		return fmt.Errorf("gateway URL, direct URL and API key are required")
	}
	if cfg.WarmupIterations < 1 || cfg.MeasuredIterations < 1 || cfg.Concurrency < 1 || cfg.Repeat < 3 {
		return fmt.Errorf("benchmark requires warmup, positive workload/concurrency and at least 3 repetitions")
	}
	if len(cfg.SourceCommit) != 40 || !cfg.CleanTree {
		return fmt.Errorf("benchmark requires an exact 40-character commit and clean tree provenance")
	}
	for name, value := range map[string]string{
		"image digest": cfg.ImageDigest, "dependency lock digest": cfg.DependencyLockDigest,
	} {
		if len(value) != 71 || value[:7] != "sha256:" {
			return fmt.Errorf("%s must be a sha256 digest", name)
		}
	}
	if cfg.ImageRef == "" || cfg.Command == "" || cfg.HardwareClass == "" {
		return fmt.Errorf("image ref, command and hardware class are required")
	}
	if cfg.Producer != "local" && cfg.Producer != "github-actions" && cfg.Producer != "other-ci" {
		return fmt.Errorf("unsupported benchmark producer %q", cfg.Producer)
	}
	return nil
}

func measure(ctx context.Context, client *http.Client, url, apiKey string, requests, concurrency int) (endpointStats, error) {
	latencies := make([]float64, requests)
	jobs := make(chan int)
	var wait sync.WaitGroup
	var counterMu sync.Mutex
	rejects := 0
	failures := 0

	started := time.Now()
	for worker := 0; worker < concurrency; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for index := range jobs {
				request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				if err != nil {
					counterMu.Lock()
					failures++
					counterMu.Unlock()
					continue
				}
				request.Header.Set("X-API-Key", apiKey)
				requestStart := time.Now()
				response, err := client.Do(request)
				latencies[index] = float64(time.Since(requestStart).Nanoseconds()) / 1e6
				if err != nil {
					counterMu.Lock()
					failures++
					counterMu.Unlock()
					continue
				}
				_, copyErr := io.Copy(io.Discard, response.Body)
				closeErr := response.Body.Close()
				counterMu.Lock()
				if response.StatusCode == http.StatusTooManyRequests {
					rejects++
				}
				if response.StatusCode >= http.StatusBadRequest && response.StatusCode != http.StatusTooManyRequests {
					failures++
				}
				if copyErr != nil || closeErr != nil {
					failures++
				}
				counterMu.Unlock()
			}
		}()
	}
	for index := range requests {
		jobs <- index
	}
	close(jobs)
	wait.Wait()
	duration := time.Since(started).Seconds()
	if duration <= 0 {
		return endpointStats{}, fmt.Errorf("invalid measurement duration")
	}

	sort.Float64s(latencies)
	return endpointStats{
		p50:        percentile(latencies, 0.50),
		p95:        percentile(latencies, 0.95),
		p99:        percentile(latencies, 0.99),
		throughput: float64(requests) / duration,
		rejects:    rejects,
		failures:   failures,
	}, nil
}

func buildMetrics(direct, gateway []endpointStats) []Metric {
	metric := func(name, unit, direction string, samples []float64, failures int) Metric {
		sorted := append([]float64(nil), samples...)
		sort.Float64s(sorted)
		return Metric{
			Name: name, Value: percentile(sorted, 0.50), Unit: unit, Direction: direction,
			Samples: samples, Failures: failures,
			Summary: map[string]any{"min": sorted[0], "median": percentile(sorted, 0.50), "max": sorted[len(sorted)-1]},
		}
	}

	directP50, directP95, directP99, directRPS := extract(direct)
	gatewayP50, gatewayP95, gatewayP99, gatewayRPS := extract(gateway)
	overheadP50 := subtract(gatewayP50, directP50)
	overheadP95 := subtract(gatewayP95, directP95)
	overheadP99 := subtract(gatewayP99, directP99)
	gatewayRejects := counts(gateway, func(stats endpointStats) int { return stats.rejects })
	gatewayFailures := counts(gateway, func(stats endpointStats) int { return stats.failures })
	directFailures := counts(direct, func(stats endpointStats) int { return stats.failures })

	return []Metric{
		metric("overhead_p50_ms", "ms", "lower_is_better", overheadP50, sum(gatewayFailures)+sum(directFailures)),
		metric("overhead_p95_ms", "ms", "lower_is_better", overheadP95, sum(gatewayFailures)+sum(directFailures)),
		metric("overhead_p99_ms", "ms", "lower_is_better", overheadP99, sum(gatewayFailures)+sum(directFailures)),
		metric("direct_throughput_rps", "requests/second", "higher_is_better", directRPS, sum(directFailures)),
		metric("gateway_throughput_rps", "requests/second", "higher_is_better", gatewayRPS, sum(gatewayFailures)),
		metric("gateway_rejects", "requests", "lower_is_better", gatewayRejects, 0),
		metric("gateway_failures", "requests", "lower_is_better", gatewayFailures, sum(gatewayFailures)),
	}
}

func extract(runs []endpointStats) (p50, p95, p99, throughput []float64) {
	for _, run := range runs {
		p50 = append(p50, run.p50)
		p95 = append(p95, run.p95)
		p99 = append(p99, run.p99)
		throughput = append(throughput, run.throughput)
	}
	return p50, p95, p99, throughput
}

func subtract(left, right []float64) []float64 {
	values := make([]float64, len(left))
	for index := range left {
		values[index] = left[index] - right[index]
	}
	return values
}

func counts(runs []endpointStats, selectValue func(endpointStats) int) []float64 {
	values := make([]float64, len(runs))
	for index, run := range runs {
		values[index] = float64(selectValue(run))
	}
	return values
}

func sum(values []float64) int {
	total := 0
	for _, value := range values {
		total += int(value)
	}
	return total
}

func percentile(sorted []float64, quantile float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := int(float64(len(sorted)-1)*quantile + 0.5)
	return sorted[index]
}

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func newUUID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])
}
