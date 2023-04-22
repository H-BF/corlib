package tests

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/H-BF/corlib/logger"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/H-BF/corlib/server"
	"github.com/H-BF/corlib/server/interceptors"
	prometheusMetrics "github.com/H-BF/corlib/server/metrics/prometheus"
	"github.com/H-BF/corlib/server/tests/strlib"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

//TestPrometheusServerMetrics ...
func TestPrometheusServerMetrics(t *testing.T) {
	logger.SetLevel(zap.InfoLevel)

	var (
		endpoint *pkgNet.Endpoint
		err      error
	)
	endpoint, err = pkgNet.ParseEndpoint("tcp://127.0.0.1:7002")
	if !assert.NoError(t, err) {
		return
	}

	nnn := 0
	service := new(StrLibImpl)
	service.ProvideMock().
		On("Uppercase", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, req *strlib.UppercaseQuery) (*strlib.UppercaseResponse, error) {
			nnn++
			if nnn == 4 {
				panic(errors.New("123"))
			}
			v := req.GetValue()
			return &strlib.UppercaseResponse{Value: strings.ToUpper(v)}, nil
		})

	runCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	const (
		userAgent = "testUserAgent"
		namespace = "test"
		subsystem = "test"
	)
	pm := prometheusMetrics.NewMetrics(prometheusMetrics.WithSubsystem(subsystem),
		prometheusMetrics.WithNamespace(namespace))

	reg := prometheus.NewRegistry()
	err = reg.Register(pm)
	if !assert.NoError(t, err) {
		return
	}

	runners := []func() error{
		func() error { //сервер
			recovery := interceptors.NewRecovery(
				interceptors.RecoveryWithObservers(pm.PanicsObserver())) //подключаем prometheus счетчик паник

			srv, e := server.NewAPIServer(
				server.WithServices(service),
				server.WithRecovery(recovery),
				server.WithStatsHandlers(pm.StatHandlers()...), //подключаем prometheus метрики
			)

			if e != nil {
				cancel()
				return e
			}
			return srv.Run(runCtx, endpoint)
		},
		func() error { //клиент
			defer cancel()
			opts := []grpcRetry.CallOption{
				grpcRetry.WithBackoff(grpcRetry.BackoffExponential(5 * time.Millisecond)),
				grpcRetry.WithMax(100),
			}
			gConn, e := grpc.DialContext(runCtx, endpoint.String(),
				grpc.WithUserAgent(userAgent),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithUnaryInterceptor(grpcRetry.UnaryClientInterceptor(opts...)))
			if e != nil {
				return e
			}
			defer gConn.Close() //nolint
			strlibClient := strlib.NewStrlibClient(gConn)
			qry := strlib.UppercaseQuery{
				Value: "qawsedrf",
			}
			for i := 0; i < 20; i++ {
				_, e = strlibClient.Uppercase(runCtx, &qry)
				if status.Code(e) == codes.Unavailable {
					return e
				}
			}
			return nil
		},
	}
	//запускаем сервер и клент - ждем остановки
	err = parallel.ExecAbstract(len(runners), int32(len(runners))-1, func(i int) error {
		return runners[i]()
	})
	if !assert.NoError(t, err) {
		return
	}
	ha := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	recorder := httptest.NewRecorder()
	//запрашиваем статистику GRPC сервера
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	ha.ServeHTTP(recorder, r)
	recorder.Flush()
	resp := recorder.Result()
	if !assert.Equal(t, http.StatusOK, resp.StatusCode) {
		return
	}
	if !assert.NotNil(t, resp.Body) {
		return
	}
	defer resp.Body.Close() //nolint:gosec
	var payload []byte
	payload, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	//смотрим наличие статистики
	t.Log("\n----------== check-server-statistics ==----------\n", string(payload))
	rePat := []string{
		`(?m)test_test_connections.+local_address=".+\d+`,
		`(?m)test_test_messages.+method=".+service=".+state=".+\d+`,
		`(?m)test_test_methods_finished.+client_name="testUserAgent".+grpc_code=".+method=".+,service=".+\d+`,
		`(?m)test_test_methods_started.+client_name="testUserAgent".+method=".+service=".+\d+`,
		`(?m)test_test_response_time_bucket.+method=".+service=".+le=".+\d+`,
		`(?m)test_test_methods_panicked.+client_name="testUserAgent".+method=".+,service=".+\d+`,
	}
	for i := range rePat {
		pattern := rePat[i]
		re := regexp.MustCompile(pattern)
		found := re.FindIndex(payload)
		if !assert.NotEmptyf(t, found, "on-metric-pattern:'%s'", pattern) {
			return
		}
	}
	t.Log("\r--==all right==------------------------\n")
}
