package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Ports the 13 cases from the old src/lib/redirects.test.ts. Targets carry the
// trailing slash the Go port adds; the matching semantics are unchanged.
func TestFindRedirect(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		wantTo  string
		wantHit bool
	}{
		{"exact match", "/book", "/learn/", true},
		{"exact match with trailing slash", "/book/", "/learn/", true},
		{"non-existent path", "/nonexistent", "", false},
		{"wildcard /blog/*", "/blog/some-article", "/learn/", true},
		{"wildcard /blog/* nested", "/blog/2024/01/article", "/learn/", true},
		{"wildcard /wiki/*", "/wiki/some-page", "/learn/", true},
		{"wildcard /b/*", "/b/some-post", "/learn/", true},
		{"exact preferred over wildcard (/wiki)", "/wiki", "/learn/", true},
		{"exact preferred over wildcard (/wiki/Main_Page)", "/wiki/Main_Page", "/learn/", true},
		{"wildcard fallback (/wiki/something-else)", "/wiki/something-else", "/learn/", true},
		{"exact match not wildcard (/b)", "/b", "/learn/", true},
		{"partial prefix does not match (/blogpost)", "/blogpost", "", false},
		{"exact trailing slash not wildcard (/wiki/)", "/wiki/", "/learn/", true},
		{"wildcard needs content after slash (/blog/)", "/blog/", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := findRedirect(tc.path)
			if ok != tc.wantHit {
				t.Fatalf("hit: want %v, got %v", tc.wantHit, ok)
			}
			if ok && got.to != tc.wantTo {
				t.Fatalf("to: want %q, got %q", tc.wantTo, got.to)
			}
			if ok && got.status != http.StatusPermanentRedirect {
				t.Fatalf("status: want 308, got %d", got.status)
			}
		})
	}
}

// The legacy map must run before static resolution: /index.html is both a real
// file in the static root and a mapped redirect to /.
func TestLegacyRedirectPrecedesStaticFile(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusPermanentRedirect {
		t.Fatalf("status: want 308, got %d", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/" {
		t.Fatalf("Location: want %q, got %q", "/", got)
	}
}

func TestLegacyRedirectServedOverHTTP(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil, nil)
	cases := []struct {
		name    string
		method  string
		path    string
		wantLoc string
	}{
		{"exact redirect", http.MethodGet, "/book", "/learn/"},
		{"query preserved", http.MethodGet, "/book?a=1&b=2", "/learn/?a=1&b=2"},
		{"wildcard redirect", http.MethodGet, "/blog/anything", "/learn/"},
		{"head mirrors get", http.MethodHead, "/book", "/learn/"},
		// The incoming path is slash-normalized before lookup, so a legacy path
		// that only existed in trailing-slash form must still redirect.
		{"trailing-slash-only legacy path", http.MethodGet, "/support-us/", "/"},
		{"trailing-slash-only to learn", http.MethodGet, "/programming-notes-wiki/", "/learn/"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusPermanentRedirect {
				t.Fatalf("status: want 308, got %d", rec.Code)
			}
			if got := rec.Header().Get("Location"); got != tc.wantLoc {
				t.Fatalf("Location: want %q, got %q", tc.wantLoc, got)
			}
		})
	}
}
