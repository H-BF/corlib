package ot

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	jaegerExp "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// ErrUnknownExporterKind unknown exporter kind
var ErrUnknownExporterKind = errors.New("unknown exporter kind")

type (
	//ExporterKindOf selects what is exporter we need
	ExporterKindOf interface {
		exporterIs()
	}
	//JaegerCollector jaeger collector via http
	JaegerCollector struct {
		ExporterKindOf
		EndpointURL string
		UserInfo    func() (user string, password string)
	}
	//JaegerAgent jaeger exporter via jaeger agent sidecar
	JaegerAgent struct {
		ExporterKindOf
		Host string
		Port string
	}
	//Stdout simplest sink via io,Writer
	Stdout struct {
		ExporterKindOf
		io.Writer
	}
	//Noop no operation exporter
	Noop struct {
		ExporterKindOf
	}
)

// NewExporter makes new instance of trace.SpanExporter
func NewExporter(_ context.Context, kindOf ExporterKindOf) (trace.SpanExporter, error) {
	const api = "ot.NewExporter"
	var (
		ret trace.SpanExporter
		err error
	)
	switch conf := kindOf.(type) {
	case JaegerAgent:
		var opts []jaegerExp.AgentEndpointOption
		if len(conf.Host) > 0 {
			opts = append(opts, jaegerExp.WithAgentHost(conf.Host))
		}
		if len(conf.Port) > 0 {
			opts = append(opts, jaegerExp.WithAgentPort(conf.Port))
		}
		ret, err = jaegerExp.New(
			jaegerExp.WithAgentEndpoint(opts...),
		)
	case JaegerCollector:
		opts := []jaegerExp.CollectorEndpointOption{
			jaegerExp.WithEndpoint(conf.EndpointURL),
		}
		if conf.UserInfo != nil {
			u, p := conf.UserInfo()
			if len(u) > 0 {
				opts = append(opts, jaegerExp.WithUsername(u))
			}
			if len(p) > 0 {
				opts = append(opts, jaegerExp.WithPassword(p))
			}
		}
		httpClient := retryablehttp.NewClient()
		httpClient.Logger = nil
		httpClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
			_ = numTries
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			return nil, err
		}
		httpClient.RetryMax = 5
		httpClient.RetryWaitMin = time.Millisecond
		httpClient.RetryWaitMax = 10 * time.Millisecond
		opts = append(opts, jaegerExp.WithHTTPClient(httpClient.StandardClient()))
		ret, err = jaegerExp.New(
			jaegerExp.WithCollectorEndpoint(opts...),
		)
	case Stdout:
		ret, err = stdouttrace.New(stdouttrace.WithWriter(conf.Writer))
	case Noop:
		ret = new(tracetest.NoopExporter)
	default:
		err = ErrUnknownExporterKind
	}
	return ret, errors.Wrap(err, api)
}
