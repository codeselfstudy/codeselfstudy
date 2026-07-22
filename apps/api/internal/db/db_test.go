package db

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestIsRemote(t *testing.T) {
	// Mirrors driverFor: remote libsql/Turso schemes vs local SQLite. Decides
	// whether migrations run on startup (local) or via the release step (remote).
	cases := map[string]bool{
		"libsql://example.turso.io":        true,
		"libsql://db.turso.io?authToken=x": true,
		"http://localhost:8080/db":         true,
		"https://example.com":              true,
		"wss://example.com":                true,
		"ws://example.com":                 true,
		"dev.db":                           false,
		":memory:":                         false,
		"file:/tmp/x.db":                   false,
		"sqlite:/tmp/x.db":                 false,
		"/var/lib/app.db":                  false,
	}
	for url, want := range cases {
		t.Run(url, func(t *testing.T) {
			if got := IsRemote(url); got != want {
				t.Fatalf("IsRemote(%q) = %v, want %v", url, got, want)
			}
		})
	}
}

// TestMigrateAgainstLibsql exercises the migration against a live libsql/Turso
// endpoint named by LIBSQL_TEST_URL. It is skipped when the var is unset so CI
// stays hermetic (mirroring the env-gated Gemini live test). Point it at a local
// server for a quick check:
//
//	turso dev --port 8080 --db-file /tmp/t.db &
//	LIBSQL_TEST_URL=http://127.0.0.1:8080 go test ./internal/db -run TestMigrateAgainstLibsql
//
// or at a throwaway Turso database (libsql://…?authToken=…) to verify the
// cloud-only failure this path guards against. Runs twice to prove idempotency.
// NOTE: local sqld does not reproduce the original Turso crash — only a real
// Turso endpoint does — so a green local run is necessary but not sufficient.
func TestMigrateAgainstLibsql(t *testing.T) {
	url := os.Getenv("LIBSQL_TEST_URL")
	if url == "" {
		t.Skip("set LIBSQL_TEST_URL (a `turso dev` http:// URL or a Turso libsql:// URL) to run")
	}
	if !IsRemote(url) {
		t.Fatalf("LIBSQL_TEST_URL %q is not a remote libsql URL", url)
	}
	d, err := Open(url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	ctx := context.Background()
	if err := Migrate(ctx, d); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := Migrate(ctx, d); err != nil {
		t.Fatalf("second migrate (idempotent): %v", err)
	}
}

func TestResolveURL(t *testing.T) {
	const tok = "tok123"
	cases := []struct {
		name, rawURL, token, want string
	}{
		{"bare libsql gets token", "libsql://db.turso.io", tok, "libsql://db.turso.io?authToken=tok123"},
		{"https gets token", "https://db.turso.io", tok, "https://db.turso.io?authToken=tok123"},
		{"embedded authToken wins", "libsql://db.turso.io?authToken=abc", tok, "libsql://db.turso.io?authToken=abc"},
		{"embedded auth_token wins", "libsql://db.turso.io?auth_token=abc", tok, "libsql://db.turso.io?auth_token=abc"},
		{"local sqlite path unchanged", "dev.db", tok, "dev.db"},
		{"memory unchanged", ":memory:", tok, ":memory:"},
		{"empty url unchanged", "", tok, ""},
		{"empty token unchanged", "libsql://db.turso.io", "", "libsql://db.turso.io"},
		{"embedded jwt wins", "libsql://db.turso.io?jwt=abc", tok, "libsql://db.turso.io?jwt=abc"},
		{"preserves other query params", "libsql://db.turso.io?foo=bar", tok, "libsql://db.turso.io?authToken=tok123&foo=bar"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveURL(tc.rawURL, tc.token); got != tc.want {
				t.Fatalf("ResolveURL(%q, %q) = %q, want %q", tc.rawURL, tc.token, got, tc.want)
			}
		})
	}
}

func TestRedactToken(t *testing.T) {
	const url = "libsql://db.turso.io?authToken=SUPERSECRET123"

	// authTokenOf pulls the injected token back out of a resolved URL, and is
	// empty for a URL that carries none.
	if tok := authTokenOf(url); tok != "SUPERSECRET123" {
		t.Fatalf("authTokenOf = %q, want SUPERSECRET123", tok)
	}
	if tok := authTokenOf("libsql://db.turso.io?authToken=SUPERSECRET123&mode=ro"); tok != "SUPERSECRET123" {
		t.Fatalf("authTokenOf with trailing param = %q, want SUPERSECRET123", tok)
	}
	if authTokenOf("file:./dev.db") != "" {
		t.Error("authTokenOf should be empty for a local file URL")
	}

	// A driver error embedding the URL must come out with the token redacted but
	// the useful cause intact.
	errStr := `db: ping: dial ` + url + `: connection refused`
	got := RedactToken(errStr, url)
	if strings.Contains(got, "SUPERSECRET123") {
		t.Errorf("RedactToken left the token: %q", got)
	}
	if !strings.Contains(got, "connection refused") {
		t.Errorf("RedactToken dropped the cause: %q", got)
	}

	// No token in the URL → nothing to redact (local sqlite path, empty URL).
	if RedactToken("plain error", "file:./dev.db") != "plain error" {
		t.Error("RedactToken with a tokenless URL should be a no-op")
	}
	if RedactToken("plain error", "") != "plain error" {
		t.Error("RedactToken with an empty URL should be a no-op")
	}
}
