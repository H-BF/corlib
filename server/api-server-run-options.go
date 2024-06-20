package server

import (
	"context"
	"time"

	pkgNet "github.com/H-BF/corlib/pkg/net"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

// RunWithAPIServer add one more API server
func RunWithAPIServer(endpoint *pkgNet.Endpoint, srv *APIServer) RunAPIServersOption {
	return runAPIServersOptionsApplier(func(o *runAPIServersOptions) error {
		nw := endpoint.Network()
		switch nw {
		case "tcp", "unix":
		default:
			return errors.Errorf("unusable network '%s' from endpoint", nw)
		}
		o.apiServers = append(o.apiServers, struct {
			endpoint *pkgNet.Endpoint
			serv     *APIServer
		}{endpoint: endpoint, serv: srv})
		return nil
	})
}

// RunWithGracefulStop add time period to shutdown services gracefully
func RunWithGracefulStop(period time.Duration) RunAPIServersOption {
	return runAPIServersOptionsApplier(func(o *runAPIServersOptions) error {
		o.gracefulStopPeriod = period
		return nil
	})
}

var (
	_ *chi.Mux = nil
	_          = RunWithAPIServer
	_          = RunWithGracefulStop
)

type runAPIServersOptions struct {
	ctx                context.Context
	gracefulStopPeriod time.Duration
	apiServers         []struct {
		endpoint *pkgNet.Endpoint // "TCP" | "UNIX" address => tcp://192.168.1.1:500 | unix://path-to-socket
		serv     *APIServer
	}
}

type runAPIServersOptionsApplier func(o *runAPIServersOptions) error

func (f runAPIServersOptionsApplier) apply(o *runAPIServersOptions) error {
	return f(o)
}
