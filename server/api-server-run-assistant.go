package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"reflect"
	rt "runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	otPriv "github.com/H-BF/corlib/internal/pkg/ot"
	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/conventions"
	"github.com/H-BF/corlib/pkg/events"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/server/interceptors"
	"github.com/H-BF/corlib/server/internal"
	"github.com/H-BF/corlib/server/swagger_ui"

	"github.com/go-chi/chi"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type runAPIServersAssistant struct {
	eventFailure  events.Event
	multiplexers  map[uintptr]cmux.CMux
	grpcListeners map[uintptr]net.Listener
	gwListeners   map[uintptr]net.Listener
	gwProxies     map[uintptr]*grpc.ClientConn
	grpcServers   map[uintptr]*grpc.Server
	httpServers   map[uintptr]*http.Server
	services      map[uintptr]name2service

	onceInit    sync.Once
	onceCleanup sync.Once
}

var (
	nextRunID = uintptr(0)
)

func (ass *runAPIServersAssistant) init() {
	ass.onceInit.Do(func() {
		ass.eventFailure = events.NewEvent(0)
		in := []interface{}{
			&ass.gwProxies,
			&ass.multiplexers,
			&ass.httpServers,
			&ass.gwListeners,
			&ass.grpcListeners,
			&ass.grpcServers,
			&ass.services,
		}
		for _, v := range in {
			t := reflect.TypeOf(v).Elem()
			reflect.ValueOf(v).Elem().Set(reflect.MakeMap(t))
		}
	})
}

func (ass *runAPIServersAssistant) cleanup() {
	ass.onceInit.Do(func() {})
	ass.onceCleanup.Do(func() {
		for _, closer := range ass.multiplexers {
			closer.Close()
		}
		for _, pxy := range ass.gwProxies {
			_ = pxy.Close()
		}
	})
}

func (ass *runAPIServersAssistant) newGRPCproxyConn(ctx context.Context, ep *pkgNet.Endpoint, tlsConf *tls.Config) (*grpc.ClientConn, error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(100),
	}
	var endpointAddr string
	if ep.IsUnixDomain() {
		endpointAddr = ep.FQN()
	} else {
		endpointAddr = ep.String()
	}
	creds := insecure.NewCredentials()
	if tlsConf != nil { //Use TLS-ed proxy
		newConf := &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
			Certificates:       tlsConf.Certificates,
		}
		creds = credentials.NewTLS(newConf)
	}
	dialOpts := []grpc.DialOption{
		grpc.WithUserAgent("grpc-proxy"),
		grpc.WithChainUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(opts...),
		),
		grpc.WithTransportCredentials(creds),
	}
	ret, err := grpc.DialContext(ctx, endpointAddr, dialOpts...)
	if err == nil {
		rt.SetFinalizer(ret, func(o *grpc.ClientConn) {
			_ = o.Close()
		})
		return ret, nil
	}
	return nil, err
}

