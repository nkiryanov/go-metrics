package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
)


func Connect(ctx context.Context, dbURI string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, err
	}

	return dbpool, err
}
