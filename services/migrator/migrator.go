package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	
	"notes/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func SetupDB(ctx context.Context, connectionString string) (*sql.DB, error) {
	slog.InfoContext(ctx, "Setting up database")
	db, err := sql.Open("mysql", connectionString)
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

func Migrate(ctx context.Context, db *sql.DB, connectionString string) error {
	slog.InfoContext(ctx, "Migrating database")
	// This is important to initialize the driver
	// this might be a bug in golang-migrate, but I'm not sure just yet
	_, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("cannot connect to db: %w", err)
	}
	slog.InfoContext(ctx, "driver initialized")

	dsn := fmt.Sprintf("mysql://%s", connectionString)

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