func (ass *runAPIServersAssistant) construct(runner *runAPIServersOptions) error {
	for _, item := range runner.apiServers {
		server, endpoint, hasGrpcAPI := item.serv, item.endpoint, len(item.serv.apis) > 0

		var (
			gwOpts  []runtime.ServeMuxOption
			grpcS   *grpc.Server
			gw      *runtime.ServeMux
			err     error
			gwProxy *grpc.ClientConn
		)
		if hasGrpcAPI {
			grpcOpts := append([]grpc.ServerOption{},
				grpc.ChainUnaryInterceptor(server.grpcUnaryInterceptors...),
				grpc.ChainStreamInterceptor(server.grpcStreamInterceptors...),
			)
			grpcOpts = append(grpcOpts, server.grpcOptions...)
			if len(server.grpcStatsHandlers) > 0 {
				opt := grpc.StatsHandler(interceptors.Chain2StatsHandler(server.grpcStatsHandlers...))
				grpcOpts = append(grpcOpts, opt)
			}
			if len(server.grpcTapHandlers) > 0 {
				grpcOpts = append(grpcOpts, grpc.InTapHandle(interceptors.TapInHandleChain(server.grpcTapHandlers).TapInHandle))
			}
			if server.tlsConf != nil {
				creds := credentials.NewTLS(server.tlsConf)
				grpcOpts = append(grpcOpts, grpc.Creds(creds))
			}

			gwOpts = append(gwOpts, runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
				_ = ctx
				md, n, h := metadata.MD{}, len(conventions.SysHeaderPrefix), request.Header
				otPriv.TextMapCarrierFromGrpcMD{MD: md}.FillFromHTTPHeader(request.Header)
				for k, values := range h {
					doSet := len(k) >= len(conventions.SysHeaderPrefix) &&
						strings.EqualFold(k[:n], conventions.SysHeaderPrefix)
					/*//TODO: Remove this
					if !doSet {
						headers := [...]string{
							conventions.UserAgentHeader, conventions.HostNameHeader,
						}
						for i := 0; !doSet && i < len(headers); i++ {
							doSet = strings.EqualFold(k, headers[i])
						}
					}
					*/
					if doSet {
						md.Set(k, values...)
					}
				}
				return md
			}))
			gwOpts = append(gwOpts, server.gatewayOptions...)
			grpcS = grpc.NewServer(grpcOpts...)
		}
		i := atomic.AddUintptr(&nextRunID, 1)
		for serviceName, api := range server.apis {
			if err = api.RegisterGRPC(runner.ctx, grpcS); err != nil {
				return errors.Wrapf(err, "unable register service '%s'", serviceName)
			}
			if pxy, _ := api.(APIGatewayProxy); pxy != nil {
				if gw == nil {
					gw = runtime.NewServeMux(gwOpts...)
				}
				if gwProxy == nil {
					if gwProxy, err = ass.newGRPCproxyConn(runner.ctx, endpoint, server.tlsConf); err != nil {
						return errors.Wrapf(err, "unable create proxy gateway conn for endpoint '%s'", endpoint)
					}
					ass.gwProxies[i] = gwProxy
				}
				if err = pxy.RegisterProxyGW(runner.ctx, gw, gwProxy); err != nil {
					return errors.Wrapf(err, "unable register proxy gateway to service '%s'",
						serviceName)
				}
			}
		}
		nw, addr := endpoint.Network(), endpoint.String()
		if gw != nil || len(server.httpHandlers) > 0 {
			chiMux := chi.NewMux()
			for pattern, handler := range server.httpHandlers {
				chiMux.Mount(pattern, http.StripPrefix(pattern, handler))
			}
			if gw != nil {
				chiMux.Mount("/", gw)
				if server.docs != nil { //mount swagger documents
					server.docs.Host = addr
					suffix := strings.TrimSpace(server.urlDocsSuffix)
					_ = suffix
					var swaggerHandler http.Handler
					if swaggerHandler, err = swagger_ui.NewHandler(server.docs); err != nil {
						return errors.Wrap(err, "mount swagger docs")
					}
					chiMux.Mount("/docs", http.StripPrefix("/docs", swaggerHandler))
				}
			}
			ass.httpServers[i] = &http.Server{
				ReadHeaderTimeout: 0,
				Handler:           chiMux,
				BaseContext: func(_ net.Listener) context.Context {
					return runner.ctx
				},
			}
		}
		var mx cmux.CMux
		if mx, err = internal.NewCMux(endpoint); err != nil {
			return errors.Wrapf(err, "unable listen to '%s://%s'", nw, addr)
		}
		ass.multiplexers[i] = mx
		if gw != nil || len(server.httpHandlers) > 0 {
			ass.gwListeners[i] = mx.Match(cmux.HTTP1Fast())
		}
		if hasGrpcAPI {
			ass.services[i] = server.apis
			ass.grpcServers[i] = grpcS
			ass.grpcListeners[i] = mx.Match(cmux.Any())
			//mx.Match(cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc"))
			addGRPCreflector(grpcS)
		}
	}
	return nil
}

func (ass *runAPIServersAssistant) makeWait2CloseRunner(ctx context.Context) func() error {
	eventFailure := ass.eventFailure
	return func() error {
		select {
		case <-ctx.Done():
		case <-eventFailure.Done():
		}
		for _, m := range ass.multiplexers {
			m.Close()
		}
		return nil
	}
}

func (ass *runAPIServersAssistant) makeMultiplexersRunners(ctx context.Context) []func() error {
	var ret []func() error
	failureEvent := ass.eventFailure
	for _, mx := range ass.multiplexers {
		ret = append(ret, func() error {
			chErr := make(chan error, 1)
			go func() {
				chErr <- mx.Serve()
				close(chErr)
			}()
			var err error
			select {
			case err = <-chErr:
			case <-ctx.Done():
				err = ctx.Err()
			}
			noErr := errors.Is(err, cmux.ErrListenerClosed) ||
				errors.Is(err, cmux.ErrServerClosed) ||
				errors.Is(err, context.Canceled)
			if noErr {
				return nil
			}
			if opErr := (*net.OpError)(nil); errors.As(err, &opErr) {
				s := opErr.Err.Error()
				noErr = opErr.Op == "accept" &&
					strings.Contains(s, "closed") &&
					strings.Contains(s, "connection")
				if noErr {
					return nil
				}
			}
			failureEvent.Fire()
			return err
		})
	}
	return ret
}

