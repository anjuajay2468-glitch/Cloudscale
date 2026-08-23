package auth

import "net/http"

const APIKeyHeader = "X-API-Key"

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Enabled() {
			http.Error(
				w,
				"authentication is not configured",
				http.StatusInternalServerError,
			)
			return
		}

		providedKey := r.Header.Get(APIKeyHeader)

		role, ok := a.Authenticate(providedKey)

		if !ok {
			w.Header().Set("WWW-Authenticate", "API-Key")
			http.Error(
				w,
				"unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		if !a.Authorize(role, r.Method) {
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
