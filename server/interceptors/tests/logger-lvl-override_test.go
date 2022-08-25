package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/conventions"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/H-BF/corlib/server"
	srvTests "github.com/H-BF/corlib/server/tests"
	"github.com/H-BF/corlib/server/tests/strlib"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type uppercaseSubstitute = func(context.Context, *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error)

type (
	fishBone struct {
		serverOptions []server.APIServerOption
		endPt         *pkgNet.Endpoint
		v             atomic.Value
	}
	strLibClient struct {
		strlib.StrlibClient
		io.Closer
	}
)

func (fb *fishBone) client4GW() *retryablehttp.Client {
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = nil
	httpClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}
	httpClient.RetryMax = 100
	httpClient.RetryWaitMin = time.Millisecond
	httpClient.RetryWaitMax = 100 * time.Millisecond
	return httpClient
}

func (fb *fishBone) client4GRPC(ctx context.Context) (*strLibClient, error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(5 * time.Millisecond)),
		grpc_retry.WithMax(100),
	}
	gConn, err := grpc.DialContext(ctx, fb.endPt.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))

	if err != nil {
		return nil, err
	}
	return &strLibClient{
		StrlibClient: strlib.NewStrlibClient(gConn),
		Closer:       gConn,
	}, nil
}

func (fb *fishBone) newServer() (*server.APIServer, error) {
	service := new(srvTests.StrLibImpl)
	service.ProvideMock().
		On("Uppercase", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, r *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) {
			switch f := fb.v.Load().(type) {
			case uppercaseSubstitute:
				return f(ctx, r)
			}
			return nil, errors.New("nothing to execute")
		})
	opts := append([]server.APIServerOption{server.WithServices(service)}, fb.serverOptions...)
	return server.NewAPIServer(opts...)
}

func (fb *fishBone) urlUppercaseAPI() string {
	return "http://" + fb.endPt.String() + "/v1/uppercase"
}

func (fb *fishBone) gwReq2Uppercase(ctx context.Context, r *strlib.UppercaseQuery) (*retryablehttp.Request, error) {
	payload := bytes.NewBuffer(nil)
	e := json.NewEncoder(payload).Encode(r)
	if e != nil {
		return nil, e
	}
	var req *retryablehttp.Request
	req, e = retryablehttp.NewRequest(http.MethodPost, fb.urlUppercaseAPI(), payload)
	if e != nil {
		return nil, e
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func Test_LoggerLevelOverride(t *testing.T) {
	ep, err := pkgNet.ParseEndpoint("127.0.0.1:7100")
	if !assert.NoError(t, err) {
		return
	}
	buf := bytes.NewBuffer(nil)
	loggerSink := io.MultiWriter(os.Stdout, buf)
	l := logger.NewWithSink(zap.InfoLevel, loggerSink)
	logger.SetLogger(l)
	bone := &fishBone{endPt: ep}
	func2Test := func(ctx context.Context, query *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) { //nolint:unparam
		v := query.GetValue()
		var inf conventions.GrpcMethodInfo
		_ = inf.FromContext(ctx)
		logger.Debugf(ctx, "%s:Uppercase('%s')", inf.ServiceFQN, v)
		return new(strlib.UppercaseResponse), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	bone.v.Store(func2Test)
	var srv *server.APIServer
	srv, err = bone.newServer()
	if !assert.NoError(t, err) {
		return
	}
	const (
		checkDebugFromGRPC = "debug-from-grpc"
		checkDebugFromGW   = "debug-from-gw"
	)
	runners := []func() error{
		func() error {
			//time.Sleep(time.Second)
			e := srv.Run(ctx, bone.endPt)
			assert.NoError(t, e)
			return e
		},
		func() error {
			const (
				debugLevel = "DEBUG"
			)
			defer cancel()
			client, e := bone.client4GRPC(ctx)
			if !assert.NoError(t, e) {
				return e
			}
			defer client.Close()
			ctx1 := metadata.NewOutgoingContext(context.Background(),
				metadata.Pairs(conventions.LoggerLevelHeader, debugLevel))
			r := &strlib.UppercaseQuery{Value: checkDebugFromGRPC}
			_, e = client.Uppercase(ctx1, r)
			if !assert.NoError(t, e, e) {
				return e
			}
			r.Value = checkDebugFromGW
			var req *retryablehttp.Request
			req, e = bone.gwReq2Uppercase(ctx1, r)
			if !assert.NoError(t, e, e) {
				return e
			}
			req.Header.Add(conventions.LoggerLevelHeader, debugLevel)
			_, e = bone.client4GW().Do(req)
			assert.NoError(t, e, e)
			return e
		},
	}
	nRunners := len(runners)
	err = parallel.ExecAbstract(nRunners, int32(nRunners)-1, func(i int) error {
		return runners[i]()
	})
	if !assert.NoError(t, err) {
		return
	}
	_ = logger.FromContext(ctx).Sync()
	sampleOfLogger := buf.String()
	assert.GreaterOrEqual(t, strings.Index(sampleOfLogger, checkDebugFromGRPC), 0)
	assert.GreaterOrEqual(t, strings.Index(sampleOfLogger, checkDebugFromGW), 0)
	assert.GreaterOrEqual(t, strings.Index(sampleOfLogger, strlib.Strlib_ServiceDesc.ServiceName), 0)
}
