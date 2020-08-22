package transport

import (
	"github.com/severgroup-tt/gopkg-errors/transport"

	"github.com/utrack/clay/v2/transport/httpruntime"
)

// Override
func Override(options *Options) {
	OverrideMarshaller(options)
	OverrideErrorRenderer()
}

// OverrideMarshaller
func OverrideMarshaller(options *Options) {
	jsonMarshaller := NewMarshaller(options)
	httpruntime.OverrideMarshaler(jsonMarshaller.ContentType(), jsonMarshaller)
}

// OverrideErrorRenderer
func OverrideErrorRenderer() {
	httpruntime.SetError = transport.ErrorRenderer
	httpruntime.TransformUnmarshalerError = TransformUnmarshalerError
}

func NewOptions() *Options {
	return &Options{
		emitDefaults: true,
	}
}

func (m *Options) WithMarshallerEmitDefaults(val bool) *Options {
	m.emitDefaults = val
	return m
}

type Options struct {
	emitDefaults bool
}
