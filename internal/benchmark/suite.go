package benchmark

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Result struct {
	Project      string  `json:"project"`
	Metric       string  `json:"metric"`
	Value        float64 `json:"value"`
	Unit         string  `json:"unit"`
	DirectAvgMs  float64 `json:"direct_avg_ms"`
	GatewayAvgMs float64 `json:"gateway_avg_ms"`
	Timestamp    string  `json:"timestamp,omitempty"`
	Command      string  `json:"command,omitempty"`
}

func Run(gatewayURL, directURL string, requests int) (*Result, error) {
	directAvg, err := measure(directURL+"/echo", requests)
	if err != nil {
		return nil, fmt.Errorf("direct measure: %w", err)
	}

	gatewayAvg, err := measure(gatewayURL+"/echo", requests)
	if err != nil {
		return nil, fmt.Errorf("gateway measure: %w", err)
	}

	overhead := gatewayAvg - directAvg
	if overhead < 0 {
		overhead = 0
	}

	return &Result{
		Project:      "api-gateway-lite",
		Metric:       "overhead_ms",
		Value:        overhead,
		Unit:         "ms",
		DirectAvgMs:  directAvg,
		GatewayAvgMs: gatewayAvg,
	}, nil
}

func measure(url string, n int) (float64, error) {
	var total time.Duration
	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < n; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, err
		}
		req.Header.Set("X-API-Key", "bench-key")

		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			return 0, err
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		total += time.Since(start)
	}

	return float64(total.Milliseconds()) / float64(n), nil
}
