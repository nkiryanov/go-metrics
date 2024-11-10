package pgstorage

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type PgStorage struct {
	db DBTX
}

func New(ctx context.Context, connString string) *PgStorage {
	m, err := migrate.New("file://db/migrations", strings.Replace(connString, "postgres://", "pgx5://", 1))
	if err != nil {
		logger.Slog.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Slog.Fatal(err)
	}

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		logger.Slog.Fatal("Can't connect to db %s", err)
	}

	return &PgStorage{db: dbpool}
}

func (q *PgStorage) WithTx(tx pgx.Tx) *PgStorage {
	return &PgStorage{
		db: tx,
	}
}
