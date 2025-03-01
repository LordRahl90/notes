package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"notes/migrations"
	"notes/server"
	"notes/services/tracing"

	"github.com/getsentry/sentry-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

	db, err := setupDB(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to setup db", "error", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close db", "error", err)
		}
	}()

	if err := migrateDatabase(ctx, db); err != nil {
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

func setupDB(ctx context.Context) (*sql.DB, error) {
	slog.InfoContext(ctx, "Setting up database")
	dsn := getDsn()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to db: %w", err)
	}
	slog.InfoContext(ctx, "Database connection opened")
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot ping db: %w", err)
	}
	slog.InfoContext(ctx, "all is well and good with db initialization")
	return db, nil
}

func migrateDatabase(ctx context.Context, db *sql.DB) error {
	slog.InfoContext(ctx, "Migrating database")
	// This is important to initialize the driver
	// this might be a bug in golang-migrate, but I'm not sure just yet
	_, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("cannot connect to db: %w", err)
	}
	slog.InfoContext(ctx, "driver initialized")

	dsn := fmt.Sprintf("mysql://%s", getDsn())

	source, err := iofs.New(migrations.Migrations, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return err
	}

	return m.Up()
}

func getDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))
}
