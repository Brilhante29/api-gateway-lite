package telemetry

import (
	"net/http"
)

func Middleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := Tracer().Start(r.Context(), "HTTP "+r.Method+" "+r.URL.Path)
			defer span.End()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
