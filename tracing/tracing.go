package tracing

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns the tracer
func Tracer() trace.Tracer {
	return otel.Tracer("notes",
		trace.WithInstrumentationVersion("1.0.0"),
	)
}
