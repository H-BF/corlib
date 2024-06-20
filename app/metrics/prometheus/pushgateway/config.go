package pushgateway

import (
	"context"
	"net/http"
	"time"

	"github.com/H-BF/corlib/pkg/backoff"
	"github.com/H-BF/corlib/pkg/scheduler"
	"github.com/H-BF/corlib/pkg/tm"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

// EndpointUrlProvider gateway endpoint URL provider
type EndpointUrlProvider = func(ctx context.Context) (string, error) //nolint:revive

// TaskManagerProvider ...
type TaskManagerProvider = func(ctx context.Context) tm.TaskManger

// Strategy PUSH|ADD strategy
type Strategy int

const (
	//PushStrategy use to substitute data in queue
	PushStrategy Strategy = iota

	//AddStrategy use to add data in queue
	AddStrategy

	//DelStrategy use to remove data from queue
	DelStrategy
)

// Config Gateway client config
type Config struct {
	JobName         string                 //required
	GwEndpointURL   EndpointUrlProvider    //required
	Gatherers       prometheus.Gatherers   //?required
	Collectors      []prometheus.Collector //?required
	JobScheduler    scheduler.Scheduler    //required
	JobBackoff      backoff.Backoff
	AuthProvider    AuthProvider
	HttpClient      *http.Client //nolint:revive
	RequestDuration time.Duration
	Strategy        Strategy
	ExpFmt          expfmt.Format
	TaskManagerProvider
}

// Validate check if config is valid
func (c *Config) Validate() error {
	const api = "config.Validate"

	if len(c.JobName) == 0 {
		return errors.Wrap(errors.New("no 'JobName' is provided"), api)
	}
	if len(c.Collectors) == 0 && len(c.Gatherers) == 0 {
		return errors.Wrap(errors.New("'Collectors' or/and Gatherers' is/are required"), api)
	}
	if c.JobScheduler == nil {
		return errors.Wrap(errors.New("'JobScheduler' is required"), api)
	}
	if c.GwEndpointURL == nil {
		return errors.Wrap(errors.New("'GwEndpointURL' is required"), api)
	}
	return nil
}
