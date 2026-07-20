package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyForwardsRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("upstream ok"))
	}))
	defer upstream.Close()

	proxy, err := NewProxy(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "upstream ok" {
		t.Errorf("expected 'upstream ok', got %q", rec.Body.String())
	}
}

func TestProxyBadGateway(t *testing.T) {
	proxy, err := NewProxy("http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}
