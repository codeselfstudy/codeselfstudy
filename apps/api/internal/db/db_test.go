package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenRejectsRemoteSchemes(t *testing.T) {
	cases := []string{
		"libsql://example.turso.io",
		"http://localhost:8080/db",
		"https://example.com",
		"wss://example.com",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			if _, err := Open(u); err == nil {
				t.Fatalf("expected error for %q, got nil", u)
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
