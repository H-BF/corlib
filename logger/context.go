package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type contextKey struct{}

const (
	traceID = "trace_id"
	spanID  = "span_id"
)

// ToContext returns new context with specified sugared logger inside.
func ToContext(ctx context.Context, l TypeOfLogger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

//IsLevelEnabled is log level enabled
func IsLevelEnabled(ctx context.Context, lvl LogLevel) bool {
	if l, ok := ctx.Value(contextKey{}).(TypeOfLogger); ok {
		return l.Enabled(lvl)
	}
	return Level().Enabled(lvl)
}

// FromContext returns logger from context if set. Otherwise returns global `global` logger.
// In both cases returned logger is populated with `trace_id` & `span_id`.
func FromContext(ctx context.Context) TypeOfLogger {
	l, ok := ctx.Value(contextKey{}).(TypeOfLogger)
	if !ok {
		l = Global()
	}
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		return loggerWithSpanContext(l, spanCtx)
	}
	return l
}

func loggerWithSpanContext(l TypeOfLogger, sc trace.SpanContext) TypeOfLogger {
	return TypeOfLogger{
		LevelEnabler: l.LevelEnabler,
		SugaredLogger: l.Desugar().With(
			zap.Stringer(traceID, sc.TraceID()),
			zap.Stringer(spanID, sc.SpanID()),
		).Sugar(),
	}
}
