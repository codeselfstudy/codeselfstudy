package resolve

import (
	"context"
	"fmt"
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

func TestResolveMetaRefreshHop(t *testing.T) {
	// HubSpot-style wall: the HTTP chain ends on a 200 interstitial whose meta
	// refresh points at a second tracker hop, which 302s to the deal page. The
	// fragment from the input must survive to the final URL.
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/wall", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<html><head><meta http-equiv="refresh" content="0;url=/click2"></head></html>`))
	})
	mux.HandleFunc("/click2", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/deal?id=9&utm_source=news", http.StatusFound)
	})
	mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/wall?linkID=xyz#offer")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/deal?id=9#offer"; got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveJSWallHop(t *testing.T) {
	// Tracker interstitials redirect with a handful of script idioms; each must
	// hop, and tracking params on the destination must still be stripped.
	scripts := []string{
		`window.location.href = "%s/deal?id=5&utm_source=n";`,
		`location.replace('%s/deal?id=5&utm_source=n');`,
		`document.location = "%s/deal?id=5&utm_source=n";`,
	}
	for _, script := range scripts {
		mux := http.NewServeMux()
		srv := httptest.NewServer(mux)
		mux.HandleFunc("/wall", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><script>`+script+`</script></body></html>`, srv.URL)
		})
		mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r := newResolver(true)
		got, err := r.Resolve(context.Background(), srv.URL+"/wall")
		if err != nil {
			t.Errorf("script %q: Resolve: %v", script, err)
		}
		if want := srv.URL + "/deal?id=5"; got != want {
			t.Errorf("script %q: Resolve = %q, want %q", script, got, want)
		}
		srv.Close()
	}
}

func TestResolveJSVarWallHop(t *testing.T) {
	// The location assignment's right-hand side is a variable; the URL literal
	// it was assigned must be found and followed.
	scripts := []string{
		`var u = "%s/deal?id=5&utm_source=n"; window.location.replace(u);`,
		`var u = "%s/deal?id=5&utm_source=n"; document.location = u;`,
	}
	for _, script := range scripts {
		mux := http.NewServeMux()
		srv := httptest.NewServer(mux)
		mux.HandleFunc("/wall", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><script>`+script+`</script></body></html>`, srv.URL)
		})
		mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r := newResolver(true)
		got, err := r.Resolve(context.Background(), srv.URL+"/wall")
		if err != nil {
			t.Errorf("script %q: Resolve: %v", script, err)
		}
		if want := srv.URL + "/deal?id=5"; got != want {
			t.Errorf("script %q: Resolve = %q, want %q", script, got, want)
		}
		srv.Close()
	}
}

func TestResolveJSVarInHelperFunction(t *testing.T) {
	// HubSpot's shape: the URL literal lives inside a helper function, the
	// variable actually assigned to location is built at runtime. The literal
	// must still be found by identifier name.
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/wall", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head><script>
function getTargetURLWithState(state) {
  var targetURL = "%s/deal?id=7&utm_source=n";
  return targetURL.substring(0, targetURL.length - 2) + state;
}
var targetURL;
try { targetURL = getTargetURLWithState("1"); } catch(e) { targetURL = getTargetURLWithState("-1"); }
if (/Android/.test(window.navigator.userAgent)) {
  document.location = targetURL;
} else {
  window.location.replace(targetURL);
}
</script></head><body><p>You're being redirected</p><a href="%s/deal?id=7&utm_source=n">click here</a></body></html>`, srv.URL, srv.URL)
	})
	mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/wall")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/deal?id=7"; got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveAnchorFallbackAfterUnresolvableJS(t *testing.T) {
	// A script redirect through a variable whose URL is built at runtime (no
	// literal anywhere) must fall back to the page's single absolute anchor —
	// the interstitial's no-JS path.
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/wall", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><script>var u = base + path; window.location.replace(u);</script>
<a href="%s/deal?id=3&utm_source=n">click here</a></body></html>`, srv.URL)
	})
	mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/wall")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/deal?id=3"; got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveAnchorFallbackNeedsSingleAnchor(t *testing.T) {
	// Same unresolvable script redirect, but two anchors: ambiguous, no hop.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><script>var u = base + path; window.location.replace(u);</script>
<a href="https://a.example/x">one</a> <a href="https://b.example/y">two</a></body></html>`))
	}))
	defer srv.Close()

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/wall")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/wall"; got != want {
		t.Errorf("Resolve = %q, want %q (multiple anchors must not hop)", got, want)
	}
}

func TestResolveAnchorFallbackNeedsDetectedJSRedirect(t *testing.T) {
	// A small page with one anchor but no script redirect is content, not an
	// interstitial; it must not hop. Likewise a lone *relative* anchor even
	// with the redirect detected.
	pages := []string{
		`<html><body><a href="https://a.example/x">a link</a></body></html>`,
		`<html><body><script>var u = base + path; window.location.replace(u);</script><a href="/x">rel</a></body></html>`,
	}
	for _, page := range pages {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(page))
		}))

		r := newResolver(true)
		got, err := r.Resolve(context.Background(), srv.URL+"/page")
		if err != nil {
			t.Errorf("page %q: Resolve: %v", page, err)
		}
		if want := srv.URL + "/page"; got != want {
			t.Errorf("page %q: Resolve = %q, want %q (must not hop)", page, got, want)
		}
		srv.Close()
	}
}

