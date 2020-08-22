package middleware

import (
	"github.com/go-chi/chi/middleware"
	"net/http"
	"testing"
)

func TestSetId(t *testing.T) {
	t.Run("SetUuid", func(t *testing.T) {
		var r = &http.Request{}
		var x = r.WithContext(SetRequestId(r.Context()))
		y := GetRequestId(x.Context())
		if x.Context().Value(middleware.RequestIDKey) != y {
			t.Error("uuid Keys do not match")
		}
	})
}
