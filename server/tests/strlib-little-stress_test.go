package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/H-BF/corlib/logger"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/tests/strlib"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type littleStressSuite struct {
	endpoint         *pkgNet.Endpoint
	t                *testing.T
	grpcRequestsSent int
	grpcRequestsGot  int
	httpRequestsSent int
	httpRequestsGot  int
}

func (sui *littleStressSuite) runServer(runningPeriod time.Duration, shutdownPeriod time.Duration) <-chan error {
	chErr := make(chan error, 1)
	var (
		ch1 chan error
		ch2 chan error
		err error
	)
	sui.grpcRequestsSent = 0
	sui.grpcRequestsGot = 0
	sui.httpRequestsSent = 0
	sui.httpRequestsGot = 0
	defer func() {
		if err != nil {
			select {
			case chErr <- err:
			default:
			}
			close(chErr)
		}
	}()

	service := &WithOnStartStopEvents{
		new(StrLibImpl),
	}

	var (
		ctxReq       context.Context
		ctxReqCancel func()
	)

	service.ProvideMock().
		On("Uppercase", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) {
			v := req.GetValue()
			return &strlib.UppercaseResponse{Value: strings.ToUpper(v)}, nil
		}).
		On("OnStart").
		Return(func() {
			go func() {
				ch1 = make(chan error, 1)
				ch1 <- sui.runGrpcRequests(ctxReq)
				close(ch1)
			}()
			go func() {
				ch2 = make(chan error, 1)
				ch2 <- sui.runGwRequests(ctxReq)
				close(ch2)
			}()
		}).
		On("OnStop").
		Return(func() {
			ctxReqCancel()
		})

	var s *server.APIServer
	s, err = server.NewAPIServer(server.WithServices(service))
	if err != nil {
		return chErr
	}
	go func() {
		defer close(chErr)
		ctx, cancel := context.WithTimeout(context.Background(), runningPeriod)
		defer cancel()

		ctxReq, ctxReqCancel = context.WithCancel(context.Background())
		defer ctxReqCancel()

		var e error
		select {
		case chErr <- s.Run(ctx, sui.endpoint, server.RunWithGracefulStop(shutdownPeriod)):
		case e = <-ch1:
		case e = <-ch2:
		}
		logger.Infof(ctx, "GRPC-Stat[Sent:%v, Got:%v]", sui.grpcRequestsSent, sui.grpcRequestsGot)
		logger.Infof(ctx, "GW-Stat[Sent:%v, Got:%v]", sui.httpRequestsSent, sui.httpRequestsGot)
		if e != nil {
			chErr <- e
		}
	}()
	return chErr
}

func (sui *littleStressSuite) runGrpcRequests(ctx context.Context) error {
	const api = "run-GRPC-Requests"
	gConn, err := grpc.DialContext(ctx, sui.endpoint.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return errors.Wrapf(err, "%s: connect to GRPC server", api)
	}
	logger.Info(ctx, "GRPC requests started")
	defer func() {
		logger.Info(ctx, "GRPC requests stopped")
	}()
	defer gConn.Close() //nolint
	client := strlib.NewStrlibClient(gConn)
	sampleString := []byte("qwertyuiopasdfghjklzxcvbnm")
	for {
		sui.grpcRequestsSent++
		expected := string(bytes.ToUpper(sampleString))
		req := strlib.UppercaseQuery{Value: string(sampleString)}
		var resp *strlib.UppercaseResponse
		if resp, err = client.Uppercase(ctx, &req); err != nil {
			if c := status.Code(errors.Cause(err)); c == codes.Canceled {
				return nil
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return errors.Wrap(err, api)
		}
		if gotResult := resp.GetValue(); expected != gotResult {
			return errors.Errorf("%s: '%s' != '%s'", api, expected, gotResult)
		}
		sui.grpcRequestsGot++
		rand.Shuffle(len(sampleString), func(i, j int) {
			sampleString[i], sampleString[j] = sampleString[j], sampleString[i]
		})
	}
}

func (sui *littleStressSuite) runGwRequests(ctx context.Context) error {
	const api = "run-GW-Requests"

	logger.Info(ctx, "GW requests started")
	defer func() {
		logger.Info(ctx, "GW requests stopped")
	}()

	apiURL := "http://" + sui.endpoint.String() + "/v1/uppercase"
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
	sampleString := []byte("qwertyuiopasdfghjklzxcvbnm")

	bodyReader := new(struct {
		io.Reader
	})

	buf := bytes.NewBuffer(nil)
	jsonEnc := json.NewEncoder(buf)
	jsonDec := json.NewDecoder(bodyReader)
	req := new(strlib.UppercaseQuery)

	for {
		sui.httpRequestsSent++
		rand.Shuffle(len(sampleString), func(i, j int) {
			sampleString[i], sampleString[j] = sampleString[j], sampleString[i]
		})
		req.Value = string(sampleString)
		buf.Reset()
		err := jsonEnc.Encode(&req)
		if err != nil {
			return errors.Wrapf(err, "%s: make POST json-payload", api)
		}
		req, _ := retryablehttp.NewRequest(http.MethodPost, apiURL, buf)
		req = req.WithContext(ctx)
		//req.Close = true
		req.Header.Set("Content-Type", "application/json")
		var httpResp *http.Response
		httpResp, err = httpClient.Do(req)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			logger.Error(ctx, err)
			continue
		}
		if httpResp == nil {
			return errors.Errorf("%s: response == NIL", api)
		}
		if httpResp.StatusCode != http.StatusOK {
			return errors.Errorf("%s: status: %v", api, httpResp.StatusCode)
		}
		if httpResp.Body == nil {
			return errors.Errorf("%s: response.Body == NIL", api)
		}
		bodyReader.Reader = httpResp.Body
		var r strlib.UppercaseResponse
		err = jsonDec.Decode(&r)
		_ = httpResp.Body.Close()
		if err != nil {
			return errors.Errorf("%s: decode body tp JSOM", api)
		}
		expected := strings.ToUpper(string(sampleString))
		if res := r.GetValue(); expected != res {
			return errors.Errorf("%s: expected '%s' != result '%s'", expected, res, api)
		}
		sui.httpRequestsGot++
	}
}

func TestLittleStress(t *testing.T) {
	logger.SetLevel(zap.InfoLevel)
	sui := &littleStressSuite{t: t}
	var err error
	sui.endpoint, err = pkgNet.ParseEndpoint("tcp://127.0.0.1:7001")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	for err = range sui.runServer(5*time.Second, 30*time.Second) {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			continue
		}
		assert.NoError(t, err)
	}
}
