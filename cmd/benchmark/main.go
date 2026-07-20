package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Brilhante29/api-gateway-lite/internal"
	"github.com/Brilhante29/api-gateway-lite/internal/benchmark"
	"github.com/Brilhante29/api-gateway-lite/internal/ratelimit"
	"github.com/Brilhante29/api-gateway-lite/internal/telemetry"
)

func main() {
	targetMux := http.NewServeMux()
	targetMux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	targetMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	targetServer := &http.Server{Addr: ":9091", Handler: targetMux}
	go func() {
		if err := targetServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("target server exited: %v", err)
		}
	}()

	cfg := internal.Config{
		UpstreamURL: "http://localhost:9091",
		ListenAddr:  ":9090",
		APIKey:      "bench-key",
		RateLimit:   10000,
		RateBurst:   20000,
		Telemetry:   "none",
	}

	_, cleanup, err := telemetry.SetupTracerProvider("api-gateway-lite")
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer cleanup()

	proxy, err := internal.NewProxy(cfg.UpstreamURL)
	if err != nil {
		log.Fatalf("proxy: %v", err)
	}

	limiter := ratelimit.NewLimiter(cfg.RateLimit, cfg.RateBurst)

	mux := http.NewServeMux()
	mux.Handle("/", proxy)

	handler := telemetry.Middleware("api-gateway-lite")(mux)
	handler = ratelimit.Middleware(limiter)(handler)
	handler = internal.AuthMiddleware(cfg.APIKey)(handler)

	gatewayServer := &http.Server{Addr: cfg.ListenAddr, Handler: handler}
	go func() {
		if err := gatewayServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("gateway exited: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	result, err := benchmark.Run("http://localhost:9090", "http://localhost:9091", 100)
	if err != nil {
		log.Fatalf("benchmark: %v", err)
	}

	result.Timestamp = time.Now().UTC().Format(time.RFC3339)
	result.Command = "go run ./cmd/benchmark"

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))

	if err := os.MkdirAll("benchmarks/results", 0755); err == nil {
		os.WriteFile("benchmarks/results/latest.json", data, 0644)
	}

	targetServer.Close()
	gatewayServer.Close()
}
