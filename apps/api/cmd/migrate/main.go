// Migrate is a thin CLI around goose that runs against DATABASE_URL using
// the embedded migrations. The server applies migrations on startup, so
// this command is only needed for explicit `up` / `down` / `status` runs
// in dev and for `create` to scaffold a new migration file.
//
// Usage:
//
//	go run ./cmd/migrate up
//	go run ./cmd/migrate status
//	go run ./cmd/migrate down
//	go run ./cmd/migrate create <name>   # scaffolds in apps/api/internal/db/migrations/
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/db"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]

	if cmd == "create" {
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		if err := scaffold(os.Args[2]); err != nil {
			log.Fatal(err)
		}
		return
	}

	dbURL := db.ResolveURL(os.Getenv("DATABASE_URL"), os.Getenv("TURSO_AUTH_TOKEN"))
	if dbURL == "" {
		log.Fatal("DATABASE_URL is empty")
	}
	conn, err := db.Open(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	switch cmd {
	case "up":
		if err := db.Migrate(context.Background(), conn); err != nil {
			log.Fatal(err)
		}
	case "down":
		// Drop one migration. Useful for dev only — never run against a
		// shared DB without thinking about it.
		if err := goose.SetDialect("sqlite3"); err != nil {
			log.Fatal(err)
		}
		if err := goose.Down(conn, migrationsDir()); err != nil {
			log.Fatal(err)
		}
	case "status":
		if err := goose.SetDialect("sqlite3"); err != nil {
			log.Fatal(err)
		}
		if err := goose.Status(conn, migrationsDir()); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

// scaffold writes a new empty migration to the on-disk migrations directory
// so it can be edited and committed. The embedded FS used at runtime picks
// it up next time the binary is rebuilt.
func scaffold(name string) error {
	dir := migrationsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	if err := goose.Create(nil, dir, name, "sql"); err != nil {
		return fmt.Errorf("goose create: %w", err)
	}
	return nil
}

// migrationsDir resolves the on-disk migrations directory relative to this
// source file, so `go run ./cmd/migrate create` works from the repo root or
// from apps/api/.
func migrationsDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "internal/db/migrations"
	}
	// cmd/migrate/main.go → apps/api/internal/db/migrations
	apiRoot := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(apiRoot, "internal", "db", "migrations")
}

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, "usage: migrate <up|down|status|create <name>>")
}
