package benchmark

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunProducesComparableV2Evidence(t *testing.T) {
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer direct.Close()
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer gateway.Close()

	digest := "sha256:" + strings.Repeat("a", 64)
	result, err := Run(context.Background(), Config{
		GatewayURL: gateway.URL, DirectURL: direct.URL, APIKey: "key",
		WarmupIterations: 3, MeasuredIterations: 12, Concurrency: 2, Repeat: 3,
		Command: "test", SourceCommit: strings.Repeat("b", 40), CleanTree: true,
		ImageRef: "api-gateway-lite:test", ImageDigest: digest,
		DependencyLockDigest: digest, Producer: "local", HardwareClass: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.SchemaVersion != 2 || result.Execution.Repeat != 3 {
		t.Fatalf("unexpected V2 identity: schema=%d repeat=%d", result.SchemaVersion, result.Execution.Repeat)
	}
	if len(result.Metrics) != 14 {
		t.Fatalf("metrics = %d, want 14", len(result.Metrics))
	}
	for _, metric := range result.Metrics {
		if len(metric.Samples) != 3 {
			t.Errorf("metric %s samples = %d, want 3", metric.Name, len(metric.Samples))
		}
	}
	metrics := make(map[string]Metric, len(result.Metrics))
	for _, metric := range result.Metrics {
		metrics[metric.Name] = metric
	}
	if metrics["overhead_p50_ms"].Value <= 0 {
		t.Errorf("primary metric = %+v", metrics["overhead_p50_ms"])
	}
	for _, name := range []string{"direct_p95_ms", "gateway_p95_ms", "overhead_p95_ms", "direct_throughput_rps", "gateway_throughput_rps", "gateway_rejects", "direct_failures", "gateway_failures"} {
		if _, ok := metrics[name]; !ok {
			t.Errorf("missing metric %q", name)
		}
	}
	if len(result.Provenance.ArtifactDigest) != 71 {
		t.Errorf("artifact digest = %q", result.Provenance.ArtifactDigest)
	}
	if !strings.Contains(result.ComparabilityKey, "12x2:r3") {
		t.Errorf("comparability key = %q", result.ComparabilityKey)
	}
}

func TestRunRejectsWeakProvenance(t *testing.T) {
	_, err := Run(context.Background(), Config{Repeat: 2})
	if err == nil {
		t.Fatal("weak benchmark configuration accepted")
	}
}
