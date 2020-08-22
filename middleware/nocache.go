package middleware

import (
	"github.com/go-chi/chi/middleware"
	"net/http"
)

func NewNoCacheMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return middleware.NoCache(next)
	}
}
