package ot

import (
	"errors"
	"io"
	"sync"

	otPriv "github.com/H-BF/corlib/internal/pkg/ot"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type clientStreamWrapper struct {
	sync.RWMutex
	grpc.ClientStream
	*grpc.StreamDesc
	span   trace.Span
	msgIn  int
	msgOut int
}

var _ grpc.ClientStream = (*clientStreamWrapper)(nil)

// Header impl grpc.ClientStream
func (impl *clientStreamWrapper) Header() (metadata.MD, error) {
	md, err := impl.ClientStream.Header()
	if err != nil {
		impl.endSpan(err)
	}
	return md, err
}

// CloseSend impl grpc.ClientStream
func (impl *clientStreamWrapper) CloseSend() error {
	if err := impl.ClientStream.CloseSend(); err != nil {
		impl.endSpan(err)
		return err
	}
	return nil
}

// SendMsg impl grpc.ClientStream
func (impl *clientStreamWrapper) SendMsg(m interface{}) error {
	if err := impl.ClientStream.SendMsg(m); err != nil {
		impl.endSpan(err)
		return err
	}
	impl.onMessage(m, false)
	return nil
}

// RecvMsg impl grpc.ClientStream
func (impl *clientStreamWrapper) RecvMsg(m interface{}) error {
	err := impl.ClientStream.RecvMsg(m)
	if errors.Is(err, io.EOF) {
		impl.endSpan(nil)
		return err
	}
	if err != nil {
		impl.endSpan(err)
		return err
	}
	impl.onMessage(m, true)
	if !impl.ServerStreams {
		impl.endSpan(nil)
	}
	return nil
}

func (impl *clientStreamWrapper) endSpan(err error) {
	impl.Lock()
	span := impl.span
	impl.span = nil
	impl.Unlock()
	spanEndFromGRPC(span, err)
}

func (impl *clientStreamWrapper) onMessage(m interface{}, isReceived bool) {
	impl.RLock()
	defer impl.RUnlock()
	span := impl.span
	if span == nil {
		return
	}
	var attrs []attribute.KeyValue
	if isReceived {
		attrs = []attribute.KeyValue{
			otPriv.RPCMessageIDKey.Int(impl.msgIn),
			otPriv.RPCMessageReceived,
		}
	} else {
		attrs = []attribute.KeyValue{
			otPriv.RPCMessageIDKey.Int(impl.msgOut),
			otPriv.RPCMessageSent,
		}
	}
	if protoMsg, ok := m.(proto.Message); ok {
		attrs = append(attrs, otPriv.RPCMessageSizeKey.Int(proto.Size(protoMsg)))
	}
	span.AddEvent("message", trace.WithAttributes(attrs...))
	if isReceived {
		impl.msgIn++
	} else {
		impl.msgOut++
	}
}

func spanEndFromGRPC(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		st, _ := status.FromError(err)
		if st != nil {
			span.SetStatus(codes.Error, st.Message())
			span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int(int(st.Code())))
		}
	} else {
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(semconv.RPCGRPCStatusCodeOk)
	}
	span.End()
}