func (ass *runAPIServersAssistant) notifyServerStart(id uintptr) {
	for _, s := range ass.services[id] {
		if f, _ := s.(APIServiceOnStartEvent); f != nil {
			f.OnStart()
		}
	}
}

func (ass *runAPIServersAssistant) notifyServerStop(id uintptr) {
	for _, s := range ass.services[id] {
		if f, _ := s.(APIServiceOnStopEvent); f != nil {
			f.OnStop()
		}
	}
}

func (ass *runAPIServersAssistant) makeGrpcRunners(ctx context.Context, gracefulStopPeriod time.Duration) []func() error {
	var runners []func() error
	failureEvent := ass.eventFailure

	for i, srv := range ass.grpcServers {
		listener := ass.grpcListeners[i]
		servID := fmt.Sprintf("%s://%s", listener.Addr().Network(), listener.Addr().String())
		runners = append(runners, func() (err error) { //nolint:dupl
			const api = "GRPC"
			logger.Infof(ctx, "[%s]: server '%s' is ready", api, servID)
			ass.notifyServerStart(i)
			defer func() {
				ass.notifyServerStop(i)
				err = errors.Wrapf(err, "%s: on serving '%s'", api, servID)
				failureEvent.FireIf(err != nil)
				ass.serviceRIP(ctx, srv, gracefulStopPeriod)
				logger.Infof(ctx, "[%s]: server '%s' has off", api, servID)
			}()
			err = srv.Serve(listener)
			if err != nil {
				select {
				case <-ctx.Done():
					err = nil
				default:
					for _, e := range []error{cmux.ErrServerClosed, cmux.ErrListenerClosed, grpc.ErrServerStopped} {
						if errors.Is(err, e) {
							err = nil
							break
						}
					}
				}
			}
			return
		})
	}
	return runners
}

func (ass *runAPIServersAssistant) makeGwRunners(ctx context.Context, gracefulStopPeriod time.Duration) []func() error {
	var runners []func() error
	eventFailure := ass.eventFailure

	for i, httpSrv := range ass.httpServers {
		listener := ass.gwListeners[i]
		servID := fmt.Sprintf("%s://%s", listener.Addr().Network(), listener.Addr().String())
		runners = append(runners, func() (err error) { //nolint:dupl
			const api = "HTTP"
			logger.Infof(ctx, "[%s]: server '%s' is ready", api, servID)
			defer func() {
				err = errors.Wrapf(err, "%s: on serving '%s'", api, servID)
				eventFailure.FireIf(err != nil)
				ass.serviceRIP(ctx, httpSrv, gracefulStopPeriod)
				logger.Infof(ctx, "[%s]: server '%s' has off", api, servID)
			}()
			err = httpSrv.Serve(listener)
			select {
			case <-ctx.Done():
				err = nil
			default:
				for _, e := range []error{cmux.ErrServerClosed, cmux.ErrListenerClosed, http.ErrServerClosed} {
					if errors.Is(err, e) {
						err = nil
						break
					}
				}
			}
			return
		})
	}
	return runners
}

func (ass *runAPIServersAssistant) serviceRIP(_ context.Context, s interface{}, gracefulStopPeriod time.Duration) {
	var (
		stop         func()
		gracefulStop func(context.Context)
	)
	switch v := s.(type) {
	case *grpc.Server:
		stop = v.Stop
		gracefulStop = func(_ context.Context) {
			v.GracefulStop()
		}
	case *http.Server:
		stop = func() {
			v.SetKeepAlivesEnabled(false)
			_ = v.Close()
		}
		gracefulStop = func(c context.Context) {
			v.SetKeepAlivesEnabled(false)
			_ = v.Shutdown(c)
		}
	default:
		return
	}
	if gracefulStopPeriod <= 0 {
		stop()
	} else {
		stopped := make(chan struct{})
		ctxGra, cancel := context.WithTimeout(context.Background(), gracefulStopPeriod)
		defer cancel()
		go func() {
			defer close(stopped)
			gracefulStop(ctxGra)
		}()
		select {
		case <-stopped:
			return
		case <-ass.eventFailure.Done():
			stop()
		case <-ctxGra.Done():
			stop()
		}
	}
}
