package interceptors

import (
	"context"

	"github.com/H-BF/corlib/pkg/conventions"
	"google.golang.org/grpc/stats"
)

//NewMethodNamePropagator ...
func NewMethodNamePropagator() stats.Handler {
	return new(methodNamePropagator)
}

type methodNamePropagator struct {
	StatsHandlerBase
}

//TagRPC ...
func (h *methodNamePropagator) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	var p conventions.GrpcMethodInfo
	if err := p.Init(info.FullMethodName); err != nil {
		panic(err)
	}
	return p.WrapContext(ctx)
}
