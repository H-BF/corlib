package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestFromContext_GlobalLogger(t *testing.T) {
	logger := FromContext(context.Background())

	assert.Equal(t, Global(), logger)
}

func TestFromContext_WithLogger(t *testing.T) {
	l := New(zapcore.DebugLevel)
	ctx := ToContext(context.Background(), l)

	assert.Equal(t, l, FromContext(ctx))
}

func TestLoggerWithSpanContext(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	aLogger := NewWithSink(zap.InfoLevel, io.MultiWriter(os.Stdout, buf))
	expo, err := stdouttrace.New()
	if !assert.NoError(t, err) {
		return
	}
	provider := sdkTrace.NewTracerProvider(
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
		sdkTrace.WithBatcher(expo))

	if !assert.NotNil(t, provider) {
		return
	}
	tracer := provider.Tracer("test")
	if !assert.NotNil(t, tracer) {
		return
	}
	ctx, span := tracer.Start(context.Background(), "test")
	ctx = ToContext(ctx, aLogger)
	if !assert.NotNil(t, span) {
		return
	}
	Infof(ctx, "check '%s' and '%s' are present here", traceID, spanID)
	_ = aLogger.Sync()
	var decoded map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &decoded)
	if !assert.NoError(t, err) {
		return
	}
	spanCtx := span.SpanContext()
	if !assert.Equal(t, spanCtx.TraceID().String(), decoded[traceID]) {
		return
	}
	assert.Equal(t, spanCtx.SpanID().String(), decoded[spanID])
}
