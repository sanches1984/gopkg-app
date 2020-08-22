package middleware

import (
	"context"
	"time"

	logger "github.com/severgroup-tt/gopkg-logger"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
)

const (
	DefaultMaxRetry   = uint(5)
	DefaultRetryDelay = 500 * time.Millisecond
)

func NewRetryUnaryInterceptor(service string, attempts uint, delay time.Duration) grpc.UnaryClientInterceptor {
	if delay < time.Millisecond {
		delay = DefaultRetryDelay
	}
	if attempts < 1 {
		attempts = DefaultMaxRetry
	}
	retryInterceptor := grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(attempts))

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		retryOpts := append(opts, grpc_retry.WithBackoff(func(attempt uint) time.Duration {
			logger.Error(ctx, "Failed to fetch GRPC data. "+
				"Service: %s, Host: %s, Url: %s, Request: %v, RetryAttempt: %d ",
				service, cc.Target(), method, req, attempt)

			return delay
		}))
		return retryInterceptor(ctx, method, req, reply, cc, invoker, retryOpts...)
	}
}
