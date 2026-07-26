package resolve

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResolveFollowsRedirectsAndStrips(t *testing.T) {
	// Tracker-style chain: /click → /hop → the deal page, whose final URL mixes
	// tracking params with a real one. The real param must survive.
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/click", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/hop", http.StatusFound)
	})
	mux.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/deal?id=9&utm_source=news&mcID=abc", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/click?linkID=xyz")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/deal?id=9"; got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveDirectURLStripsTracking(t *testing.T) {
	// No redirect: the final URL is the input, minus tracking params, with the
	// fragment kept.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/deal?utm_source=x&id=9&fbclid=f#offer")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/deal?id=9#offer"; got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveOnlyTrackingParamsDropsQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/deal?utm_source=x&utm_medium=email")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/deal"; got != want {
		t.Errorf("Resolve = %q, want %q (query should be gone entirely)", got, want)
	}
}

func TestResolveFinalErrorStatusStillResolves(t *testing.T) {
	// A 404 at the destination is still a successful resolution — the chain
	// ended there, and that URL is the canonical link.
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/click", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/gone?utm_source=x", http.StatusFound)
	})
	mux.HandleFunc("/gone", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/click")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/gone"; got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveFetchFailureKeepsOriginal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	srv.Close() // dead server: connection refused

	orig := srv.URL + "/deal?utm_source=x"
	r := newResolver(true)
	got, err := r.Resolve(context.Background(), orig)
	if err == nil {
		t.Fatal("expected an error from a dead server")
	}
	if got != orig {
		t.Errorf("Resolve = %q, want the original %q (params must NOT be stripped without a fetch)", got, orig)
	}
}

func TestResolveTimeoutKeepsOriginal(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release
	}))
	defer func() { close(release); srv.Close() }()

	r := newResolver(true)
	r.timeout = 50 * time.Millisecond
	orig := srv.URL + "/slow"
	got, err := r.Resolve(context.Background(), orig)
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if got != orig {
		t.Errorf("Resolve = %q, want original %q", got, orig)
	}
}

func TestResolveTooManyRedirectsKeepsOriginal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path+"x", http.StatusFound) // endless chain
	}))
	defer srv.Close()

	r := newResolver(true)
	orig := srv.URL + "/r"
	got, err := r.Resolve(context.Background(), orig)
	if err == nil || !strings.Contains(err.Error(), "too many redirects") {
		t.Fatalf("err = %v, want too-many-redirects", err)
	}
	if got != orig {
		t.Errorf("Resolve = %q, want original %q", got, orig)
	}
}

func TestResolveSkipsNonHTTPInput(t *testing.T) {
	r := newResolver(true)
	for _, in := range []string{"", "mailto:a@b.example", "ftp://x.example/f", "/relative/path"} {
		got, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Errorf("Resolve(%q) err = %v, want nil (deliberate skip)", in, err)
		}
		if got != in {
			t.Errorf("Resolve(%q) = %q, want unchanged", in, got)
		}
	}
}

func TestResolveUnparseableInputKeptWithError(t *testing.T) {
	r := newResolver(true)
	in := "http://x.example/%zz"
	got, err := r.Resolve(context.Background(), in)
	if err == nil {
		t.Fatal("expected a parse error")
	}
	if got != in {
		t.Errorf("Resolve = %q, want unchanged input", got)
	}
}

func TestResolvePageReturnsBodyBehindRedirect(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/click", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/deal?utm_source=x", http.StatusFound)
	})
	mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<html><body>the deal page</body></html>"))
	})

	r := newResolver(true)
	got, body, err := r.ResolvePage(context.Background(), srv.URL+"/click")
	if err != nil {
		t.Fatalf("ResolvePage: %v", err)
	}
	if want := srv.URL + "/deal"; got != want {
		t.Errorf("ResolvePage url = %q, want %q", got, want)
	}
	if !strings.Contains(string(body), "the deal page") {
		t.Errorf("body = %q, want the page content", body)
	}
}

func TestResolvePageCapsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		big := strings.Repeat("x", maxBodyBytes+1000)
		w.Write([]byte(big))
	}))
	defer srv.Close()

	r := newResolver(true)
	_, body, err := r.ResolvePage(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("ResolvePage: %v", err)
	}
	if len(body) != maxBodyBytes {
		t.Errorf("body length = %d, want capped at %d", len(body), maxBodyBytes)
	}
}

