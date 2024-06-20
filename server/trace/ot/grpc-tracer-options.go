package ot

// WithTracerProvider add tracer provider
func WithTracerProvider(tp ServerTracerProvider) GRPCTracerOption {
	return grpcTracerOption(func(t *GRPCTracer) {
		t.tracerProvider = tp
	})
}

var (
	_ = WithTracerProvider
)

type grpcTracerOption func(*GRPCTracer)

func (f grpcTracerOption) apply(t *GRPCTracer) {
	f(t)
}
