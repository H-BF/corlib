package ot

import (
	"context"

	sdkRes "go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type (
	//SpanProcessorRegistrar reg/unregister of span processors
	SpanProcessorRegistrar interface {
		RegisterSpanProcessor(sdkTrace.SpanProcessor)
		UnregisterSpanProcessor(sdkTrace.SpanProcessor)
	}

	//ExplicitShutdown ...
	ExplicitShutdown interface {
		Shutdown(ctx context.Context) error
	}

	//TracerProvider ...
	TracerProvider interface {
		trace.TracerProvider
		SpanProcessorRegistrar
		ExplicitShutdown
	}

	//TraceProviderDeps TracerProvider deps
	TraceProviderDeps struct {
		Sampler        sdkTrace.Sampler
		SpanProcessors []sdkTrace.SpanProcessor //hide exporters there
		Resource       *sdkRes.Resource         //optional
		//Exporter     sdkTrace.SpanExporter
	}
)

// NewAppTraceProvider creates trace.TracerProvider instance
func NewAppTraceProvider(_ context.Context, deps TraceProviderDeps) TracerProvider {
	if deps.Sampler == nil {
		deps.Sampler = sdkTrace.AlwaysSample()
	}
	opts := []sdkTrace.TracerProviderOption{
		sdkTrace.WithSampler(deps.Sampler),
		//sdkTrace.WithSpanProcessor(sdkTrace.NewBatchSpanProcessor(deps.Exporter)),
	}
	if deps.Resource != nil {
		opts = append(opts, sdkTrace.WithResource(deps.Resource))
	}
	for _, sp := range deps.SpanProcessors {
		opts = append(opts, sdkTrace.WithSpanProcessor(sp))
	}
	return sdkTrace.NewTracerProvider(opts...)
}
