package middleware

import (
	"github.com/go-chi/chi/middleware"
	"net/http"
)

func NewHeartbeatMiddleware() func(next http.Handler) http.Handler {
	return middleware.Heartbeat("/ping")
}
