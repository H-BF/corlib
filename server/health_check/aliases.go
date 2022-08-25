package health_check

import (
	"google.golang.org/grpc"
	hc "google.golang.org/grpc/health/grpc_health_v1"
)

type (
	ResponseStatus = hc.HealthCheckResponse_ServingStatus //nolint
	Request        = hc.HealthCheckRequest                //nolint
	Response       = hc.HealthCheckResponse               //nolint
	Unimplemented  = hc.UnimplementedHealthServer         //nolint
)

const (
	StatusUnknown        ResponseStatus = hc.HealthCheckResponse_UNKNOWN         //nolint
	StatusServing        ResponseStatus = hc.HealthCheckResponse_SERVING         //nolint
	StatusNotServing     ResponseStatus = hc.HealthCheckResponse_NOT_SERVING     //nolint
	StatusServiceUnknown ResponseStatus = hc.HealthCheckResponse_SERVICE_UNKNOWN //nolint
)

func RegisterGRPC(s grpc.ServiceRegistrar, srv hc.HealthServer) { //nolint
	s.RegisterService(&hc.Health_ServiceDesc, srv)
}

var (
	_           = RegisterGRPC
	ServiceDesc = hc.Health_ServiceDesc //nolint
	_           = StatusUnknown | StatusServing | StatusNotServing | StatusServiceUnknown
)
