package internal

import (
	"context"

	"google.golang.org/grpc"
)

//ServerStreamWithContext override ServerStream with Context
func ServerStreamWithContext(ctx context.Context, s grpc.ServerStream) grpc.ServerStream {
	return &serverStreamWithContext{
		ServerStream: s,
		ctx:          ctx,
	}
}

type serverStreamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *serverStreamWithContext) Context() context.Context {
	return ss.ctx
}
