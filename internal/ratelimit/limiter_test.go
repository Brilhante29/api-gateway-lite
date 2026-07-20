package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterAllows(t *testing.T) {
	l := NewLimiter(100, 200)
	if !l.Allow() {
		t.Error("expected Allow to return true")
	}
}

func TestLimiterBlocked(t *testing.T) {
	l := NewLimiter(0.001, 1)
	if !l.Allow() {
		t.Error("expected first request to be allowed")
	}
	if l.Allow() {
		t.Error("expected second request to be blocked")
	}
}

func TestLimiterRefills(t *testing.T) {
	l := NewLimiter(100, 200)
	for i := 0; i < 200; i++ {
		l.Allow()
	}
	if l.Allow() {
		t.Error("expected bucket to be empty after exhausting burst")
	}
	time.Sleep(20 * time.Millisecond)
	if !l.Allow() {
		t.Error("expected bucket to have refilled after waiting")
	}
}

func TestLimiterBurst(t *testing.T) {
	l := NewLimiter(1000, 5)
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Errorf("expected request %d to be allowed within burst", i+1)
		}
	}
	if l.Allow() {
		t.Error("expected request beyond burst to be blocked")
	}
}
