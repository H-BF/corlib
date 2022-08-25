//nolint:revive
package ot

import (
	_ "go.opentelemetry.io/otel"
	_ "go.opentelemetry.io/otel/exporters/jaeger"
	_ "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	_ "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	_ "go.opentelemetry.io/otel/sdk/trace"
)
