package db

import (
	"embed"
	"errors"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Run embedded migrations on dbURI
// Check the example at https://github.com/golang-migrate/migrate/blob/v4.18.1/source/iofs/example_test.go
func Migrate(dbURI string) error {
	var err error

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithSourceInstance(
		"iofs",
		source,
		strings.Replace(dbURI, "postgres://", "pgx5://", 1), // go-migrate doesn't understand postgres:// url if pgx driver uses
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
