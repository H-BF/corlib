package server

import (
	"context"
	"net/http"

	"github.com/H-BF/corlib/server/interceptors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/tap"
)

type (
	//APIServer an API server
	APIServer struct {
		urlDocsSuffix          string
		docs                   *SwaggerSpec
		grpcOptions            []grpc.ServerOption
		gatewayOptions         []runtime.ServeMuxOption
		grpcUnaryInterceptors  []grpc.UnaryServerInterceptor
		grpcStreamInterceptors []grpc.StreamServerInterceptor
		grpcStatsHandlers      []stats.Handler
		grpcTapHandlers        []tap.ServerInHandle
		apis                   name2service
		httpHandlers           httpHandlers
		addDefInterceptors     interceptors.DefInterceptor
		recovery               *interceptors.Recovery
		grpcTracer             GRPCTracer
	}

	//GRPCTracer tracer
	GRPCTracer interface {
		TraceUnaryCalls(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
		TraceStreamCalls(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
	}

	//APIServerOption опции для APIServer
	APIServerOption interface {
		apply(*APIServer) error
	}
)

//NewAPIServer makes new API server
func NewAPIServer(options ...APIServerOption) (*APIServer, error) {
	const api = "NewAPIServer"

	ret := &APIServer{
		apis:               make(name2service),
		httpHandlers:       make(httpHandlers),
		addDefInterceptors: interceptors.DefAll,
	}
	ret.grpcStatsHandlers = append(ret.grpcStatsHandlers, interceptors.NewMethodNamePropagator())
	for _, o := range options {
		if err := o.apply(ret); err != nil {
			return nil, errors.Wrapf(err, "%s: applying otions", api)
		}
	}
	if len(ret.apis) == 0 {
		return &APIServer{httpHandlers: ret.httpHandlers, apis: ret.apis}, nil
	}
	err := ret.addService(&healthCheckService{
		services: ret.apis,
	})
	if err != nil {
		return nil, errors.Wrap(err, api)
	}

	var defUnary []grpc.UnaryServerInterceptor
	var defStream []grpc.StreamServerInterceptor
	if t := ret.grpcTracer; t != nil {
		defUnary = append(defUnary, t.TraceUnaryCalls)
		defStream = append(defStream, t.TraceStreamCalls)
	}

	var logMethods bool
	for i := interceptors.DefInterceptor(1); i&interceptors.DefAll == i; i <<= 1 {
		switch i {
		case interceptors.DefLogServerAPI:
			logMethods = i&ret.addDefInterceptors != 0
		case interceptors.DefRecovery:
			r := ret.recovery
			if i&ret.addDefInterceptors != 0 {
				r = interceptors.NewRecovery()
			}
			if r != nil {
				defStream = append(defStream, r.Stream)
				defUnary = append(defUnary, r.Unary)
			}
		case interceptors.DefLogLevelOverride:
			if i&ret.addDefInterceptors != 0 {
				r := interceptors.LogLevelOverrider
				defStream = append(defStream, r.Stream)
				defUnary = append(defUnary, r.Unary)
			}
		default:
			return nil, errors.Errorf("%s: unknown default-inerceptor-ID: %v", api, i)
		}
	}
	ret.grpcUnaryInterceptors = append(defUnary, ret.grpcUnaryInterceptors...)
	ret.grpcStreamInterceptors = append(defStream, ret.grpcStreamInterceptors...)
	if logMethods {
		ret.grpcUnaryInterceptors = append(ret.grpcUnaryInterceptors, interceptors.LogServerAPI.Unary)
		ret.grpcStreamInterceptors = append(ret.grpcStreamInterceptors, interceptors.LogServerAPI.Stream)
	}

	return ret, nil
}

type (
	name2service = map[string]APIService
	httpHandlers = map[string]http.Handler
)

func (srv *APIServer) addService(service APIService) error {
	d := service.Description()
	if _, isIn := srv.apis[d.ServiceName]; isIn {
		return errors.Errorf("service '%s' is always in", d.ServiceName)
	}
	srv.apis[d.ServiceName] = service
	return nil
}
