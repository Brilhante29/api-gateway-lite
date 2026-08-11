package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Brilhante29/api-gateway-lite/internal"
	"github.com/Brilhante29/api-gateway-lite/internal/ratelimit"
	"github.com/Brilhante29/api-gateway-lite/internal/telemetry"
)

func main() {
	cfg := internal.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	_, shutdownTelemetry, err := telemetry.SetupTracerProvider(ctx, cfg.ServiceName, cfg.Telemetry)
	if err != nil {
		log.Fatalf("setup telemetry: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			log.Printf("shutdown telemetry: %v", err)
		}
	}()

	proxy, err := internal.NewProxy(cfg.UpstreamURL)
	if err != nil {
		log.Fatalf("create proxy: %v", err)
	}

	redisClient := ratelimit.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	limiter := ratelimit.NewRedisLimiter(redisClient, cfg.RateLimit, cfg.RateBurst, cfg.RateLimitPrefix)
	defer func() {
		if err := limiter.Close(); err != nil {
			log.Printf("close Redis: %v", err)
		}
	}()

	startupCtx, cancelStartup := context.WithTimeout(ctx, 5*time.Second)
	defer cancelStartup()
	if err := limiter.Ping(startupCtx); err != nil {
		log.Fatalf("connect to Redis: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           internal.NewGatewayHandler(cfg, proxy, limiter, limiter),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("gateway listening on %s, upstream: %s", cfg.ListenAddr, cfg.UpstreamURL)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown HTTP server: %v", err)
		}
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve HTTP: %v", err)
		}
	}
}
