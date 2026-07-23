package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/ingest"
)

// fixtureDir writes a minimal prerendered static tree to a temp dir and
// returns its path. Mirrors what `bun run build` produces under
// apps/web/.output/public/ once Phase 3 wires the build pipeline.
func fixtureDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"index.html":            "<!doctype html><title>Home</title>home",
		"about/index.html":      "<!doctype html><title>About</title>about",
		"404.html":              "<!doctype html><title>Not Found</title>missing",
		"favicon.ico":           "icon-bytes",
		"team/alice/index.html": "<!doctype html><title>Alice</title>alice",
		"sw.js":                 "/* kill switch */",
		"service-worker.js":     "/* kill switch */",
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
	e := newServer(fixtureDir(t), nil, nil, nil)
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
	e := newServer(fixtureDir(t), nil, nil, nil)
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

// Service worker scripts sit at fixed, unhashed URLs; they must be served
// Cache-Control: no-cache so a returning visitor's browser (and the CDN) always
// revalidates and the kill switches can be retired cleanly. Ordinary assets
// must be left untouched.
func TestServiceWorkerScriptsSentNoCache(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil)
	cases := []struct {
		name        string
		path        string
		wantNoCache bool
	}{
		{"sw.js", "/sw.js", true},
		{"service-worker.js", "/service-worker.js", true},
		{"ordinary asset", "/favicon.ico", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status: want 200, got %d", rec.Code)
			}
			cc := rec.Header().Get("Cache-Control")
			if gotNoCache := strings.Contains(cc, "no-cache"); gotNoCache != tc.wantNoCache {
				t.Fatalf("%s Cache-Control=%q: no-cache want=%v got=%v", tc.path, cc, tc.wantNoCache, gotNoCache)
			}
		})
	}
}

// Service worker scripts must carry no-cache even when the file is missing:
// caches can store 404s, and a negatively cached /sw.js would hide a later
// kill-switch deploy.
func TestMissingServiceWorkerStillNoCache(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{"index.html": "home", "404.html": "missing"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	e := newServer(dir, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: want 404, got %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "no-cache") {
		t.Fatalf("missing /sw.js: want Cache-Control no-cache, got %q", cc)
	}
}

// HEAD must work everywhere GET works (RFC 9110 §9.3.2). Curl -I, fly's
// healthcheck probes, and ad-hoc monitoring all use HEAD; without explicit
// support Echo returns 405 instead of mirroring the GET status.
func TestHeadMirrorsGet(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil)
	cases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"healthz", "/healthz", http.StatusNoContent},
		{"root index", "/", http.StatusOK},
		{"about index", "/about/", http.StatusOK},
		{"missing page", "/nonexistent/", http.StatusNotFound},
		{"missing api", "/api/missing", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodHead, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status: want %d got %d body=%q", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestSSGFallbackOnUnknownPageRoute(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil)
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
	e := newServer(fixtureDir(t), nil, nil, nil)
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
	e := newServer(fixtureDir(t), nil, nil, nil)
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
	e := newServer(dir, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/nope/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: want 404, got %d", rec.Code)
	}
}

func TestTrailingSlashCanonicalization(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil)
	cases := []struct {
		name     string
		method   string
		path     string
		wantCode int
		wantLoc  string
	}{
		{"extensionless page redirects", http.MethodGet, "/about", http.StatusMovedPermanently, "/about/"},
		{"nested page redirects", http.MethodGet, "/team/alice", http.StatusMovedPermanently, "/team/alice/"},
		{"slashed page served", http.MethodGet, "/about/", http.StatusOK, ""},
		{"root not redirected", http.MethodGet, "/", http.StatusOK, ""},
		{"query string preserved", http.MethodGet, "/about?x=1&y=2", http.StatusMovedPermanently, "/about/?x=1&y=2"},
		{"asset not redirected", http.MethodGet, "/favicon.ico", http.StatusOK, ""},
		{"unresolvable not redirected", http.MethodGet, "/nope", http.StatusNotFound, ""},
		{"head mirrors get redirect", http.MethodHead, "/about", http.StatusMovedPermanently, "/about/"},
		// A double-slash path must resolve to the on-origin canonical form, not
		// a protocol-relative Location that points off-site.
		{"double slash canonicalized on origin", http.MethodGet, "//about", http.StatusMovedPermanently, "/about/"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Fatalf("status: want %d, got %d", tc.wantCode, rec.Code)
			}
			if tc.wantLoc != "" {
				if got := rec.Header().Get("Location"); got != tc.wantLoc {
					t.Fatalf("Location: want %q, got %q", tc.wantLoc, got)
				}
			}
		})
	}
}

func TestIngestRouteWiredWhenEnabled(t *testing.T) {
	// A non-nil ingest handler mounts POST /api/ingest behind bearer auth; an
	// unauthenticated request reaches that auth (401), proving the route is wired
	// and not swallowed by the /api/* JSON-404 reservation.
	ing := ingest.New(ingest.Config{IngestToken: "t"}, nil, nil, nil)
	e := newServer(fixtureDir(t), nil, nil, ing)
	req := httptest.NewRequest(http.MethodPost, "/api/ingest", strings.NewReader("x"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST /api/ingest without token: want 401, got %d", rec.Code)
	}
}

func TestIngestRouteAbsentWhenDisabled(t *testing.T) {
	// With no ingest handler, POST /api/ingest falls through to the /api/* JSON
	// 404 reservation (the pipeline is disabled).
	e := newServer(fixtureDir(t), nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/ingest", strings.NewReader("x"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /api/ingest with pipeline disabled: want 404, got %d", rec.Code)
	}
}

// TestNewVerifierFromEnvCanonicalNamesOnly locks the PUBLIC_/VITE_ aliases out
// of the verifier. It used to accept three names for the client id while
// session.LoadConfig accepted a different set, so a deploy carrying only an
// alias brought the verifier up and left sign-in off: /api/me answered while
// /auth/login returned 404.
func TestNewVerifierFromEnvCanonicalNamesOnly(t *testing.T) {
	clear := func(t *testing.T) {
		t.Helper()
		for _, k := range []string{
			"WORKOS_CLIENT_ID", "WORKOS_API_HOSTNAME",
			"PUBLIC_WORKOS_CLIENT_ID", "PUBLIC_WORKOS_API_HOSTNAME",
			"VITE_WORKOS_CLIENT_ID", "VITE_WORKOS_API_HOSTNAME",
		} {
			t.Setenv(k, "")
		}
	}

	t.Run("canonical names build a verifier", func(t *testing.T) {
		clear(t)
		t.Setenv("WORKOS_CLIENT_ID", "client_canonical")
		t.Setenv("WORKOS_API_HOSTNAME", "api.workos.com")
		if newVerifierFromEnv() == nil {
			t.Fatal("newVerifierFromEnv() = nil with canonical names set")
		}
	})

	for _, prefix := range []string{"PUBLIC_", "VITE_"} {
		t.Run(prefix+"aliases are ignored", func(t *testing.T) {
			clear(t)
			t.Setenv(prefix+"WORKOS_CLIENT_ID", "client_alias")
			t.Setenv(prefix+"WORKOS_API_HOSTNAME", "api.workos.com")
			if v := newVerifierFromEnv(); v != nil {
				t.Fatalf("newVerifierFromEnv() = %v with only %s aliases set, want nil", v, prefix)
			}
		})
	}
}
