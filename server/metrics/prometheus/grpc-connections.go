package prometheus_metrics

import (
	"context"
	"fmt"
	"sync"

	"github.com/H-BF/corlib/pkg/conventions"
	"github.com/H-BF/corlib/server/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/stats"
)

func newConnectionsCountMetric(options serverMetricsOptions) prometheus.Collector {
	labels := []string{LabelLocalAddr, LabelRemoteAddr, LabelUserAgent, LabelRemoteHostname}
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "connections",
		Help:      "connection count at moment on a server",
	}, labels)
	return &connMetric{GaugeVec: vec, connections: make(map[string]*connLabels)}
}

type (
	connMetric struct {
		sync.Mutex
		*prometheus.GaugeVec
		interceptors.StatsHandlerBase
		connectionTag int
		connections   map[string]*connLabels
	}

	connLabels struct {
		sync.Once
		labs prometheus.Labels
	}
)

var _ stats.Handler = (*connMetric)(nil)

func (met *connMetric) TagConn(ctx context.Context, tagInfo *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, &met.connectionTag, tagInfo)
}

// HandleConn ...
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

	met.Lock()
	defer met.Unlock()
	if connBegin {
		met.connections[connTag.RemoteAddr.String()] = &connLabels{}
	} else {
		if cl, ok := met.connections[connTag.RemoteAddr.String()]; ok {
			if len(cl.labs) != 0 {
				met.With(cl.labs).Dec()
			}
			delete(met.connections, connTag.RemoteAddr.String())
		}
	}
}

// HandleRPC ...
func (met *connMetric) HandleRPC(_ context.Context, stat stats.RPCStats) {
	if stat.IsClient() {
		return
	}

	if inHeader, ok := stat.(*stats.InHeader); ok {
		met.Lock()
		defer met.Unlock()
		cl, ok := met.connections[inHeader.RemoteAddr.String()]
		if !ok {
			return
		}
		cl.Do(func() {
			cl.labs = prometheus.Labels{
				LabelLocalAddr:      fmt.Sprintf("%s://%s", inHeader.LocalAddr.Network(), inHeader.LocalAddr.String()),
				LabelRemoteAddr:     inHeader.RemoteAddr.String(),
				LabelUserAgent:      "unknown",
				LabelRemoteHostname: "unknown",
			}
			for labelName, headerName := range label2header {
				if len(inHeader.Header[headerName]) > 0 {
					cl.labs[labelName] = inHeader.Header[headerName][0]
				}
			}
			met.With(cl.labs).Inc()
		})

	}
}

var label2header = map[string]string{
	LabelUserAgent:      conventions.UserAgentHeaderNoVer,
	LabelRemoteHostname: conventions.RemoteHostnameHeader,
}
