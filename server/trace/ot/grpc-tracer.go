package ot

import (
	"context"
	"net"
	"path"
	"strconv"

	appIdentity "github.com/H-BF/corlib/app/identity"
	otPriv "github.com/H-BF/corlib/internal/pkg/ot"
	"github.com/H-BF/corlib/pkg/conventions"
	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/internal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

//GRPCTracerOption опции
type GRPCTracerOption interface {
	apply(*GRPCTracer)
}

//ServerTracerProvider ...
type ServerTracerProvider = trace.TracerProvider

//GRPCTracer трейсим серверные GRPC вызовы
type GRPCTracer struct {
	propagator     propagation.TextMapPropagator
	tracerProvider ServerTracerProvider
}

var (
	_ server.GRPCTracer = (*GRPCTracer)(nil)
)

//NewGRPCServerTracer ...
func NewGRPCServerTracer(options ...GRPCTracerOption) *GRPCTracer {
	ret := new(GRPCTracer)
	for _, o := range options {
		o.apply(ret)
	}
	ret.propagator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	return ret
}

var (
	_ = NewGRPCServerTracer
)

func (impl *GRPCTracer) methodInfo(ctx context.Context, fullMethodName string) *conventions.GrpcMethodInfo {
	var ret conventions.GrpcMethodInfo
	if !ret.FromContext(ctx) {
		e := ret.Init(fullMethodName)
		if e != nil {
			panic(e)
		}
	}
	return &ret
}

func (impl *GRPCTracer) attrsFromPeer(ctx context.Context) []attribute.KeyValue {
	var ret []attribute.KeyValue
	if p, _ := peer.FromContext(ctx); p != nil && p.Addr != nil {
		a := p.Addr
		switch a.Network() {
		case "tcp":
			host, port, e := net.SplitHostPort(a.String())
			if e != nil {
				return nil
			}
			ret = append(ret, semconv.NetTransportTCP, semconv.NetPeerIPKey.String(host))
			var n int
			if n, e = strconv.Atoi(port); e != nil {
				return nil
			}
			ret = append(ret, semconv.NetPeerPortKey.Int(n))
		case "unix":
			ret = append(ret, semconv.NetTransportUnix, otPriv.NetPeerUnixSocketKey.String(a.String()))
		}
	}
	return ret
}

func (impl *GRPCTracer) spanStart(ctx context.Context, fullMethodName string) (context.Context, trace.Span) {
	if md, _ := metadata.FromIncomingContext(ctx); md != nil {
		ctx1 := impl.propagator.Extract(ctx, otPriv.TextMapCarrierFromGrpcMD{MD: md})
		if spanCtx := trace.SpanContextFromContext(ctx1); spanCtx.IsValid() {
			ctx = trace.ContextWithRemoteSpanContext(ctx, spanCtx)
		}
	}
	tp := impl.tracerProvider
	if tp == nil {
		return ctx, nil
	}
	minfo := impl.methodInfo(ctx, fullMethodName)
	attrs := append(impl.attrsFromPeer(ctx),
		otPriv.RPCSystemGRPC,
		semconv.RPCServiceKey.String(minfo.ServiceFQN),
		semconv.RPCMethodKey.String(minfo.Method),
	)
	tracer := tp.Tracer(path.Join(appIdentity.Name, "grpc-server"))
	return tracer.Start(
		ctx,
		minfo.String()[1:], //without prefixed slash
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attrs...),
	)
}

func (impl *GRPCTracer) spanEnd(span trace.Span, err error) {
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

//TraceUnaryCalls unary interceptor
func (impl *GRPCTracer) TraceUnaryCalls(ctx context.Context, req interface{}, i *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var span trace.Span
	ctx, span = impl.spanStart(ctx, i.FullMethod)
	resp, err = handler(ctx, req)
	impl.spanEnd(span, err)
	return
}

//TraceStreamCalls stream interceptor
func (impl *GRPCTracer) TraceStreamCalls(srv interface{}, ss grpc.ServerStream, i *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx, span := impl.spanStart(ss.Context(), i.FullMethod)
	ss1 := internal.ServerStreamWithContext(ctx,
		&serverStreamWrapper{ServerStream: ss, span: span},
	)
	err := handler(srv, ss1)
	impl.spanEnd(span, err)
	return err
}