func TestResolveMetaRefreshRelativeQuoted(t *testing.T) {
	// A quoted, uppercase URL= with a relative target must resolve against the
	// interstitial's own URL.
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/t/wall", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><meta http-equiv="Refresh" content="1; URL='/deal'"></head></html>`))
	})
	mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/t/wall")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/deal"; got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
}

func TestResolveMetaRefreshSlowDelayIgnored(t *testing.T) {
	// A long refresh delay is a page reloading itself, not a redirector.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><meta http-equiv="refresh" content="30;url=/elsewhere"></head></html>`))
	}))
	defer srv.Close()

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/page")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/page"; got != want {
		t.Errorf("Resolve = %q, want %q (slow refresh must not hop)", got, want)
	}
}

func TestResolveJSIgnoredOnLargeBody(t *testing.T) {
	// A big page with an incidental location assignment is content, not a
	// tracker interstitial; it must not hop.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><script>location.href = "/elsewhere";</script>`))
		w.Write([]byte(strings.Repeat("<p>content</p>", jsSniffLimit/10)))
		w.Write([]byte(`</body></html>`))
	}))
	defer srv.Close()

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/article")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/article"; got != want {
		t.Errorf("Resolve = %q, want %q (large page must not hop)", got, want)
	}
}

func TestResolveBodyHopLoopCapped(t *testing.T) {
	// Two interstitials refreshing to each other must terminate at the hop cap
	// with a usable URL, not spin. With maxBodyHops=3: a→b→a→b, done.
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	wall := func(next string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><head><meta http-equiv="refresh" content="0;url=%s"></head></html>`, next)
		}
	}
	mux.HandleFunc("/a", wall("/b"))
	mux.HandleFunc("/b", wall("/a"))

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/a")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/b"; got != want {
		t.Errorf("Resolve = %q, want %q (cap lands on the last fetched hop)", got, want)
	}
}

func TestResolveSelfRefreshStops(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head><meta http-equiv="refresh" content="0;url=%s"></head></html>`, r.URL.Path)
	}))
	defer srv.Close()

	r := newResolver(true)
	got, err := r.Resolve(context.Background(), srv.URL+"/live")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := srv.URL + "/live"; got != want {
		t.Errorf("Resolve = %q, want %q (self-refresh must not loop)", got, want)
	}
}

func TestResolveBodyHopSchemeRefused(t *testing.T) {
	// A body redirect to a non-http(s) scheme is not followed.
	for _, target := range []string{"javascript:alert(1)", "file:///etc/passwd"} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><head><meta http-equiv="refresh" content="0;url=%s"></head></html>`, target)
		}))

		r := newResolver(true)
		got, err := r.Resolve(context.Background(), srv.URL+"/wall")
		if err != nil {
			t.Errorf("target %q: Resolve: %v", target, err)
		}
		if want := srv.URL + "/wall"; got != want {
			t.Errorf("target %q: Resolve = %q, want %q", target, got, want)
		}
		srv.Close()
	}
}

func TestResolveBodyHopFetchFailureKeepsLastReached(t *testing.T) {
	// When a body-mined hop points at a dead host, the interstitial actually
	// reached is still the usable URL — cleaned — and ResolvePage must not hand
	// the interstitial body to the miner.
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	dead.Close()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/wall", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head><meta http-equiv="refresh" content="0;url=%s/gone"></head></html>`, dead.URL)
	})

	r := newResolver(true)
	got, body, err := r.ResolvePage(context.Background(), srv.URL+"/wall?utm_source=x")
	if err == nil {
		t.Fatal("expected an error from the dead hop")
	}
	if want := srv.URL + "/wall"; got != want {
		t.Errorf("ResolvePage url = %q, want %q (last reached, cleaned)", got, want)
	}
	if body != nil {
		t.Errorf("body = %q, want nil (interstitial must not be mined)", body)
	}
}

func TestResolvePageBodyAfterBodyHop(t *testing.T) {
	// ResolvePage must return the destination page's body, not the
	// interstitial's.
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/wall", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><meta http-equiv="refresh" content="0;url=/deal"></head></html>`))
	})
	mux.HandleFunc("/deal", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>the deal page</body></html>`))
	})

	r := newResolver(true)
	got, body, err := r.ResolvePage(context.Background(), srv.URL+"/wall")
	if err != nil {
		t.Fatalf("ResolvePage: %v", err)
	}
	if want := srv.URL + "/deal"; got != want {
		t.Errorf("ResolvePage url = %q, want %q", got, want)
	}
	if !strings.Contains(string(body), "the deal page") {
		t.Errorf("body = %q, want the destination page content", body)
	}
}

func TestParseRefreshContent(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"zero delay", "0;url=https://x.example/d", "https://x.example/d"},
		{"space and quotes", `1; URL='/deal'`, "/deal"},
		{"double quotes", `0; url="/deal"`, "/deal"},
		{"empty delay", ";url=/d", "/d"},
		{"too slow", "30;url=/d", ""},
		{"bare delay", "5", ""},
		{"no url key", "0;something=/d", ""},
		{"unparseable delay", "soon;url=/d", ""},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseRefreshContent(tc.in); got != tc.want {
				t.Errorf("parseRefreshContent(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
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
		{"hubspot email params stripped", "_hsenc=p2ANqtz&_hsmi=430&_hsfp=123&ecid=ACsprv&id=9", "id=9"},
		{"hubspot params case-insensitive", "_HsEnc=x&ECID=y&id=9", "id=9"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripTracking(tc.in); got != tc.want {
				t.Errorf("stripTracking(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
