package db

import (
	"context"
	"embed"
	"errors"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Run embedded migrations on dbURI
// Check the example at https://github.com/golang-migrate/migrate/blob/v4.18.1/source/iofs/example_test.go
// dsn: database source name in format postgres://...
func Migrate(dsn string) error {
	var err error

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithSourceInstance(
		"iofs",
		source,
		strings.Replace(dsn, "postgres://", "pgx5://", 1), // go-migrate doesn't understand postgres:// url if pgx driver uses
	)
	if err != nil {
		return err
	}

	// Run migrations
	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}


// Connect to the DB
// dsn: database source name in format postgres://...
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return dbpool, err
}


// Do both: Connect and Migrate
// dsn: database source name in format postgres://...
func ConnectAndMigrate(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	err := Migrate(dsn)
	if err != nil {
		return nil, err
	}

	dbpool, err := Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return dbpool, err
}
