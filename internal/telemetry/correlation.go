package telemetry

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const CorrelationIDHeader = "X-Correlation-ID"

func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(CorrelationIDHeader)
		if !validCorrelationID(correlationID) {
			correlationID = newCorrelationID()
		}

		r.Header.Set(CorrelationIDHeader, correlationID)
		w.Header().Set(CorrelationIDHeader, correlationID)
		next.ServeHTTP(w, r)
	})
}

func validCorrelationID(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func newCorrelationID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(value)
}
