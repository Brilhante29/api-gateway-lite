package ratelimit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

type limiterStub struct {
	decision Decision
	err      error
	subject  string
}

func (stub *limiterStub) Allow(_ context.Context, subject string) (Decision, error) {
	stub.subject = subject
	return stub.decision, stub.err
}

func TestMiddlewareAllowsAndPublishesQuotaHeaders(t *testing.T) {
	limiter := &limiterStub{decision: Decision{
		Allowed: true, Limit: 10, Remaining: 9, ResetAfter: 2 * time.Second,
	}}
	handler := Middleware(limiter, func(*http.Request) (string, bool) {
		return "principal-1", true
	})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if limiter.subject != "principal-1" {
		t.Fatalf("subject = %q, want principal-1", limiter.subject)
	}
	for name, want := range map[string]string{
		"RateLimit-Limit":       "10",
		"RateLimit-Remaining":   "9",
		"X-RateLimit-Limit":     "10",
		"X-RateLimit-Remaining": "9",
	} {
		if got := recorder.Header().Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if _, err := strconv.ParseInt(recorder.Header().Get("X-RateLimit-Reset"), 10, 64); err != nil {
		t.Errorf("X-RateLimit-Reset is not an epoch timestamp: %v", err)
	}
}

func TestMiddlewareRejectsAndFailsClosed(t *testing.T) {
	tests := []struct {
		name     string
		limiter  *limiterStub
		wantCode int
	}{
		{
			name: "quota exhausted",
			limiter: &limiterStub{decision: Decision{
				Allowed: false, Limit: 1, Remaining: 0, ResetAfter: time.Second, RetryAfter: time.Second,
			}},
			wantCode: http.StatusTooManyRequests,
		},
		{
			name:     "Redis unavailable",
			limiter:  &limiterStub{err: errors.New("connection refused")},
			wantCode: http.StatusServiceUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := Middleware(test.limiter, func(*http.Request) (string, bool) {
				return "principal-1", true
			})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("downstream handler must not run")
			}))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			if recorder.Code != test.wantCode {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantCode)
			}
		})
	}
}

func TestRedisLimiterSharesQuotaAcrossInstances(t *testing.T) {
	address := os.Getenv("TEST_REDIS_ADDR")
	if address == "" {
		t.Skip("TEST_REDIS_ADDR is required for the Redis integration contract")
	}

	clientA := NewRedisClient(address, "", 0)
	clientB := NewRedisClient(address, "", 0)
	t.Cleanup(func() { _ = clientA.Close() })
	t.Cleanup(func() { _ = clientB.Close() })

	prefix := "api-gateway-lite:test:" + strconv.FormatInt(time.Now().UnixNano(), 10)
	limiterA := NewRedisLimiter(clientA, 0.01, 2, prefix)
	limiterB := NewRedisLimiter(clientB, 0.01, 2, prefix)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	first, err := limiterA.Allow(ctx, "shared-principal")
	if err != nil || !first.Allowed {
		t.Fatalf("first decision = %+v, err = %v", first, err)
	}
	second, err := limiterB.Allow(ctx, "shared-principal")
	if err != nil || !second.Allowed {
		t.Fatalf("second decision = %+v, err = %v", second, err)
	}
	third, err := limiterA.Allow(ctx, "shared-principal")
	if err != nil {
		t.Fatal(err)
	}
	if third.Allowed || third.Remaining != 0 {
		t.Fatalf("shared quota was not enforced: %+v", third)
	}
}
