package interceptors

import (
	"context"

	"google.golang.org/grpc/stats"
)

// Chain2StatsHandler wraps passed stats.Handlers in StatsHandlerChainWrapper
func Chain2StatsHandler(handlersChain ...stats.Handler) stats.Handler {
	return statsHandlerChain(handlersChain)
}

//statsHandlerChain implements stats.Handler interface
type statsHandlerChain []stats.Handler

var _ stats.Handler = (*statsHandlerChain)(nil)

// TagRPC attaches RPC tagInfo to all handlers in chain
func (h statsHandlerChain) TagRPC(ctx context.Context, tagInfo *stats.RPCTagInfo) context.Context {
	for _, ch := range h {
		ctx = ch.TagRPC(ctx, tagInfo)
	}
	return ctx
}

// HandleRPC processes RPCStats in all handlers
func (h statsHandlerChain) HandleRPC(ctx context.Context, s stats.RPCStats) {
	for _, ch := range h {
		ch.HandleRPC(ctx, s)
	}
}

// TagConn attaches conn tagInfo to all handlers in chain
func (h statsHandlerChain) TagConn(ctx context.Context, tagInfo *stats.ConnTagInfo) context.Context {
	for _, ch := range h {
		ctx = ch.TagConn(ctx, tagInfo)
	}

	return ctx
}

// HandleConn processes ConnStats in all handlers
func (h statsHandlerChain) HandleConn(ctx context.Context, s stats.ConnStats) {
	for _, ch := range h {
		ch.HandleConn(ctx, s)
	}
}
