package notes

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"log"
	"log/slog"
	"notes/services/migrator"
	"os"
	"testing"

	"notes/services/entities"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

var db *sql.DB

func TestMain(m *testing.M) {
	code := 1

	dbase, err := setupDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := migrator.Migrate(context.TODO(), dbase, getDsn()); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		} else {
			slog.Info("database is already up to date", "error", err)
		}
	}
	defer func() {
		if err := dbase.Close(); err != nil {
			log.Fatal(err)
		}
		os.Exit(code)
	}()
	db = dbase
	code = m.Run()
}

func TestCreateNote(t *testing.T) {
	service := New(db)
	req := entities.NoteReq{
		UserID:  "test-user",
		Title:   "Test Note",
		Content: "This is a test note content.",
	}

	err := service.CreateNote(t.Context(), req)
	require.NoError(t, err)
}

func setupDatabase() (*sql.DB, error) {
	dsn := getDsn()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func getDsn() string {
	return "notes_user:p@ssword@tcp(localhost:3308)/notes?parseTime=true&timeout=5s"
}
