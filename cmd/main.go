package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"notes/server"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	slogmulti "github.com/samber/slog-multi"
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

	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://2b4ca728309c454895d70a3274dd1d90@app.glitchtip.com/8011",
		//Dsn: "https://21907032037b4790a4ca161e0fec8689@app.glitchtip.com/1",
	})
	if err != nil {
		log.Fatal(err)
	}

	sentry.CaptureException(errors.New("sentry error handling"))
	sentry.Flush(time.Second * 3)

	if err := setupOtel(ctx); err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
	}

	logger := slogmulti.Fanout(otelslog.NewHandler("notes"), slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(slog.New(logger))

	slog.InfoContext(ctx, "starting up see slog", "day", "today", "time",
		time.Now(), "item", uuid.NewString(), "content", `{"message": "hello world"}`)

	svr := server.New()
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = ":80"
	} else {
		appPort = ":" + appPort
	}

	svrErr := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "starting server", "app_port", appPort)
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
				attribute.String("environment", os.Getenv("ENVIRONMENT")),
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
				attribute.String("environment", os.Getenv("ENVIRONMENT")),
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
				attribute.String("environment", os.Getenv("ENVIRONMENT")),
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
