package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/Brilhante29/api-gateway-lite/internal/ratelimit"
	"github.com/Brilhante29/api-gateway-lite/internal/telemetry"
)

type HealthChecker interface {
	Ping(context.Context) error
}

func NewGatewayHandler(cfg Config, proxy http.Handler, limiter ratelimit.Limiter, health HealthChecker) http.Handler {
	protected := ratelimit.Middleware(limiter, func(request *http.Request) (string, bool) {
		return PrincipalFromContext(request.Context())
	})(proxy)
	protected = AuthMiddleware(cfg.APIKey)(protected)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := health.Ping(ctx); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/", protected)

	return telemetry.CorrelationMiddleware(telemetry.Middleware(cfg.ServiceName)(mux))
}
