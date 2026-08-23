package auth

import (
	"crypto/subtle"
	"net/http"
)

const InternalTokenHeader = "X-CloudScale-Internal-Token"

func InternalMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			http.Error(
				w,
				"internal authentication is not configured",
				http.StatusInternalServerError,
			)
			return
		}

		provided := r.Header.Get(InternalTokenHeader)

		providedBytes := []byte(provided)
		tokenBytes := []byte(token)

		valid := len(providedBytes) == len(tokenBytes)

		if valid {
			valid = subtle.ConstantTimeCompare(
				providedBytes,
				tokenBytes,
			) == 1
		}

		if !valid {
			http.Error(
				w,
				"forbidden",
				http.StatusForbidden,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
