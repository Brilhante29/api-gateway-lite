package internal

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

type principalContextKey struct{}

const APIKeyHeader = "X-API-Key"

func AuthMiddleware(apiKey string) func(http.Handler) http.Handler {
	expected := []byte(apiKey)
	sum := sha256.Sum256(expected)
	principal := hex.EncodeToString(sum[:16])

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := []byte(r.Header.Get(APIKeyHeader))
			if len(provided) != len(expected) || subtle.ConstantTimeCompare(provided, expected) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Never forward gateway credentials to the upstream service.
			r.Header.Del(APIKeyHeader)
			ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func PrincipalFromContext(ctx context.Context) (string, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(string)
	return principal, ok && principal != ""
}
