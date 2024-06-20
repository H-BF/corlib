package server

import (
	"context"

	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/pkg/parallel"

	"github.com/pkg/errors"
)

// RunAPIServersOption option interface to call RunAPIServers
type RunAPIServersOption interface {
	apply(*runAPIServersOptions) error
}

// Run can run one or more API servers
func (srv *APIServer) Run(ctx context.Context, endpoint *pkgNet.Endpoint, options ...RunAPIServersOption) error {
	const api = "APIServer.Run"

	runnerOptions := &runAPIServersOptions{ctx: ctx}
	err := RunWithAPIServer(endpoint, srv).apply(runnerOptions)
	if err != nil {
		return errors.Wrap(err, api)
	}
	for _, o := range options {
		if err = o.apply(runnerOptions); err != nil {
			return errors.Wrapf(err, "%s: applying options", api)
		}
	}
	var ass runAPIServersAssistant
	ass.init()
	defer ass.cleanup()
	if err = ass.construct(runnerOptions); err != nil {
		return errors.Wrap(err, api)
	}

	var runners []func() error
	runners = append(runners, ass.makeWait2CloseRunner(ctx))
	runners = append(runners, ass.makeMultiplexersRunners(ctx)...)
	runners = append(runners, ass.makeGrpcRunners(ctx, runnerOptions.gracefulStopPeriod)...)
	runners = append(runners, ass.makeGwRunners(ctx, runnerOptions.gracefulStopPeriod)...)
	err = parallel.ExecAbstract(len(runners), int32(len(runners))-1, func(i int) error {
		return runners[i]()
	})
	return errors.Wrap(err, api)
}
