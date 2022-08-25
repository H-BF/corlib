package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	ot4client "github.com/H-BF/corlib/client/trace/ot"
	"github.com/H-BF/corlib/logger"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/tests/strlib"
	"github.com/H-BF/corlib/server/trace/ot"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	stdoutTrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	sdkTraceTest "go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGrpc2GrpcTrace(t *testing.T) {
	logger.SetLevel(zap.InfoLevel)

	endpoint, err := pkgNet.ParseEndpoint("tcp://127.0.0.1:7003")
	if !assert.NoError(t, err) {
		return
	}

	ass := new(testOTelTracerAssist)
	serverSpanStore := bytes.NewBuffer(nil)
	clientSpanStore := bytes.NewBuffer(nil)
	serverSpanRecorder := sdkTraceTest.NewSpanRecorder()
	clientSpanRecorder := sdkTraceTest.NewSpanRecorder()
	serverTraceProvider := ass.makeTraceProvider(
		ass.ioWriter2SpanProcessor(serverSpanStore),
		serverSpanRecorder,
	)
	clientTraceProvider := ass.makeTraceProvider(
		ass.ioWriter2SpanProcessor(clientSpanStore),
		clientSpanRecorder,
	)

	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	clientCtx, clientCancel := context.WithCancel(context.Background())
	defer clientCancel()

	runners := []func() error{
		func() error {
			defer clientCancel()
			defer serverCancel()
			select {
			case <-serverCtx.Done():
			case <-clientCtx.Done():
			case <-time.After(20 * time.Second):
				return context.DeadlineExceeded
			}
			return nil
		},
		func() error {
			defer clientCancel()
			defer serverCancel()
			srv, e := ass.makeServer(serverTraceProvider)
			if e != nil {
				return e
			}
			return srv.Run(serverCtx, endpoint)
		},
		func() error {
			defer clientCancel()
			defer serverCancel()
			cc, e := ass.makeClientConn(clientCtx, endpoint, clientTraceProvider)
			if e != nil {
				return e
			}
			defer cc.Close()
			client := strlib.NewStrlibClient(cc)
			_, e = client.Uppercase(clientCtx, &strlib.UppercaseQuery{Value: "qwerty"})
			return e
		},
	}
	err = parallel.ExecAbstract(len(runners), int32(len(runners)-1), func(i int) error {
		return runners[i]()
	})
	if !assert.NoError(t, err) {
		return
	}

	_ = serverTraceProvider.ForceFlush(context.Background())
	_ = clientTraceProvider.ForceFlush(context.Background())
	if !assert.True(t, len(serverSpanRecorder.Ended()) > 0 && len(serverSpanRecorder.Ended()) == len(serverSpanRecorder.Started())) {
		return
	}
	if !assert.True(t, len(clientSpanRecorder.Ended()) > 0 && len(clientSpanRecorder.Ended()) == len(clientSpanRecorder.Started())) {
		return
	}
	serverSpanCtx := serverSpanRecorder.Ended()[0].SpanContext()
	clientSpanCtx := clientSpanRecorder.Ended()[0].SpanContext()
	if !assert.True(t, serverSpanCtx.IsValid()) {
		return
	}
	if !assert.True(t, clientSpanCtx.IsValid()) {
		return
	}
	if !assert.Equal(t, clientSpanCtx.TraceID(), serverSpanCtx.TraceID()) {
		return
	}
	if !assert.NotEqual(t, clientSpanCtx.SpanID(), serverSpanCtx.SpanID()) {
		return
	}

	t.Log("\n\r--------======= ALL RIGHT =======--------\n",
		"server-span:", serverSpanStore.String(),
		"\nclient-span:", clientSpanStore.String(),
	)

}