func TestResolvePageSkipsNonHTML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("%PDF-1.7"))
	}))
	defer srv.Close()

	r := newResolver(true)
	got, body, err := r.ResolvePage(context.Background(), srv.URL+"/file?utm_source=x")
	if err != nil {
		t.Fatalf("ResolvePage: %v", err)
	}
	if body != nil {
		t.Errorf("body = %q, want nil for non-HTML", body)
	}
	if want := srv.URL + "/file"; got != want {
		t.Errorf("url = %q, want %q (still resolved)", got, want)
	}
}

func TestResolvePageErrorStatusNoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html>not found</html>"))
	}))
	defer srv.Close()

	r := newResolver(true)
	got, body, err := r.ResolvePage(context.Background(), srv.URL+"/gone")
	if err != nil {
		t.Fatalf("ResolvePage: %v", err)
	}
	if body != nil {
		t.Errorf("body = %q, want nil for an error status", body)
	}
	if want := srv.URL + "/gone"; got != want {
		t.Errorf("url = %q, want %q", got, want)
	}
}

func TestResolvePageFetchFailureKeepsOriginal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	srv.Close()

	orig := srv.URL + "/deal"
	r := newResolver(true)
	got, body, err := r.ResolvePage(context.Background(), orig)
	if err == nil {
		t.Fatal("expected an error from a dead server")
	}
	if got != orig || body != nil {
		t.Errorf("ResolvePage = (%q, %q), want original URL and nil body", got, body)
	}
}

func TestResolverGuardBlocksLoopback(t *testing.T) {
	// The production constructor must refuse loopback at the dial layer — this
	// proves the guard is actually wired into the transport.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := New()
	orig := srv.URL + "/deal"
	got, err := r.Resolve(context.Background(), orig)
	if err == nil {
		t.Fatal("expected the SSRF guard to refuse a loopback address")
	}
	if got != orig {
		t.Errorf("Resolve = %q, want original %q", got, orig)
	}
}

func TestRejectNonPublic(t *testing.T) {
	cases := []struct {
		address string
		wantErr bool
	}{
		{"10.0.0.8:80", true},                               // RFC 1918
		{"192.168.1.1:443", true},                           // RFC 1918
		{"127.0.0.1:80", true},                              // loopback
		{"169.254.169.254:80", true},                        // link-local (cloud metadata)
		{"[::1]:443", true},                                 // v6 loopback
		{"[fdaa::3]:80", true},                              // ULA — Fly 6PN
		{"[fe80::1]:80", true},                              // v6 link-local
		{"0.0.0.0:80", true},                                // unspecified
		{"224.0.0.1:80", true},                              // multicast
		{"100.64.0.5:80", true},                             // CGNAT low end (RFC 6598)
		{"100.127.255.254:443", true},                       // CGNAT high end
		{"100.63.255.255:80", false},                        // just below CGNAT — public
		{"100.128.0.1:80", false},                           // just above CGNAT — public
		{"example.com:80", true},                            // Control sees IPs only; a name here is wrong
		{"93.184.216.34:443", false},                        // public v4
		{"[2606:2800:220:1:248:1893:25c8:1946]:443", false}, // public v6
	}
	for _, tc := range cases {
		err := rejectNonPublic("tcp", tc.address, nil)
		if (err != nil) != tc.wantErr {
			t.Errorf("rejectNonPublic(%q) err = %v, wantErr %v", tc.address, err, tc.wantErr)
		}
	}
}

func TestStripTracking(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"empty", "", ""},
		{"mixed order preserved", "b=2&utm_source=x&a=1", "b=2&a=1"},
		{"case-insensitive mcID", "mcID=102:abc&x=1", "x=1"},
		{"utm prefix any suffix", "utm_whatever=1&utm_campaign=c", ""},
		{"valueless key", "utm_source&id=9", "id=9"},
		{"unknown params kept verbatim", "linkID={$linkID}&q=a%20b", "q=a%20b"},
		{"not a tracking prefix", "utmx=1", "utmx=1"},
		{"percent-encoded tracking key stripped", "%75tm_source=x&id=9", "id=9"},
		{"percent-encoded kept param stays encoded", "%69d=9", "%69d=9"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripTracking(tc.in); got != tc.want {
				t.Errorf("stripTracking(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
