package telemetry

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Middleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		withCorrelationAttribute := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trace.SpanFromContext(r.Context()).SetAttributes(
				attribute.String("correlation.id", r.Header.Get(CorrelationIDHeader)),
			)
			next.ServeHTTP(w, r)
		})
		return otelhttp.NewHandler(
			withCorrelationAttribute,
			serviceName,
			otelhttp.WithSpanNameFormatter(func(operation string, request *http.Request) string {
				return request.Method + " " + request.URL.Path
			}),
		)
	}
}
