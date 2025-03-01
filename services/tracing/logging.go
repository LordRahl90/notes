package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otelLog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func setupLogs(ctx context.Context) (*otelLog.LoggerProvider, error) {
	logExporter, err := otlploghttp.New(ctx, otlploghttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	return otelLog.NewLoggerProvider(
		otelLog.WithResource(
			resource.NewWithAttributes("notes",
				attribute.String("service.name", "notes"),
				attribute.String("environment", os.Getenv("ENVIRONMENT")),
				attribute.String("app.version", "1.0.0")),
		),
		otelLog.WithProcessor(
			otelLog.NewBatchProcessor(logExporter),
		),
	), nil
}
