package prometheus_metrics

import (
	"context"

	"github.com/H-BF/corlib/pkg/conventions"
	"github.com/H-BF/corlib/server/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type totalRequestsMetric struct {
	interceptors.StatsHandlerBase
	messages       *prometheus.CounterVec
	methodStarted  *prometheus.CounterVec
	methodFinished *prometheus.CounterVec
	methodPanicked *prometheus.CounterVec
}

var _ stats.Handler = (*totalRequestsMetric)(nil)

func newTotalRequestsMetrics(options serverMetricsOptions) prometheus.Collector {
	messages := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "messages",
		Help:      "received and sent message counters",
	}, []string{LabelService, LabelMethod, LabelState})

	started := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "methods_started",
		Help:      "started methods counter",
	}, []string{LabelService, LabelMethod, LabelClientName})

	finished := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "methods_finished",
		Help:      "finished methods counter",
	}, []string{LabelService, LabelMethod, LabelClientName, LabelGRPCCode})

	panicked := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "methods_panicked",
		Help:      "panicked methods counter",
	}, []string{LabelService, LabelMethod, LabelClientName})

	return &totalRequestsMetric{
		messages:       messages,
		methodStarted:  started,
		methodFinished: finished,
		methodPanicked: panicked,
	}
}

func (met *totalRequestsMetric) Describe(c chan<- *prometheus.Desc) {
	collectors := []prometheus.Collector{met.messages, met.methodStarted, met.methodFinished, met.methodPanicked}
	for _, coll := range collectors {
		coll.Describe(c)
	}

}

func (met *totalRequestsMetric) Collect(c chan<- prometheus.Metric) {
	collectors := []prometheus.Collector{met.messages, met.methodStarted, met.methodFinished, met.methodPanicked}
	for _, coll := range collectors {
		coll.Collect(c)
	}
}

//HandleRPC ...
func (met *totalRequestsMetric) HandleRPC(ctx context.Context, stat stats.RPCStats) {
	if stat.IsClient() {
		return
	}
	var mi conventions.GrpcMethodInfo
	if !mi.FromContext(ctx) {
		return
	}
	labs := prometheus.Labels{
		LabelService: mi.ServiceFQN,
		LabelMethod:  mi.Method,
	}
	var vec *prometheus.CounterVec
	switch t := stat.(type) {
	case *stats.Begin:
		labs[LabelClientName] = conventions.ClientName.Incoming(ctx, "unknown")
		vec = met.methodStarted
	case *stats.End:
		labs[LabelClientName] = conventions.ClientName.Incoming(ctx, "unknown")
		labs[LabelGRPCCode] = status.Code(t.Error).String()
		vec = met.methodFinished
	case *stats.InPayload:
		labs[LabelState] = Received
		vec = met.messages
	case *stats.OutPayload:
		labs[LabelState] = Sent
		vec = met.messages
	default:
		return
	}
	vec.With(labs).Inc()
}

func (met *totalRequestsMetric) observePanic(event interceptors.OnPanicEvent) {
	labs := prometheus.Labels{
		LabelService:    event.Info.ServiceFQN,
		LabelMethod:     event.Info.Method,
		LabelClientName: conventions.ClientName.Incoming(event.Ctx, "unknown"),
	}
	met.methodPanicked.With(labs).Inc()
}
