package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate brings db up to the latest schema using the embedded goose
// migrations. Safe to call on every startup — goose tracks which versions
// have run in its own table and skips already-applied migrations.
//
// It deliberately uses goose's legacy top-level API (SetBaseFS + UpContext)
// rather than the newer Provider, because the two behave very differently
// against Turso. The Provider holds ONE dedicated *sql.Conn for the whole run;
// Turso closes that connection's server-side stream when it goes idle over the
// HTTP (hranaV2) transport, and a dedicated conn gets no database/sql retry, so
// the Provider's first probe or DDL then dies with "sql: connection is already
// closed" — this is what crash-looped the deploy. The legacy API issues every
// statement on the pooled *sql.DB, so database/sql transparently retries
// driver.ErrBadConn on a fresh connection. The failure only reproduces against a
// real Turso database — local SQLite and a local libsql server keep the stream
// alive, so it never surfaces there (verified against a live Turso DB).
func Migrate(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("db: migrate dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("db: migrate up: %w", err)
	}
	return nil
}
