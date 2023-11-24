package prometheus_metrics

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/H-BF/corlib/server/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/stats"
)

func newConnectionsCountMetric(options serverMetricsOptions) prometheus.Collector {
	labels := [...]string{
		LabelLocalAddr, LabelRemoteAddr,
	}
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "connections",
		Help:      "connection count at moment on a server",
	}, labels[:])
	return &connMetric{GaugeVec: vec}
}

type (
	connMetric struct {
		*prometheus.GaugeVec
		interceptors.StatsHandlerBase
		connectionTag int
	}
)

var _ stats.Handler = (*connMetric)(nil)

func (met *connMetric) TagConn(ctx context.Context, tagInfo *stats.ConnTagInfo) context.Context {
	const unk = "unknown"
	promLabels := make(prometheus.Labels)
	addrs := []net.Addr{tagInfo.LocalAddr, tagInfo.RemoteAddr}
	labs := []string{LabelLocalAddr, LabelRemoteAddr}
	for i := range addrs {
		if l := addrs[i]; l != nil {
			s := l.Network()
			if s == "" {
				s = unk
			}
			if u, e := url.Parse(fmt.Sprintf("%s://%s", s, l.String())); e == nil {
				promLabels[labs[i]] = u.Hostname()
			} else {
				promLabels[labs[i]] = unk
			}
		}
	}
	return context.WithValue(ctx, &met.connectionTag, met.With(promLabels))
}

// HandleConn ...
func (met *connMetric) HandleConn(ctx context.Context, stat stats.ConnStats) {
	if stat.IsClient() {
		return
	}
	data := ctx.Value(&met.connectionTag).(prometheus.Gauge)
	switch stat.(type) {
	case *stats.ConnBegin:
		data.Inc()
	case *stats.ConnEnd:
		data.Dec()
	}
}
