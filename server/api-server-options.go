package server

import (
	"net/http"
	"strings"

	"github.com/H-BF/corlib/server/interceptors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/tap"
)

// WithDocs добавим сваггер доку
func WithDocs(docs *SwaggerSpec, anURLSuffix string) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		if srv.docs = docs; docs != nil {
			srv.urlDocsSuffix = anURLSuffix
		} else {
			srv.urlDocsSuffix = ""
		}
		return nil
	})
}

// WithServices добавим service APIs к серверу
func WithServices(services ...APIService) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		for _, s := range services {
			if err := srv.addService(s); err != nil {
				return err
			}
		}
		return nil
	})
}

// WithGrpcServerOptions добавим GRPC опции
func WithGrpcServerOptions(options ...grpc.ServerOption) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		srv.grpcOptions = append(srv.grpcOptions, options...)
		return nil
	})
}

// WithGatewayOptions добавим GW-2-GRPC опции
func WithGatewayOptions(options ...runtime.ServeMuxOption) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		srv.gatewayOptions = append(srv.gatewayOptions, options...)
		return nil
	})
}

// WithStatsHandlers добавим stats.Handler-ы
func WithStatsHandlers(handlers ...stats.Handler) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		srv.grpcStatsHandlers = append(srv.grpcStatsHandlers, handlers...)
		return nil
	})
}

// WithTapInHandlers добавим tap.ServerInHandle-ы
func WithTapInHandlers(handlers ...tap.ServerInHandle) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		srv.grpcTapHandlers = append(srv.grpcTapHandlers, handlers...)
		return nil
	})
}

// WithUnaryInterceptors добавим унарные перехвачики
func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		srv.grpcUnaryInterceptors = append(srv.grpcUnaryInterceptors, interceptors...)
		return nil
	})
}

// WithStreamInterceptors добавим стримовые перехватчики
func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		srv.grpcStreamInterceptors = append(srv.grpcStreamInterceptors, interceptors...)
		return nil
	})
}

// WithRecovery ...
func WithRecovery(recovery *interceptors.Recovery) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		srv.addDefInterceptors &= ^interceptors.DefRecovery
		srv.recovery = recovery
		return nil
	})
}

// SkipDefInterceptors ...
func SkipDefInterceptors(ids ...interceptors.DefInterceptor) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		for _, interceptorID := range ids {
			srv.addDefInterceptors &= ^(interceptorID & interceptors.DefAll)
		}
		return nil
	})
}

// WithTracer sets span tracer
func WithTracer(t GRPCTracer) APIServerOption {
	return serverOptApplier(func(srv *APIServer) error {
		srv.grpcTracer = t
		return nil
	})
}

// WithHttpHandler add HTTP handler for pattern
func WithHttpHandler(pattern string, handler http.Handler) APIServerOption { //nolint:revive
	return serverOptApplier(func(srv *APIServer) error {
		pattern = "/" + strings.Trim(
			strings.Replace(pattern, "\\", "/", -1),
			"/ ",
		)
		if len(pattern) > 1 && handler != nil {
			srv.httpHandlers[pattern] = handler
		}
		return nil
	})
}

type serverOptApplier func(srv *APIServer) error

func (f serverOptApplier) apply(srv *APIServer) error {
	return f(srv)
}

var (
	_ = WithHttpHandler
	_ = WithDocs
	_ = WithRecovery
	_ = WithServices
	_ = WithGrpcServerOptions
	_ = WithGatewayOptions
	_ = WithStatsHandlers
	_ = WithUnaryInterceptors
	_ = WithStreamInterceptors
	_ = SkipDefInterceptors
	_ = WithTapInHandlers
	_ = WithTracer
)
