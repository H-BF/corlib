package server

import (
	"context"
	"sync/atomic"

	"github.com/H-BF/corlib/server/health_check"

	"google.golang.org/grpc"
)

// HealthCheck optional to APIService interface
type HealthCheck interface {
	HealthProbe(ctx context.Context) (*health_check.Response, error)
}

type healthCheckService struct {
	health_check.Unimplemented
	services name2service
	ready    atomic.Value
}

// Description -
func (hc *healthCheckService) Description() grpc.ServiceDesc {
	return health_check.ServiceDesc
}

// RegisterGRPC -
func (hc *healthCheckService) RegisterGRPC(_ context.Context, srv *grpc.Server) error {
	health_check.RegisterGRPC(srv, hc)
	return nil
}

// OnStart -
func (hc *healthCheckService) OnStart() {
	hc.ready.Store(true)
}

// OnStop -
func (hc *healthCheckService) OnStop() {
	hc.ready.Store(false)
}

// HealthProbe -
func (hc *healthCheckService) HealthProbe(_ context.Context) (*health_check.Response, error) {
	resp := &health_check.Response{Status: health_check.StatusNotServing}
	switch imReady := hc.ready.Load().(type) {
	case bool:
		if imReady {
			resp.Status = health_check.StatusServing
		}
	}
	return resp, nil
}

// Check -
func (hc *healthCheckService) Check(ctx context.Context, req *health_check.Request) (*health_check.Response, error) {
	if ok, _ := hc.ready.Load().(bool); !ok {
		return &health_check.Response{Status: health_check.StatusNotServing}, nil
	}
	var resp *health_check.Response
	var err error
	service := req.GetService()
	if srv, ok := hc.services[service]; !ok {
		resp = &health_check.Response{Status: health_check.StatusServiceUnknown}
	} else if c, _ := srv.(HealthCheck); c != nil {
		resp, err = c.HealthProbe(ctx)
	} else {
		resp, err = hc.HealthProbe(ctx)
	}
	return resp, err
}
