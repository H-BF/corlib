package prometheus_metrics

import (
	"context"
	"time"

	"github.com/H-BF/corlib/pkg/conventions"
	"github.com/H-BF/corlib/server/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/stats"
)

func newResponseTimeHistogram(options serverMetricsOptions) prometheus.Collector {
	res := new(responseTimeHistogram)
	hist := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "response_time",
		Help:      "response time duration in milliseconds",
		Buckets:   res.defaultBucket(),
	}, []string{LabelService, LabelMethod})
	res.HistogramVec = hist
	return res
}

type responseTimeHistogram struct {
	interceptors.StatsHandlerBase
	*prometheus.HistogramVec
}

var _ stats.Handler = (*responseTimeHistogram)(nil)

func (met *responseTimeHistogram) defaultBucket() []float64 {
	return []float64{
		.0001, .0005, .00075, .001, .0025, .005, 0.0075, .01, 0.025, .05, 0.075,
		.1, .25, .5, .75, 10, 25, 50, 75, 100, 500, 1000}
}

func (met *responseTimeHistogram) HandleRPC(ctx context.Context, stat stats.RPCStats) {
	if stat.IsClient() {
		return
	}
	if end, _ := stat.(*stats.End); end != nil {
		var mi conventions.GrpcMethodInfo
		if !mi.FromContext(ctx) {
			return
		}
		labs := prometheus.Labels{
			LabelMethod:  mi.Method,
			LabelService: mi.ServiceFQN,
		}
		milliseconds := float64(end.EndTime.Sub(end.BeginTime)) / float64(time.Millisecond)
		met.With(labs).Observe(milliseconds)
	}
}
