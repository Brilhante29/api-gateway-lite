package ratelimit

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type SubjectFunc func(*http.Request) (string, bool)

func Middleware(limiter Limiter, subject SubjectFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, ok := subject(r)
			if !ok {
				http.Error(w, "rate-limit identity unavailable", http.StatusInternalServerError)
				return
			}

			decision, err := limiter.Allow(r.Context(), key)
			if err != nil {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "rate limiter unavailable", http.StatusServiceUnavailable)
				return
			}

			resetSeconds := secondsCeil(decision.ResetAfter)
			w.Header().Set("RateLimit-Limit", strconv.Itoa(decision.Limit))
			w.Header().Set("RateLimit-Remaining", strconv.Itoa(decision.Remaining))
			w.Header().Set("RateLimit-Reset", strconv.FormatInt(resetSeconds, 10))
			// Compatibility aliases used by project #12 and common API clients.
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(decision.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(decision.ResetAfter).Unix(), 10))

			if !decision.Allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(secondsCeil(decision.RetryAfter), 10))
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ContractSummary() string {
	return fmt.Sprintf("token-bucket; headers=%s,%s,%s", "RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset")
}
