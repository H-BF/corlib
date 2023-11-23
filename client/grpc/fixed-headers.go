package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// FixedHeadersUnaryInterceptor - specifies fixed headers for every rpc as kv array
func FixedHeadersUnaryInterceptor(kv ...string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, kv...)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// FixedHeadersStreamInterceptor - specifies fixed headers for every rpc as kv array
func FixedHeadersStreamInterceptor(kv ...string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, kv...)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
