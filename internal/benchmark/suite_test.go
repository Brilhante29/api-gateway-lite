package benchmark

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMeasure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	avg, err := measure(ts.URL, 5)
	if err != nil {
		t.Fatal(err)
	}
	if avg < 0 {
		t.Errorf("expected non-negative average, got %f", avg)
	}
}

func TestRun(t *testing.T) {
	direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer direct.Close()

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer gateway.Close()

	result, err := Run(gateway.URL, direct.URL, 3)
	if err != nil {
		t.Fatal(err)
	}

	if result.Project != "api-gateway-lite" {
		t.Errorf("expected project api-gateway-lite, got %s", result.Project)
	}
	if result.Metric != "overhead_ms" {
		t.Errorf("expected overhead_ms, got %s", result.Metric)
	}
	if result.Unit != "ms" {
		t.Errorf("expected ms, got %s", result.Unit)
	}
}
