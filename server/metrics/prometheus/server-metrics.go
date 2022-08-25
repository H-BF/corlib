package prometheus_metrics

import (
	"github.com/H-BF/corlib/server/interceptors"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	//Option опции метрик
	Option interface {
		apply(*serverMetricsOptions)
	}

	serverMetricsOptions struct {
		Namespace string
		Subsystem string
	}

	//ServerMetrics серверные метрики
	ServerMetrics struct {
		collectors     []prometheus.Collector
		panicsObserver interceptors.OnPanicEventObserver
	}

	serverMetricsOptionApplier func(*serverMetricsOptions)

	panicsObserver interface {
		observePanic(interceptors.OnPanicEvent)
	}
)

const ( //possible metrics labels
	LabelLocalAddr  string = "local_address"  //nolint
	LabelRemoteAddr        = "remote_address" //nolint
	LabelClientName        = "client_name"    //nolint
	LabelService           = "service"        //nolint
	LabelMethod            = "method"         //nolint
	LabelState             = "state"          //nolint
	LabelGRPCCode          = "grpc_code"      //nolint
)

const ( //label values
	Started  string = "started"  //nolint
	Finished        = "finished" //nolint
	Received        = "received" //nolint
	Sent            = "sent"     //nolint
)

const (
	DefaultNamespace = "sbr"         //nolint
	DefaultSubsystem = "grpc_server" //nolint
)

var (
	_ prometheus.Collector = (*ServerMetrics)(nil)
	_ Option               = (serverMetricsOptionApplier)(nil)
)

//NewMetrics ...
func NewMetrics(opts ...Option) *ServerMetrics {
	options := serverMetricsOptions{
		Namespace: DefaultNamespace,
		Subsystem: DefaultSubsystem,
	}
	ret := new(ServerMetrics)
	for _, o := range opts {
		o.apply(&options)
	}

	collectors := append(ret.collectors,
		newConnectionsCountMetric(options),
		newTotalRequestsMetrics(options),
		newResponseTimeHistogram(options))
	ret.collectors = collectors

	var panicObservers []panicsObserver
	for _, coll := range collectors {
		if obs, ok := coll.(panicsObserver); ok {
			panicObservers = append(panicObservers, obs)
		}
	}
	ret.panicsObserver = func(event interceptors.OnPanicEvent) {
		for _, o := range panicObservers {
			o.observePanic(event)
		}
	}
	return ret
}

//PanicsObserver observer panics
func (pMetrics *ServerMetrics) PanicsObserver() interceptors.OnPanicEventObserver {
	return pMetrics.panicsObserver
}

//StatHandlers ...
func (pMetrics *ServerMetrics) StatHandlers() []interceptors.StatsHandler {
	var ret []interceptors.StatsHandler
	for _, coll := range pMetrics.collectors {
		if h, _ := coll.(interceptors.StatsHandler); h != nil {
			ret = append(ret, h)
		}
	}
	return ret
}

//Describe impl prometheus.Collector
func (pMetrics *ServerMetrics) Describe(c chan<- *prometheus.Desc) {
	for _, coll := range pMetrics.collectors {
		coll.Describe(c)
	}
}

//Collect impl prometheus.Collector
func (pMetrics *ServerMetrics) Collect(c chan<- prometheus.Metric) {
	for _, coll := range pMetrics.collectors {
		coll.Collect(c)
	}
}

//WithNamespace sets Namespace to metrics
func WithNamespace(ns string) Option {
	var ret serverMetricsOptionApplier = func(options *serverMetricsOptions) {
		options.Namespace = ns
	}
	return ret
}

//WithSubsystem  sets Subsystem to metrics
func WithSubsystem(ss string) Option {
	var ret serverMetricsOptionApplier = func(options *serverMetricsOptions) {
		options.Subsystem = ss
	}
	return ret
}

func (f serverMetricsOptionApplier) apply(o *serverMetricsOptions) {
	f(o)
}
