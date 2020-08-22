package grpc

import (
	"google.golang.org/grpc"
	"time"
)

type OptionFn func(a *Option)

type Option struct {
	Service           string
	AppName           string
	AppVersion        string
	Addr              string
	MaxRetry          uint
	Timeout           time.Duration
	RetryDelay        time.Duration
	dialOption        []grpc.DialOption
	unaryInterceptor  []grpc.UnaryClientInterceptor
	streamInterceptor []grpc.StreamClientInterceptor
}

func WithDialOption(fn grpc.DialOption) OptionFn {
	return func(o *Option) {
		o.dialOption = append(o.dialOption, fn)
	}
}

func WithUnaryInterceptor(fn grpc.UnaryClientInterceptor) OptionFn {
	return func(o *Option) {
		o.unaryInterceptor = append(o.unaryInterceptor, fn)
	}
}

func WithStreamInterceptor(fn grpc.StreamClientInterceptor) OptionFn {
	return func(o *Option) {
		o.streamInterceptor = append(o.streamInterceptor, fn)
	}
}
