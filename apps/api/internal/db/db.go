// Package db opens a SQLite database for the API and applies the embedded
// goose migrations under migrations/. The Go side owns the schema.
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
	// SQLite is a single-writer embedded engine: serialize access through one
	// connection so PRAGMAs stick and concurrent writers never see SQLITE_BUSY
	// (the digest-claim in internal/store relies on this serialization). It also
	// keeps :memory: databases alive across queries — a fresh connection would
	// otherwise get an empty database. Foreign keys are a per-connection PRAGMA,
	// enabled here as local integrity belt-and-suspenders (the pipeline always
	// inserts an email before that email's deals); Turso does not enforce them.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db: enable foreign keys: %w", err)
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
