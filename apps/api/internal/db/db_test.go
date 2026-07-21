package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestDriverFor(t *testing.T) {
	// The scheme→driver mapping is tested as a pure function: a real libsql
	// connection needs a live Turso server, so Open() on a libsql:// URL can't be
	// exercised in a unit test (Ping would fail on the network, not the scheme).
	cases := map[string]string{
		"libsql://example.turso.io":        "libsql",
		"libsql://db.turso.io?authToken=x": "libsql",
		"http://localhost:8080/db":         "libsql",
		"https://example.com":              "libsql",
		"wss://example.com":                "libsql",
		"ws://example.com":                 "libsql",
		"dev.db":                           "sqlite",
		":memory:":                         "sqlite",
		"file:/tmp/x.db":                   "sqlite",
		"sqlite:/tmp/x.db":                 "sqlite",
		"/var/lib/app.db":                  "sqlite",
	}
	for url, want := range cases {
		t.Run(url, func(t *testing.T) {
			if got := driverFor(url); got != want {
				t.Fatalf("driverFor(%q) = %q, want %q", url, got, want)
			}
		})
	}
}

func TestOpenRejectsEmptyURL(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestOpenLocalFileWorks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	if err := d.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	ctx := context.Background()
	if err := Migrate(ctx, d); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := Migrate(ctx, d); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}
