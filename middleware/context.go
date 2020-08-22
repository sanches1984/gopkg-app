package middleware

import (
	"context"
)

var logExtraKey = new(struct{})

// LogExtraToContext set key/value log extra info to context
func LogExtraToContext(ctx context.Context, kv ...interface{}) context.Context {
	return context.WithValue(ctx, &logExtraKey, kv)
}

// logExtraFromContext get key/value log extra info from context
func logExtraFromContext(ctx context.Context) []interface{} {
	v := ctx.Value(&logExtraKey)
	if v == nil {
		return nil
	}
	return v.([]interface{})
}
