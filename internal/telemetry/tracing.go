package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func SetupTracerProvider(serviceName string) (trace.TracerProvider, func(), error) {
	var tp trace.TracerProvider

	if os.Getenv("TELEMETRY") == "stdout" {
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, nil, err
		}
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(resource.NewSchemaless(
				attribute.String("service.name", serviceName),
			)),
		)
	} else {
		tp = trace.NewNoopTracerProvider()
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	cleanup := func() {
		if sdkTp, ok := tp.(*sdktrace.TracerProvider); ok {
			_ = sdkTp.Shutdown(context.Background())
		}
	}

	return tp, cleanup, nil
}

func Tracer() trace.Tracer {
	return otel.Tracer("api-gateway-lite")
}
