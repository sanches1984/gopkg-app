package middleware

import (
	"github.com/go-chi/cors"
	"net/http"
)

func NewCorsMiddleware() func(next http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "X-Token", "X-Compress", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	return func(next http.Handler) http.Handler {
		return c.Handler(next)
	}
}
