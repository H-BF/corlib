package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type (
	//UnaryInterceptor ...
	UnaryInterceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)

	//StreamInterceptor ...
	StreamInterceptor = func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error

	//StatsHandler alias to stats.Handler
	StatsHandler = stats.Handler
)
