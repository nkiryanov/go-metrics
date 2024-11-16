package db

import (
	"context"
	"embed"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

//go:embed migrations/*.sql
var fs embed.FS

// Initialize pg connection poll and run migrations
func New(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	// Produce migrations with go-migrate
	// Use embedded migrations, look at example at https://github.com/golang-migrate/migrate/blob/v4.18.1/source/iofs/example_test.go
	{
		driver, err := iofs.New(fs, "migrations")
		if err != nil {
			logger.Slog.Error(err)
			return nil, err
		}

		// It's stupid, but go-migrate doesn't understand postgres:// url if pgx driver uses
		m, err := migrate.NewWithSourceInstance("iofs", driver, strings.Replace(connString, "postgres://", "pgx5://", 1))
		if err != nil {
			logger.Slog.Error(err)
			return nil, err
		}

		// Run migrations
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Slog.Error(err)
			return nil, err
		}
	}

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		logger.Slog.Error("Can't connect to db %s", err)
		return nil, err
	}

	return dbpool, err
}
