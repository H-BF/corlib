package pushgateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/H-BF/corlib/app/jobs"
	"github.com/H-BF/corlib/logger"
	netPkg "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	"github.com/H-BF/corlib/pkg/scheduler"
	"github.com/H-BF/corlib/pkg/tm"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func genTestSocketName() string {
	time.Sleep(100 * time.Millisecond)
	return path.Join("/tmp", fmt.Sprintf("test-%v-%v.socket", os.Getpid(), time.Now().Nanosecond()))
}

func newTestHttpClient() *http.Client { //nolint:revive
	return netPkg.UDS.EnrichClient(new(http.Client))
}

func job4gateway(ctx context.Context, endPoint *netPkg.Endpoint) error {
	metric := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ns",
		Subsystem: "test",
		Name:      "test",
		Help:      "test",
	})
	metric.Inc()
	jobConf := Config{
		JobName: "test-on-network/" + endPoint.Network(),
		GwEndpointURL: func(_ context.Context) (string, error) {
			nw := endPoint.Network()
			switch {
			case strings.EqualFold(nw, "tcp"):
				return string(netPkg.SchemeHTTP) + "://" + endPoint.String(), nil
			case strings.EqualFold(nw, "unix"):
				return string(netPkg.SchemeUnixHTTP) + "://" + endPoint.String(), nil
			}
			return "", errors.Errorf("unsupported tenwork '%s'", nw)
		},
		JobScheduler: scheduler.NewConstIntervalScheduler(5 * time.Second),
		Collectors:   []prometheus.Collector{metric},
		HttpClient:   newTestHttpClient(),
	}
	job, err := NewJob(ctx, jobConf)
	if err != nil {
		return err
	}
	defer job.Close()
	fin := make(chan error, 1)
	defer close(fin)
	eventObserve := func(event observer.EventType) {
		switch t := event.(type) {
		case jobs.OnJobLog:
			logger.Info(ctx, t)
		case jobs.OnJobFinished:
			select {
			case fin <- t.FindError():
			default:
			}
		}
	}
	obs := observer.NewObserver(eventObserve, false)
	jobs.SubscribeOnAllEvents(obs)
	job.Subject().ObserversAttach(obs)
	job.Schedule()
	job.Enable(true)

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-fin:
	}

	return err
}

func Test_PushGatewayJob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	logger.SetLevel(zap.InfoLevel)
	ctx = tm.TaskManagerToContext(ctx, tm.NewTaskManager())
	defer cancel()
	endpoints := []string{"tcp://127.0.0.1:7005",
		"unix://" + genTestSocketName()}

	for _, ep := range endpoints {
		ok := t.Run("on-endpoint('"+ep+"')", func(t2 *testing.T) {
			ep, err := netPkg.ParseEndpoint(ep)
			if !assert.NoError(t2, err) {
				return
			}
			var lst net.Listener
			lst, err = netPkg.Listen(ep)
			if !assert.NoError(t2, err) {
				return
			}
			defer lst.Close() //nolint:gosec

			var gotMethod string
			srv := http.Server{
				ReadHeaderTimeout: 0,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", `text/plain; charset=utf-8`)
					if r.Method == http.MethodDelete {
						w.WriteHeader(http.StatusAccepted)
						return
					}
					gotMethod = r.Method
					w.WriteHeader(http.StatusOK)
				}),
			}

			paraJobs := []func() error{
				func() error {
					_ = srv.Serve(lst)
					return nil
				},
				func() error {
					e := job4gateway(ctx, ep)
					_ = srv.Shutdown(ctx)
					return e
				},
			}
			err = parallel.ExecAbstract(len(paraJobs), 1, func(i int) error {
				return paraJobs[i]()
			})
			if !assert.NoError(t2, err) {
				return
			}
			assert.Equal(t2, http.MethodPut, gotMethod)
		})
		if !ok {
			return
		}
	}
}
