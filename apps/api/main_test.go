package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// fixtureDir writes a minimal prerendered static tree to a temp dir and
// returns its path. Mirrors what `bun run build` produces under
// apps/web/.output/public/ once Phase 3 wires the build pipeline.
func fixtureDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"index.html":       "<!doctype html><title>Home</title>home",
		"about/index.html": "<!doctype html><title>About</title>about",
		"404.html":         "<!doctype html><title>Not Found</title>missing",
	}
	for rel, body := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return dir
}

func TestHealthzReturns204(t *testing.T) {
	e := newServer(fixtureDir(t), nil)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: want 204, got %d", rec.Code)
	}
	if got := rec.Body.Len(); got != 0 {
		t.Fatalf("body: want empty, got %d bytes", got)
	}
}

func TestStaticFileServed(t *testing.T) {
	e := newServer(fixtureDir(t), nil)
	cases := []struct {
		name string
		path string
		body string
	}{
		{"root index", "/", "home"},
		{"about index", "/about/", "about"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status: want 200, got %d", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tc.body) {
				t.Fatalf("body: want to contain %q, got %q", tc.body, rec.Body.String())
			}
		})
	}
}

func TestSSGFallbackOnUnknownPageRoute(t *testing.T) {
	e := newServer(fixtureDir(t), nil)
	req := httptest.NewRequest(http.MethodGet, "/nonexistent/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: want 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "missing") {
		t.Fatalf("body: want 404 fixture, got %q", rec.Body.String())
	}
	if ct := rec.Header().Get(echo.HeaderContentType); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("content-type: want text/html, got %q", ct)
	}
}

func TestApiAndWsRoutesSkipSSGFallback(t *testing.T) {
	e := newServer(fixtureDir(t), nil)
	cases := []string{"/api/missing", "/ws"}
	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("status: want 404, got %d", rec.Code)
			}
			if strings.Contains(rec.Body.String(), "missing</title>") {
				t.Fatalf("api/ws routes must not get the SSG 404 page; body=%q", rec.Body.String())
			}
		})
	}
}

// Echo's gzip middleware wraps the ResponseWriter, which breaks the
// connection hijack a WebSocket upgrade needs. The middleware in newServer
// skips /ws so Phase 4's hub keeps working.
func TestGzipSkippedOnWebsocketPath(t *testing.T) {
	e := newServer(fixtureDir(t), nil)
	cases := []struct {
		path     string
		wantGzip bool
	}{
		{path: "/", wantGzip: true},
		{path: "/ws", wantGzip: false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			gotGzip := rec.Header().Get(echo.HeaderContentEncoding) == "gzip"
			if gotGzip != tc.wantGzip {
				t.Fatalf("path %s: gzip want=%v got=%v (status %d)", tc.path, tc.wantGzip, gotGzip, rec.Code)
			}
		})
	}
}

func TestSSGFallbackHandlesMissing404Html(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("home"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	e := newServer(dir, nil)

	req := httptest.NewRequest(http.MethodGet, "/nope/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: want 404, got %d", rec.Code)
	}
}
