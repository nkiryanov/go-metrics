package pgstorage

import (
	"context"
	"embed"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Ping(context.Context) error
	Close()
}

type PgStorage struct {
	db DBTX
}

// Create PgStorage
// Note: it embed migrations files, that would be run on initialization
func New(ctx context.Context, connString string) (*PgStorage, error) {
	driver, err := iofs.New(migrations, "db/migrations")
	if err != nil {
		logger.Slog.Error(err)
		return nil, err
	}

	m, err := migrate.NewWithSourceInstance("iofs", driver, strings.Replace(connString, "postgres://", "pgx5://", 1))
	if err != nil {
		logger.Slog.Error(err)
		return nil, err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Slog.Error(err)
		return nil, err
	}

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		logger.Slog.Error("Can't connect to db %s", err)
		return nil, err
	}

	return &PgStorage{db: dbpool}, nil
}

func (s *PgStorage) Close() error {
	s.db.Close()
	return nil
}
