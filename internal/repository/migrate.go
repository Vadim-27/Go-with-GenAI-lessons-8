package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	// Database driver (pgx v5) and file source for golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrate applies all pending SQL migrations from dir to the database. The
// schema is created only by these files, never by the ORM.
func Migrate(dsn, dir string) error {
	m, err := migrate.New("file://"+dir, pgxURL(dsn))
	if err != nil {
		return fmt.Errorf("init migrations: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// pgxURL rewrites a postgres:// URL to the scheme registered by the pgx v5
// driver of golang-migrate.
func pgxURL(dsn string) string {
	for _, prefix := range []string{"postgresql://", "postgres://"} {
		if strings.HasPrefix(dsn, prefix) {
			return "pgx5://" + strings.TrimPrefix(dsn, prefix)
		}
	}
	return dsn
}
