// Package db opens a SQLite database for the API. The schema is owned by
// apps/web/ via Drizzle; the Go side only reads and writes through plain
// SQL.
//
// Driver choice: modernc.org/sqlite is pure Go (no CGO) so the distroless
// runtime image stays static. Remote Turso (libsql://) is not handled here
// — it would need github.com/tursodatabase/libsql-client-go added behind a
// scheme switch. The current setup matches DATABASE_URL=dev.db (and similar
// local files) plus :memory: for tests.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// Open returns a *sql.DB pointing at databaseURL. The URL forms supported:
//
//   - "" / "dev.db" / any non-scheme value → opened as a local file.
//   - ":memory:" → in-memory database.
//   - "sqlite:..." / "file:..." → passed through to the driver verbatim.
//
// Returns an error for libsql:// or http(s):// URLs so the caller fails
// fast instead of silently degrading.
func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("db: DATABASE_URL is empty")
	}
	if rejected := unsupportedScheme(databaseURL); rejected != "" {
		return nil, fmt.Errorf("db: %s URLs are not supported by the Go backend yet; use a local sqlite file or :memory:", rejected)
	}

	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return db, nil
}

func unsupportedScheme(u string) string {
	for _, s := range []string{"libsql://", "http://", "https://", "wss://", "ws://"} {
		if strings.HasPrefix(u, s) {
			return strings.TrimSuffix(s, "://")
		}
	}
	return ""
}

// Schema is the SQL that mirrors apps/web/src/db/schema.ts. We keep it here
// (rather than running Drizzle migrations from Go) so tests can bootstrap an
// in-memory DB. The web workspace remains the source of truth for
// production migrations.
const Schema = `
CREATE TABLE IF NOT EXISTS todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    created_at INTEGER DEFAULT (unixepoch())
);
`

// ApplySchema runs Schema against db. Used by tests; production relies on
// Drizzle migrations from apps/web/.
func ApplySchema(db *sql.DB) error {
	if _, err := db.Exec(Schema); err != nil {
		return fmt.Errorf("db: apply schema: %w", err)
	}
	return nil
}
