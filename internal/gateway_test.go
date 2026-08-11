package internal

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Brilhante29/api-gateway-lite/internal/ratelimit"
	"github.com/Brilhante29/api-gateway-lite/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type gatewayLimiterStub struct {
	decision ratelimit.Decision
}

func (stub gatewayLimiterStub) Allow(context.Context, string) (ratelimit.Decision, error) {
	return stub.decision, nil
}

type healthStub struct {
	err error
}

func (stub healthStub) Ping(context.Context) error { return stub.err }

func TestGatewayPropagatesRouteCorrelationAndTrace(t *testing.T) {
	provider := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})

	observed := make(chan http.Header, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed <- r.Header.Clone()
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(r.URL.RequestURI()))
	}))
	defer upstream.Close()

	proxy, err := NewProxy(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{APIKey: "test-key", ServiceName: "gateway-test"}
	handler := NewGatewayHandler(cfg, proxy, gatewayLimiterStub{ratelimit.Decision{
		Allowed: true, Limit: 10, Remaining: 9, ResetAfter: time.Second,
	}}, healthStub{})

	request := httptest.NewRequest(http.MethodGet, "/orders/42?expand=items", nil)
	request.Header.Set(APIKeyHeader, "test-key")
	request.Header.Set(telemetry.CorrelationIDHeader, "correlation-42")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Body.String() != "/orders/42?expand=items" {
		t.Fatalf("body = %q", recorder.Body.String())
	}
	if got := recorder.Header().Get(telemetry.CorrelationIDHeader); got != "correlation-42" {
		t.Fatalf("response correlation ID = %q", got)
	}

	headers := <-observed
	if headers.Get(APIKeyHeader) != "" {
		t.Error("gateway credential reached upstream")
	}
	if headers.Get(telemetry.CorrelationIDHeader) != "correlation-42" {
		t.Errorf("upstream correlation ID = %q", headers.Get(telemetry.CorrelationIDHeader))
	}
	traceparent := headers.Get("traceparent")
	if !strings.HasPrefix(traceparent, "00-") || len(traceparent) != 55 {
		t.Errorf("upstream traceparent = %q", traceparent)
	}
}

func TestGatewayHealthReflectsRedisAvailability(t *testing.T) {
	config := Config{APIKey: "test-key", ServiceName: "gateway-test"}
	limiter := gatewayLimiterStub{decision: ratelimit.Decision{Allowed: true}}

	for _, test := range []struct {
		name   string
		health healthStub
		want   int
	}{
		{name: "ready", health: healthStub{}, want: http.StatusOK},
		{name: "Redis unavailable", health: healthStub{err: errors.New("down")}, want: http.StatusServiceUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := NewGatewayHandler(config, http.NotFoundHandler(), limiter, test.health)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
			if recorder.Header().Get(telemetry.CorrelationIDHeader) == "" {
				t.Error("health response is missing a generated correlation ID")
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	valid := Config{
		APIKey: "key", RedisAddr: "redis:6379", RateLimit: 1, RateBurst: 1,
		RateLimitPrefix: "gateway", Telemetry: "otlp",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	valid.RateBurst = 0
	if err := valid.Validate(); err == nil {
		t.Fatal("invalid rate burst accepted")
	}
}
