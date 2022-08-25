package prometheus_metrics

import (
	"context"
	"fmt"

	"github.com/H-BF/corlib/server/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/stats"
)

func newConnectionsCountMetric(options serverMetricsOptions) prometheus.Collector {
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "connections",
		Help:      "connection count at moment on a server",
	}, []string{LabelLocalAddr})
	return &connMetric{GaugeVec: vec}
}

type connMetric struct {
	*prometheus.GaugeVec
	interceptors.StatsHandlerBase
	connectionTag int
}

var _ stats.Handler = (*connMetric)(nil)

func (met *connMetric) TagConn(ctx context.Context, tagInfo *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, &met.connectionTag, tagInfo)
}

//HandleConn ...
func (met *connMetric) HandleConn(ctx context.Context, stat stats.ConnStats) {
	if stat.IsClient() {
		return
	}
	connTag, _ := ctx.Value(&met.connectionTag).(*stats.ConnTagInfo)
	if connTag == nil {
		return
	}
	var connBegin bool
	switch stat.(type) {
	case *stats.ConnBegin:
		connBegin = true
	case *stats.ConnEnd:
	default:
		return
	}
	labs := prometheus.Labels{
		LabelLocalAddr: fmt.Sprintf("%s://%s", connTag.LocalAddr.Network(), connTag.LocalAddr.String()),
	}
	g := met.With(labs)
	if connBegin {
		g.Inc()
	} else {
		g.Dec()
	}
}
