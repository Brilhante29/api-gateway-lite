package main

import (
	"log"
	"net/http"

	"github.com/Brilhante29/api-gateway-lite/internal"
	"github.com/Brilhante29/api-gateway-lite/internal/ratelimit"
	"github.com/Brilhante29/api-gateway-lite/internal/telemetry"
)

func main() {
	cfg := internal.LoadConfig()

	_, cleanup, err := telemetry.SetupTracerProvider("api-gateway-lite")
	if err != nil {
		log.Fatalf("failed to setup telemetry: %v", err)
	}
	defer cleanup()

	proxy, err := internal.NewProxy(cfg.UpstreamURL)
	if err != nil {
		log.Fatalf("failed to create proxy: %v", err)
	}

	limiter := ratelimit.NewLimiter(cfg.RateLimit, cfg.RateBurst)

	mux := http.NewServeMux()
	mux.Handle("/", proxy)

	handler := telemetry.Middleware("api-gateway-lite")(mux)
	handler = ratelimit.Middleware(limiter)(handler)
	handler = internal.AuthMiddleware(cfg.APIKey)(handler)

	log.Printf("gateway listening on %s, upstream: %s", cfg.ListenAddr, cfg.UpstreamURL)
	log.Fatal(http.ListenAndServe(cfg.ListenAddr, handler))
}
