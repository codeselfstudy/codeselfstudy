// Package db opens the API's SQLite/libSQL database and applies the embedded
// goose migrations under migrations/. The Go side owns the schema.
//
// Driver choice by URL scheme: a libsql:// (or http(s)/ws(s)) DATABASE_URL uses
// the pure-Go Turso client (github.com/tursodatabase/libsql-client-go); anything
// else — a bare path, dev.db, :memory:, file:… — uses the pure-Go
// modernc.org/sqlite driver. Both are CGO-free, so the distroless runtime image
// stays static (CGO_ENABLED=0). The Turso auth token travels inside the URL as
// ?authToken=…; there is no separate token env var.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	// Database drivers, registered for their side effects.
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// Open returns a *sql.DB pointing at databaseURL, with the driver chosen by
// scheme (see driverFor). The URL forms supported:
//
//   - "" is rejected (a database URL is required).
//   - "libsql://…", "http(s)://…", "ws(s)://…" → remote Turso (libsql driver).
//   - ":memory:", "dev.db", any bare path, "sqlite:…", "file:…" → local SQLite.
//
// For the local SQLite backend the pool is capped at one connection so PRAGMAs
// stick and the atomic digest claim in internal/store stays serialized, and
// foreign keys are enabled. Turso serializes writes server-side and does not
// enforce foreign keys, so neither is applied to the libsql backend.
func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("db: DATABASE_URL is empty")
	}

	driver := driverFor(databaseURL)
	db, err := sql.Open(driver, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	if driver == "sqlite" {
		// SQLite is a single-writer embedded engine: serialize access through one
		// connection so PRAGMAs stick and concurrent writers never see SQLITE_BUSY
		// (the digest-claim in internal/store relies on this serialization). It
		// also keeps :memory: databases alive across queries — a fresh connection
		// would otherwise get an empty database. Foreign keys are a per-connection
		// PRAGMA, enabled here as local integrity belt-and-suspenders (the pipeline
		// always inserts an email before that email's deals).
		db.SetMaxOpenConns(1)
		if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("db: enable foreign keys: %w", err)
		}
	} else {
		// libsql/Turso over the HTTP (hranaV2) transport: each database/sql
		// connection is a server-side stream Turso may close when idle. Cap the
		// pool at one so this low-traffic pipeline reuses a single stream instead
		// of churning several; Turso serializes writes server-side, so a single
		// connection costs nothing here. Schema migration tolerates a closed stream
		// on its own — see Migrate, which uses goose's pooled, retryable legacy API
		// rather than the dedicated-connection Provider.
		db.SetMaxOpenConns(1)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return db, nil
}

// driverFor selects the database/sql driver by URL scheme: the libsql client for
// remote Turso URLs (libsql/http(s)/ws(s)), the pure-Go modernc sqlite driver for
// everything else (bare paths, :memory:, file:/sqlite: URIs). Kept as a pure
// function so the scheme mapping is unit-testable without opening a connection —
// a real libsql connection needs a live Turso server, which CI does not have.
func driverFor(databaseURL string) string {
	switch {
	case strings.HasPrefix(databaseURL, "libsql://"),
		strings.HasPrefix(databaseURL, "http://"),
		strings.HasPrefix(databaseURL, "https://"),
		strings.HasPrefix(databaseURL, "wss://"),
		strings.HasPrefix(databaseURL, "ws://"):
		return "libsql"
	default:
		return "sqlite"
	}
}

// IsRemote reports whether databaseURL points at a remote libsql/Turso server
// rather than a local SQLite file or :memory:. It uses the same scheme test as
// driverFor. Callers use it to decide where migrations run: on startup for a
// local database (dev ergonomics), or out of band via the migrate CLI /
// `server -migrate` release step for a remote one, so a migration failure
// aborts a deploy instead of crash-looping the serving process.
func IsRemote(databaseURL string) bool {
	return driverFor(databaseURL) == "libsql"
}
