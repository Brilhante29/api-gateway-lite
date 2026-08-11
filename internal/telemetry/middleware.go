package telemetry

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Middleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(
			next,
			serviceName,
			otelhttp.WithSpanNameFormatter(func(operation string, request *http.Request) string {
				return request.Method + " " + request.URL.Path
			}),
		)
	}
}
