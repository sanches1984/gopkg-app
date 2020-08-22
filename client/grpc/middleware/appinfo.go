package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	AppNameHeader    = "x-app-name"
	AppVersionHeader = "x-app-version"
)

func NewAppInfoUnaryInterceptor(sourceAppName string, sourceAppVersion string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, AppNameHeader, sourceAppName)
		ctx = metadata.AppendToOutgoingContext(ctx, AppVersionHeader, sourceAppVersion)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func NewAppInfoStreamInterceptor(sourceAppName string, sourceAppVersion string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, AppNameHeader, sourceAppName)
		ctx = metadata.AppendToOutgoingContext(ctx, AppVersionHeader, sourceAppVersion)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
