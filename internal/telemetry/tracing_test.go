package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
)

func TestOTLPModeExportsTracesOverHTTP(t *testing.T) {
	var requests atomic.Int32
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/traces" {
			t.Errorf("export path = %q, want /v1/traces", r.URL.Path)
		}
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer collector.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", collector.URL)
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
	provider, shutdown, err := SetupTracerProvider(context.Background(), "telemetry-test", "otlp")
	if err != nil {
		t.Fatal(err)
	}

	_, span := provider.Tracer("test").Start(context.Background(), "export-me")
	span.End()
	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := provider.ForceFlush(flushCtx); err != nil {
		t.Fatalf("force flush: %v", err)
	}
	if err := shutdown(flushCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if requests.Load() == 0 {
		t.Fatal("OTLP collector received no trace export")
	}
}
