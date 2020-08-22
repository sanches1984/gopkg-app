package middleware

import (
	"context"
	"net/http"

	servertiming "github.com/mitchellh/go-server-timing"
)

type timingResponseWriter struct {
	headerWritten bool
	startMetric   *servertiming.Metric
	http.ResponseWriter
}

func (v *timingResponseWriter) Write(bytes []byte) (int, error) {
	if !v.headerWritten {
		v.startMetric.Stop()
		// github.com/mitchellh/go-server-timing пишет заголовки при отдаче кода ответа
		// если начало писать ответ - считаем что запрос завершится успешно
		// Костыль, если произошла ошибка второй раз не отправляем заголовок
		if len(bytes) > 8 && string(bytes[:8]) != `{"error"` {
			v.ResponseWriter.WriteHeader(200)
		}
		v.headerWritten = true
	}
	return v.ResponseWriter.Write(bytes)
}

func NewTimingMiddleware() []func(next http.Handler) http.Handler {
	init := func(next http.Handler) http.Handler {
		return servertiming.Middleware(next, nil)
	}
	all := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := NewTimingMetric(r.Context(), "all")
			next.ServeHTTP(&timingResponseWriter{ResponseWriter: w, startMetric: m}, r)
		})
	}

	return []func(next http.Handler) http.Handler{init, all}
}

func NewTimingMetric(ctx context.Context, name string) *servertiming.Metric {
	return servertiming.FromContext(ctx).NewMetric(name).Start()
}
