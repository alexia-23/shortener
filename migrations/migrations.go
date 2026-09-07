package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationFiles embed.FS

func Up(db *sql.DB) error {
	sourceDriver, err := iofs.New(
		migrationFiles,
		".",
	)
	if err != nil {
		return fmt.Errorf(
			"create migration source: %w",
			err,
		)
	}

	databaseDriver, err := postgres.WithInstance(
		db,
		&postgres.Config{},
	)
	if err != nil {
		return fmt.Errorf(
			"create migration database driver: %w",
			err,
		)
	}

	migration, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"postgres",
		databaseDriver,
	)
	if err != nil {
		return fmt.Errorf(
			"create migration: %w",
			err,
		)
	}

	if err := migration.Up(); err != nil &&
		!errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf(
			"apply migrations: %w",
			err,
		)
	}

	return nil
}
