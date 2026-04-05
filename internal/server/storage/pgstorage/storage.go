package pgstorage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type PgStorage struct {
	db DBTX
}

func New(_ context.Context, db DBTX) *PgStorage {
	return &PgStorage{db: db}
}

func WithTx(tx pgx.Tx) *PgStorage {
	return &PgStorage{db: tx}
}

func (s *PgStorage) Close() error {
	return nil
}
