package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"notes/services/migrator"
	"os"
	"os/signal"
	"time"

	"notes/server"
	"notes/services/tracing"

	"github.com/getsentry/sentry-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	env := os.Getenv("ENVIRONMENT")
	if env == "" || env == "dev" {
		if err := godotenv.Load(); err != nil {
			log.Fatal(err)
		}
	}

	shutdown, err := tracing.SetupOtel(ctx)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			sentry.CaptureException(err)
			log.Fatal(err)
		}
	}()

	logger := slogmulti.Fanout(otelslog.NewHandler("notes"), slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(slog.New(logger))

	slog.InfoContext(ctx, "starting up see slog", "day", "today", "time",
		time.Now(), "item", uuid.NewString(), "content", `{"message": "hello world"}`)

	db, err := migrator.SetupDB(ctx, getDsn())
	if err != nil {
		slog.ErrorContext(ctx, "failed to setup db", "error", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close db", "error", err)
		}
	}()

	if err := migrator.Migrate(ctx, db, getDsn()); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}

		slog.InfoContext(ctx, "database is already up to date", "error", err)
	}

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

	slog.Info("shutdown complete")
}

func getDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))
}
