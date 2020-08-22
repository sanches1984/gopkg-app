package grpc

import (
	grpcmw "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	mw "github.com/severgroup-tt/gopkg-app/client/grpc/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// NewClient ...
func NewClient(opt Option, optFn ...OptionFn) (*grpc.ClientConn, error) {
	opt.dialOption = make([]grpc.DialOption, 0)
	opt.unaryInterceptor = make([]grpc.UnaryClientInterceptor, 0)
	opt.streamInterceptor = make([]grpc.StreamClientInterceptor, 0)
	for _, o := range optFn {
		o(&opt)
	}
	return grpc.Dial(
		opt.Addr,
		append(opt.dialOption,
			grpc.WithInsecure(),
			grpc.WithUserAgent(opt.AppName),
			grpc.WithUnaryInterceptor(grpcmw.ChainUnaryClient(
				append(opt.unaryInterceptor, defaultUnaryInterceptors(opt)...)...)),
			grpc.WithStreamInterceptor(grpcmw.ChainStreamClient(
				append(opt.streamInterceptor, defaultStreamInterceptors(opt)...)...)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				PermitWithoutStream: true, // most likely it does not work without proper setup on server side
			}),
		)...)
}

// DefaultUnaryInterceptors ...
func defaultUnaryInterceptors(opts Option) []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		mw.NewAppInfoUnaryInterceptor(opts.AppName, opts.AppVersion),
		grpc_opentracing.UnaryClientInterceptor(),
		grpc_prometheus.UnaryClientInterceptor,
		mw.NewLogUnaryInterceptor(opts.Service),
		mw.NewRetryUnaryInterceptor(opts.Service, opts.MaxRetry, opts.RetryDelay),
	}
}

// DefaultStreamInterceptors ...
func defaultStreamInterceptors(opts Option) []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		grpc_opentracing.StreamClientInterceptor(),
		grpc_prometheus.StreamClientInterceptor,
		mw.NewAppInfoStreamInterceptor(opts.AppName, opts.AppVersion),
		mw.NewLogStreamInterceptor(opts.Service),
	}
}
