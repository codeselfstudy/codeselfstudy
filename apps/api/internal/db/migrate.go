package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate brings db up to the latest schema using the embedded goose
// migrations. Safe to call on every startup — goose tracks which versions
// have run in its own table and skips already-applied migrations.
func Migrate(ctx context.Context, db *sql.DB) error {
	// goose expects the FS to be rooted at the migrations directory; the
	// embed.FS rooted at the package keeps the "migrations/" prefix.
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("db: migrate scope: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, sub)
	if err != nil {
		return fmt.Errorf("db: migrate: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("db: migrate up: %w", err)
	}
	return nil
}
