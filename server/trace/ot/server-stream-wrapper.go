package ot

import (
	"sync/atomic"

	otPriv "github.com/H-BF/corlib/internal/pkg/ot"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type serverStreamWrapper struct {
	grpc.ServerStream
	span   trace.Span
	msgIn  int32
	msgOut int32
}

var (
	_ grpc.ServerStream = (*serverStreamWrapper)(nil)
)

// SendMsg impl grpc.ServerStream
func (impl *serverStreamWrapper) SendMsg(m interface{}) (err error) {
	if err = impl.ServerStream.SendMsg(m); err == nil {
		impl.onMessage(m, false)
	}
	return
}

// RecvMsg impl grpc.ServerStream
func (impl *serverStreamWrapper) RecvMsg(m interface{}) (err error) {
	if err = impl.ServerStream.RecvMsg(m); err == nil {
		impl.onMessage(m, true)
	}
	return
}

func (impl *serverStreamWrapper) onMessage(m interface{}, isReceived bool) {
	if impl.span == nil {
		return
	}
	var attrs []attribute.KeyValue
	if isReceived {
		i := int(atomic.AddInt32(&impl.msgIn, 1))
		attrs = []attribute.KeyValue{
			otPriv.RPCMessageIDKey.Int(i - 1),
			otPriv.RPCMessageReceived,
		}
	} else {
		i := int(atomic.AddInt32(&impl.msgOut, 1))
		attrs = []attribute.KeyValue{
			otPriv.RPCMessageIDKey.Int(i - 1),
			otPriv.RPCMessageSent,
		}
	}
	if protoMsg, ok := m.(proto.Message); ok {
		attrs = append(attrs, otPriv.RPCMessageSizeKey.Int(proto.Size(protoMsg)))
	}
	impl.span.AddEvent("message", trace.WithAttributes(attrs...))
}
