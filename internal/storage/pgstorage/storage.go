package pgstorage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type PgStorage struct {
	db DBTX
}

func New(ctx context.Context, db DBTX) *PgStorage {
	return &PgStorage{db: db}
}

func WithTx(tx pgx.Tx) *PgStorage {
	return &PgStorage{db: tx}
}

func (s *PgStorage) Close() error {
	return nil
}
