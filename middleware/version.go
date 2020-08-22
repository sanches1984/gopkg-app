package middleware

import (
	"net/http"
)

const versionHeaderName = "X-App-Version"

func NewVersionMiddleware(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(versionHeaderName, version)
			next.ServeHTTP(w, r)
		})
	}
}
