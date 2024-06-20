package ot

import (
	"context"
	"net"
	"path"
	"runtime"
	"strconv"

	appIdentity "github.com/H-BF/corlib/app/identity"
	otPriv "github.com/H-BF/corlib/internal/pkg/ot"
	"github.com/H-BF/corlib/pkg/conventions"
	netPkg "github.com/H-BF/corlib/pkg/net"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientTracerProvider ...
type ClientTracerProvider = trace.TracerProvider

// NewClientGRPCTracer makes instance of *GRPCTracer
func NewClientGRPCTracer(tracerProvider ClientTracerProvider) *GRPCTracer {
	return &GRPCTracer{
		tracerProvider: tracerProvider,
		propagator:     propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	}
}

var (
	_ = NewClientGRPCTracer
)

// GRPCTracer OpenTelemetry tracer for GRPC client
type GRPCTracer struct {
	propagator     propagation.TextMapPropagator
	tracerProvider ClientTracerProvider
}

func (impl *GRPCTracer) methodInfo(fullMethodName string) *conventions.GrpcMethodInfo {
	var ret conventions.GrpcMethodInfo
	if e := ret.Init(fullMethodName); e != nil {
		panic(e)
	}
	return &ret
}

func (impl *GRPCTracer) attrsFromTarget(target string) []attribute.KeyValue {
	ep, err := netPkg.ParseEndpoint(target)
	if err != nil {
		return nil
	}
	var ret []attribute.KeyValue
	switch ep.Network() {
	case "tcp":
		host, port, e := net.SplitHostPort(ep.String())
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
		ret = append(ret, semconv.NetTransportUnix, otPriv.NetPeerUnixSocketKey.String(ep.String()))
	}
	return ret
}

func (impl *GRPCTracer) spanStart(ctx context.Context, fullMethodName string, targetAddr string) (trace.Span, context.Context) {
	tp := impl.tracerProvider
	if tp == nil {
		return nil, ctx
	}
	minfo := impl.methodInfo(fullMethodName)
	attrs := append(impl.attrsFromTarget(targetAddr),
		otPriv.RPCSystemGRPC,
		semconv.RPCServiceKey.String(minfo.ServiceFQN),
		semconv.RPCMethodKey.String(minfo.Method),
	)
	tracer := tp.Tracer(path.Join(appIdentity.Name, "grpc-client"))
	ctx1, span := tracer.Start(
		ctx,
		minfo.String()[1:],
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
	md, _ := metadata.FromOutgoingContext(ctx)
	if md == nil {
		md = make(metadata.MD)
	}
	impl.propagator.Inject(ctx1, otPriv.TextMapCarrierFromGrpcMD{MD: md})
	return span, metadata.NewOutgoingContext(ctx1, md)
}

// TraceUnaryCalls unary interceptor
func (impl *GRPCTracer) TraceUnaryCalls(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) (err error) {
	span, ctx1 := impl.spanStart(ctx, method, cc.Target())
	defer func() {
		spanEndFromGRPC(span, err)
	}()
	err = invoker(ctx1, method, req, reply, cc, callOpts...)
	return
}

// TraceStreamCalls stream interceptor
func (impl *GRPCTracer) TraceStreamCalls(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
	streamer grpc.Streamer, callOpts ...grpc.CallOption) (grpc.ClientStream, error) {
	//
	span, ctx1 := impl.spanStart(ctx, method, cc.Target())
	clientStream, err := streamer(ctx1, desc, cc, method, callOpts...)
	if err != nil {
		spanEndFromGRPC(span, err)
		return nil, err
	}
	if span == nil {
		return clientStream, nil
	}
	wrapped := &clientStreamWrapper{
		ClientStream: clientStream,
		StreamDesc:   desc,
		span:         span,
	}
	if desc.ServerStreams {
		runtime.SetFinalizer(wrapped, func(o *clientStreamWrapper) {
			spanEndFromGRPC(o.span, nil)
		})
	}
	return wrapped, nil
}
