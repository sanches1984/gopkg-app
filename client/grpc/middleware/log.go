package middleware

import (
	"context"
	logger "github.com/severgroup-tt/gopkg-logger"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Level ...
type Level int

// Log levels
const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
)

// Code ...
func (l Level) Code() int {
	return int(l)
}

// String ...
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "DEBUG"
	}
}

// NewLogUnaryInterceptor ...
func NewLogUnaryInterceptor(service string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startedAt := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)

		logLevel := INFO
		if err != nil {
			logLevel = ERROR
		}

		message := "GRPC service: %v, Host: %v, Url: %v, Request: %v, Response: %+v, ResponseError: %v, Error: %v, TimeElapsed: %v"
		values := make([]interface{}, 0)
		values = append(values, service, cc.Target(), method, req, reply, status.Code(err).String(), err, time.Since(startedAt))
		switch logLevel {
		case INFO:
			logger.Info(ctx, message, values...)
		case ERROR:
			logger.Error(ctx, message, values...)
		case DEBUG:
			logger.Debug(ctx, message, values...)
		default:
			logger.Info(ctx, message, values...)
		}

		return err
	}
}

// NewLogStreamInterceptor ...
func NewLogStreamInterceptor(service string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		startedAt := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		logLevel := INFO
		if err != nil {
			logLevel = ERROR
		}

		message := "GRPC_STREAM service: %v, Host: %v, Url: %v, ResponseStatus: %v, ResponseError: %v, Err: %v, TimeElapsed: %v"
		values := make([]interface{}, 0)
		values = append(values, service, cc.Target(), method, status.Code(err).String(), err, time.Since(startedAt))
		switch logLevel {
		case INFO:
			logger.Info(ctx, message, values...)
		case ERROR:
			logger.Error(ctx, message, values...)
		case DEBUG:
			logger.Debug(ctx, message, values...)
		default:
			logger.Info(ctx, message, values...)
		}

		return clientStream, err
	}
}