func TestGW2GrpcTrace(t *testing.T) {
	logger.SetLevel(zap.InfoLevel)

	endpoint, err := pkgNet.ParseEndpoint("tcp://127.0.0.1:7004")
	if !assert.NoError(t, err) {
		return
	}

	ass := new(testOTelTracerAssist)
	serverSpanStore := bytes.NewBuffer(nil)
	clientSpanStore := bytes.NewBuffer(nil)
	serverSpanRecorder := sdkTraceTest.NewSpanRecorder()
	clientSpanRecorder := sdkTraceTest.NewSpanRecorder()
	serverTraceProvider := ass.makeTraceProvider(
		ass.ioWriter2SpanProcessor(serverSpanStore),
		serverSpanRecorder,
	)
	clientTraceProvider := ass.makeTraceProvider(
		ass.ioWriter2SpanProcessor(clientSpanStore),
		clientSpanRecorder,
	)

	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	clientCtx, clientCancel := context.WithCancel(context.Background())
	defer clientCancel()

	runners := []func() error{
		func() error {
			defer clientCancel()
			defer serverCancel()
			select {
			case <-serverCtx.Done():
			case <-clientCtx.Done():
			case <-time.After(20 * time.Second):
				return context.DeadlineExceeded
			}
			return nil
		},
		func() error {
			defer clientCancel()
			defer serverCancel()
			srv, e := ass.makeServer(serverTraceProvider)
			if e != nil {
				return e
			}
			return srv.Run(serverCtx, endpoint)
		},
		func() error {
			defer clientCancel()
			defer serverCancel()
			httpClient := ass.makeHTTPClient(clientTraceProvider)
			apiURL := "http://" + endpoint.String() + "/v1/uppercase"
			req, e := http.NewRequest(http.MethodPost, apiURL, nil)
			if e != nil {
				return e
			}
			req = req.WithContext(clientCtx)
			var resp *http.Response
			resp, e = httpClient.Do(req)
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			return e
		},
	}
	err = parallel.ExecAbstract(len(runners), int32(len(runners)-1), func(i int) error {
		return runners[i]()
	})
	if !assert.NoError(t, err) {
		return
	}
	_ = serverTraceProvider.ForceFlush(context.Background())
	_ = clientTraceProvider.ForceFlush(context.Background())

	if !assert.True(t, len(serverSpanRecorder.Ended()) > 0 && len(serverSpanRecorder.Ended()) == len(serverSpanRecorder.Started())) {
		return
	}
	if !assert.True(t, len(clientSpanRecorder.Ended()) > 0 && len(clientSpanRecorder.Ended()) == len(clientSpanRecorder.Started())) {
		return
	}
	serverSpanCtx := serverSpanRecorder.Ended()[0].SpanContext()
	clientSpanCtx := clientSpanRecorder.Ended()[0].SpanContext()
	if !assert.True(t, serverSpanCtx.IsValid()) {
		return
	}
	if !assert.True(t, clientSpanCtx.IsValid()) {
		return
	}
	if !assert.Equal(t, clientSpanCtx.TraceID(), serverSpanCtx.TraceID()) {
		return
	}
	if !assert.NotEqual(t, clientSpanCtx.SpanID(), serverSpanCtx.SpanID()) {
		return
	}

	t.Log("\n\r--------======= ALL RIGHT =======--------\n",
		"server-span:", serverSpanStore.String(),
		"\nclient-span:", clientSpanStore.String(),
	)

}

type testOTelTracerAssist struct{}

func (ass *testOTelTracerAssist) makeClientConn(ctx context.Context, endpoint *pkgNet.Endpoint, tp trace.TracerProvider) (*grpc.ClientConn, error) {
	opts := []grpcRetry.CallOption{
		grpcRetry.WithBackoff(grpcRetry.BackoffExponential(5 * time.Millisecond)),
		grpcRetry.WithMax(100),
	}
	t := ot4client.NewClientGRPCTracer(tp)
	return grpc.DialContext(ctx,
		endpoint.String(),
		grpc.WithUserAgent("test-tracer"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(grpcRetry.UnaryClientInterceptor(opts...), t.TraceUnaryCalls),
		grpc.WithChainStreamInterceptor(t.TraceStreamCalls),
	)
}

func (ass *testOTelTracerAssist) makeServer(tp trace.TracerProvider) (*server.APIServer, error) {
	service := new(StrLibImpl)
	service.ProvideMock().
		On("Uppercase", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) {
			span := trace.SpanFromContext(ctx)
			span.AddEvent("do 'Uppercase'")
			logger.Info(ctx, "do 'Uppercase'")
			defer func() {
				span.AddEvent("did 'Uppercase'")
			}()
			v := req.GetValue()
			return &strlib.UppercaseResponse{Value: strings.ToUpper(v)}, nil
		})
	opts := []server.APIServerOption{server.WithServices(service)}
	if tp != nil {
		grpcTracer := ot.NewGRPCServerTracer(ot.WithTracerProvider(tp))
		opts = append(opts, server.WithTracer(grpcTracer))
	}
	return server.NewAPIServer(opts...)
}

func (ass *testOTelTracerAssist) ioWriter2SpanProcessor(w io.Writer) sdkTrace.SpanProcessor {
	exp, _ := stdoutTrace.New(
		stdoutTrace.WithWriter(w),
		stdoutTrace.WithPrettyPrint(),
	)
	return sdkTrace.NewSimpleSpanProcessor(exp)
}

func (ass *testOTelTracerAssist) makeTraceProvider(spanProcessors ...sdkTrace.SpanProcessor) *sdkTrace.TracerProvider {
	opts := []sdkTrace.TracerProviderOption{
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
	}
	for _, sp := range spanProcessors {
		opts = append(opts, sdkTrace.WithSpanProcessor(sp))
	}
	return sdkTrace.NewTracerProvider(opts...)
}

func (ass *testOTelTracerAssist) makeHTTPClient(tp trace.TracerProvider) *http.Client {
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = nil
	httpClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}
	httpClient.RetryMax = 2
	httpClient.RetryWaitMin = time.Millisecond
	httpClient.RetryWaitMax = 10 * time.Millisecond
	client := httpClient.StandardClient()
	return ot4client.WrapClient(client, tp)
}
