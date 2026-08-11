package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyForwardsRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test" || r.URL.Query().Get("page") != "2" {
			t.Errorf("route = %s?%s, want /test?page=2", r.URL.Path, r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("upstream ok"))
	}))
	defer upstream.Close()

	proxy, err := NewProxy(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/test?page=2", nil)
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
