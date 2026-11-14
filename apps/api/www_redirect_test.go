package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The Go server owns the canonical host: any request whose Host carries the
// "www." prefix is 308-redirected to the same scheme, path and query on the
// bare apex host. It's registered as Pre middleware, so it fires before routing
// for every path and method. Production sits behind Fly's force_https proxy,
// which sets X-Forwarded-Proto: https and delivers an origin-form
// request-target — the tests mirror that (origin-form path + explicit Host +
// XFP header) so req.RequestURI and c.Scheme() match what the app really sees.
// Since echo v4.16.0, c.Scheme() only trusts X-Forwarded-Proto from a private,
// loopback or link-local peer, so the tests also set RemoteAddr to a private
// address, as Fly's proxy connects over its private network.
// flyProxyAddr stands in for Fly's proxy, which reaches the app from a private
// network address. httptest.NewRequest defaults to 192.0.2.1, a public address.
const flyProxyAddr = "172.16.0.1:1234"

func TestNonWWWRedirect(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil, nil)
	cases := []struct {
		name    string
		method  string
		path    string
		wantLoc string
	}{
		{"root", http.MethodGet, "/", "https://codeselfstudy.com/"},
		{"path preserved", http.MethodGet, "/events/", "https://codeselfstudy.com/events/"},
		{"query preserved", http.MethodGet, "/x/?a=1&b=2", "https://codeselfstudy.com/x/?a=1&b=2"},
		// The host redirect is independent of the legacy /book -> /learn/ map
		// (which only applies once a request reaches the apex): www./book
		// redirects to the apex /book verbatim, a single host hop.
		{"exact path not legacy-remapped", http.MethodGet, "/book", "https://codeselfstudy.com/book"},
		{"head mirrors get", http.MethodHead, "/", "https://codeselfstudy.com/"},
		// 308 keeps the method, so a POST to www is told to re-POST to the apex.
		{"method preserved", http.MethodPost, "/api/ingest", "https://codeselfstudy.com/api/ingest"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Host = "www.codeselfstudy.com"
			req.Header.Set("X-Forwarded-Proto", "https") // Fly sets this behind force_https
			req.RemoteAddr = flyProxyAddr
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

// A request already on the apex host is served normally — no redirect.
func TestApexHostNotRedirected(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "codeselfstudy.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.RemoteAddr = flyProxyAddr
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", rec.Code)
	}
}

// The redirect target's scheme comes from c.Scheme(), i.e. X-Forwarded-Proto.
// Without that header (bare dev over http) the target is http, not https —
// confirming the header is what selects https in production behind the proxy.
func TestNonWWWRedirectSchemeFromForwardedProto(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "www.codeselfstudy.com"
	// no X-Forwarded-Proto set
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusPermanentRedirect {
		t.Fatalf("status: want 308, got %d", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "http://codeselfstudy.com/" {
		t.Fatalf("Location: want %q, got %q", "http://codeselfstudy.com/", got)
	}
}

// A client on a public address can't pick the redirect scheme by sending its
// own X-Forwarded-Proto: echo v4.16.0+ ignores the header from untrusted peers
// (GHSA-2ffq-g2xg-c22p), so the redirect falls back to the connection's scheme.
func TestNonWWWRedirectIgnoresSpoofedProto(t *testing.T) {
	e := newServer(fixtureDir(t), nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "www.codeselfstudy.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.RemoteAddr = "203.0.113.7:1234" // public (TEST-NET-3) address
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusPermanentRedirect {
		t.Fatalf("status: want 308, got %d", rec.Code)
	}
	if got, want := rec.Header().Get("Location"), "http://codeselfstudy.com/"; got != want {
		t.Fatalf("Location: want %q, got %q", want, got)
	}
}
