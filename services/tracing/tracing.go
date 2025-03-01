package tracing

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns the tracer
func Tracer() trace.Tracer {
	return otel.Tracer("notes",
		trace.WithInstrumentationVersion("1.0.0"),
	)
}

func setupTraceProvider(ctx context.Context) (*sdkTrace.TracerProvider, error) {
	traceExporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		return nil, err
	}
	return sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(traceExporter),
		sdkTrace.WithResource(
			resource.NewWithAttributes("notes",
				attribute.String("service.name", "notes"),
				attribute.String("environment", os.Getenv("ENVIRONMENT")),
				attribute.String("app.version", "1.0.0")),
		),
	), nil
}

func SetupOtel(ctx context.Context) (func(ctx context.Context) error, error) {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)
	slog.Info("text propagator configured")

	traceProvider, err := setupTraceProvider(ctx)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(traceProvider)
	slog.Info("tracer provider configured")

	meterProvider, err := setupMetrics(ctx)
	if err != nil {
		return nil, err
	}
	otel.SetMeterProvider(meterProvider)
	slog.Info("meter provider configured")

	logProvider, err := setupLogs(ctx)
	if err != nil {
		return nil, err
	}
	global.SetLoggerProvider(logProvider)
	slog.Info("logger provider configured")

	return func(ctx context.Context) error {
		if err := logProvider.Shutdown(ctx); err != nil {
			return err
		}

		if err := meterProvider.Shutdown(ctx); err != nil {
			return err
		}

		return traceProvider.Shutdown(ctx)
	}, runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
}
