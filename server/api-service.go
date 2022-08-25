package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

//APIService API service definition interface.
type APIService interface {
	Description() grpc.ServiceDesc
	RegisterGRPC(context.Context, *grpc.Server) error
}

//APIGatewayProxy optional additional to APIService interface
type APIGatewayProxy interface {
	RegisterProxyGW(context.Context, *runtime.ServeMux, *grpc.ClientConn) error //can return ErrNoGateway when no gateway for this API
}

//APIServiceOnStartEvent optional additional to APIService interface
type APIServiceOnStartEvent interface {
	OnStart()
}

//APIServiceOnStopEvent optional additional to APIService interface
type APIServiceOnStopEvent interface {
	OnStop()
}
