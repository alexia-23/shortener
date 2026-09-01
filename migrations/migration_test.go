package migrations

import (
	"io/fs"
	"testing"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedMigrationFiles(t *testing.T) {
	entries, err := fs.ReadDir(
		migrationFiles,
		".",
	)
	require.NoError(t, err)

	require.NotEmpty(
		t,
		entries,
		"no migration files were embedded",
	)

	for _, entry := range entries {
		t.Logf("embedded file: %q", entry.Name())

		if entry.IsDir() {
			continue
		}

		migration, err := source.DefaultParse(entry.Name())
		require.NoError(
			t,
			err,
			"golang-migrate cannot parse %q",
			entry.Name(),
		)

		t.Logf(
			"parsed migration: version=%d identifier=%q direction=%s",
			migration.Version,
			migration.Identifier,
			migration.Direction,
		)
	}
}
