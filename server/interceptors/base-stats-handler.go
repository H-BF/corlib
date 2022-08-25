package interceptors

import (
	"context"

	"google.golang.org/grpc/stats"
)

//StatsHandlerBase empty impl of stats.Handler
type StatsHandlerBase struct{}

var _ stats.Handler = (*StatsHandlerBase)(nil)

//TagRPC impl stats.Handler
func (h StatsHandlerBase) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

//HandleRPC impl stats.Handler
func (h StatsHandlerBase) HandleRPC(_ context.Context, _ stats.RPCStats) {}

//TagConn impl stats.Handler
func (h StatsHandlerBase) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

//HandleConn impl stats.Handler
func (h StatsHandlerBase) HandleConn(_ context.Context, _ stats.ConnStats) {}
