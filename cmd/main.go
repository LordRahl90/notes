package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"notes/server"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	otelLog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	traceProvider *sdkTrace.TracerProvider
	meterProvider *metric.MeterProvider
	logProvider   *otelLog.LoggerProvider
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := setupOtel(ctx); err != nil {
		log.Fatal(err)
	}

	logger := otelslog.NewLogger("notes")
	slog.SetDefault(logger)

	logger.InfoContext(ctx, "starting up")
	slog.InfoContext(ctx, "starting up see slog")

	svr := server.New()
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = ":80"
	} else {
		appPort = ":" + appPort
	}

	svrErr := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "starting server", "port", appPort)
		svrErr <- svr.Start(appPort)
	}()

	select {
	case err := <-svrErr:
		slog.ErrorContext(ctx, "an error occurred from the server", "error", err)
		return

	case <-ctx.Done():
		slog.Info("shutting down")
		stop()
	}

	shutDownCtx := context.Background()
	if traceProvider != nil {
		if err := traceProvider.Shutdown(shutDownCtx); err != nil {
			log.Fatal(err)
		}
	}

	if meterProvider != nil {
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}

	if logProvider != nil {
		if err := logProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}

	slog.Info("shutdown complete")
}

func setupOtel(ctx context.Context) error {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)
	slog.Info("text propagator configured")

	traceExporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		return err
	}
	traceProvider = sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(traceExporter),
		sdkTrace.WithResource(
			resource.NewWithAttributes("notes",
				attribute.String("service.name", "notes"),
				attribute.String("app.version", "1.0.0")),
		),
	)
	otel.SetTracerProvider(traceProvider)
	slog.Info("tracer provider configured")

	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return err
	}

	meterProvider = metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(
			resource.NewWithAttributes("notes",
				attribute.String("service.name", "notes"),
				attribute.String("app.version", "1.0.0")),
		),
	)
	otel.SetMeterProvider(meterProvider)
	slog.Info("meter provider configured")

	logExporter, err := otlploghttp.New(ctx, otlploghttp.WithInsecure())
	if err != nil {
		return err
	}
	logProvider = otelLog.NewLoggerProvider(
		otelLog.WithResource(
			resource.NewWithAttributes("notes",
				attribute.String("service.name", "notes"),
				attribute.String("app.version", "1.0.0")),
		),
		otelLog.WithProcessor(
			otelLog.NewBatchProcessor(logExporter),
		),
	)
	global.SetLoggerProvider(logProvider)
	slog.Info("logger provider configured")

	return runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
}
