// Package resolve turns newsletter deal links into clean canonical URLs. Many
// newsletters wrap deal links in a click-tracking redirector, so the extracted
// URL is the tracker, not the deal page; a quick fetch that follows the
// redirect chain finds the real destination, and known tracking parameters
// (utm_*, mcID, linkID, …) are stripped from it. Some trackers (HubSpot among
// them) end the HTTP chain on a 200 interstitial that redirects via
// <meta http-equiv="refresh"> or an inline script instead — those body-level
// redirects are followed too, under their own hop cap. Resolution is strictly
// best-effort: on any failure the input URL is returned unchanged, so a
// resolver outage can never lose or corrupt a deal.
//
// The URLs being fetched come from untrusted email content, so the resolver is
// hardened against SSRF: only http(s) is fetched, and every connection —
// including each redirect hop, after DNS resolution — refuses private,
// loopback, link-local, and otherwise non-public addresses (the server runs
// inside a Fly private network).
package resolve

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"
)

const (
	// maxRedirects bounds the redirect chain. Real trackers use one or two
	// hops; a longer chain is a loop or abuse.
	maxRedirects = 5

	// fetchTimeout bounds one Resolve call end to end (dial, TLS, redirects).
	// Deals are resolved inline during ingest, so slow hosts must give up
	// quickly and fall back to the extracted URL.
	fetchTimeout = 4 * time.Second

	// userAgent identifies the fetcher honestly; some hosts reject Go's
	// default agent outright.
	userAgent = "codeselfstudy-deals/1.0 (+https://codeselfstudy.com)"

	// maxBodyBytes caps how much of a deal page ResolvePage reads. Structured
	// data (JSON-LD) sits in the <head> or early <body>, so half a megabyte is
	// plenty and bounds memory per fetch.
	maxBodyBytes = 512 << 10

	// maxBodyHops bounds how many body-level redirects (meta refresh, script
	// location assignment) one resolution follows after the initial fetch. Real
	// trackers use a single interstitial; more is a loop.
	maxBodyHops = 3

	// jsSniffLimit is the largest body a script-based redirect is honored on.
	// Tracker interstitials are tiny; a full content page may well contain an
	// incidental, conditional location assignment that must not be treated as
	// a redirect. Meta refresh is exempt — it is an HTML-standard redirect
	// regardless of page size.
	jsSniffLimit = 32 << 10

	// maxRefreshDelay is the largest <meta refresh> delay, in seconds, still
	// treated as a redirect. Trackers use 0; a long delay is a content page
	// reloading itself, not a redirector.
	maxRefreshDelay = 5
)

var errTooManyRedirects = errors.New("too many redirects")

// Resolver resolves deal URLs. Construct with New; the zero value is not
// usable.
type Resolver struct {
	client *http.Client
	// timeout is fetchTimeout in production; tests shorten it.
	timeout time.Duration
}

// New returns a Resolver with the SSRF guard enabled.
func New() *Resolver { return newResolver(false) }

// newResolver optionally disables the non-public-address guard so tests can
// fetch from httptest servers on loopback.
func newResolver(allowPrivate bool) *Resolver {
	dialer := &net.Dialer{Timeout: fetchTimeout}
	if !allowPrivate {
		// Control runs after DNS resolution with the literal address being
		// dialed, so a tracker that redirects (or DNS-rebinds) to an internal
		// address is refused at the socket, on every hop.
		dialer.Control = rejectNonPublic
	}
	return &Resolver{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
				// One-shot fetches; don't hold idle connections to newsletter
				// trackers.
				DisableKeepAlives:   true,
				TLSHandshakeTimeout: fetchTimeout,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return errTooManyRedirects
				}
				if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
					return fmt.Errorf("refusing redirect to scheme %q", req.URL.Scheme)
				}
				return nil
			},
		},
		timeout: fetchTimeout,
	}
}

// Resolve follows rawURL's redirect chain and returns the final URL with known
// tracking parameters stripped. It always returns a usable URL: on any failure
// (non-http(s) input, timeout, refused address, too many redirects) it returns
// rawURL unchanged, with a non-nil error the caller may log. Tracking
// parameters are stripped only after a successful fetch — on an unresolved
// redirector the query often *is* the destination, so an unfetched URL is left
// exactly as extracted.
func (r *Resolver) Resolve(ctx context.Context, rawURL string) (string, error) {
	final, _, err := r.fetch(ctx, rawURL, false)
	return final, err
}

// ResolvePage is Resolve plus the destination page itself: when the final
// response is HTML and the destination answered 2xx, up to maxBodyBytes of the
// body is returned so the caller can mine it (e.g. for an expiration date). A
// nil body with a nil error means the page was deliberately skipped (non-HTML,
// or an error status); the returned URL is usable either way.
func (r *Resolver) ResolvePage(ctx context.Context, rawURL string) (string, []byte, error) {
	return r.fetch(ctx, rawURL, true)
}

// fetch implements Resolve/ResolvePage. It never returns an unusable URL: the
// first value is the cleaned final URL on success and rawURL unchanged on any
// failure. The body is always read on a 2xx HTML response — even when the
// caller wants only the URL — because it may carry a body-level redirect
// (meta refresh, script location assignment) that continues the chain.
func (r *Resolver) fetch(ctx context.Context, rawURL string, readBody bool) (string, []byte, error) {
	if strings.TrimSpace(rawURL) == "" {
		return rawURL, nil, nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, nil, fmt.Errorf("parse %q: %w", rawURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		// Not fetchable; deliberate skip rather than a failure.
		return rawURL, nil, nil
	}

	// One timeout spans the whole chain, body hops included.
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cur := u
	var reached *url.URL // destination of the last successful fetch
	var body []byte
	var bodyOK bool // body belongs to a 2xx HTML response
	for hop := 0; ; hop++ {
		resp, err := r.do(ctx, cur)
		if err != nil {
			if reached == nil {
				return rawURL, nil, err
			}
			// A body-mined hop failed mid-chain. The last page actually
			// reached is still a usable destination; its body is a tracker
			// interstitial, not a deal page, so it is not returned for
			// mining.
			return finish(reached, u), nil, err
		}

		body, bodyOK = nil, false
		if resp.StatusCode >= 200 && resp.StatusCode < 300 &&
			strings.Contains(resp.Header.Get("Content-Type"), "html") {
			// Best-effort read under the same timeout; a partial or failed
			// read just means less (or no) page to sniff and mine, not a
			// failed resolution.
			body, _ = io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
			bodyOK = true
		}
		resp.Body.Close()

		// resp.Request is the last request of the HTTP chain, i.e. the
		// destination — even when that page answers 404, its URL is still the
		// canonical link.
		reached = resp.Request.URL
		if !bodyOK || hop >= maxBodyHops {
			break
		}
		next := nextHop(body, reached)
		if next == nil {
			break
		}
		cur = next
	}

	final := finish(reached, u)
	if !readBody || !bodyOK {
		return final, nil, nil
	}
	return final, body, nil
}

// do performs one GET (following HTTP redirects) for a link in the chain.
func (r *Resolver) do(ctx context.Context, u *url.URL) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %q: %w", u.String(), err)
	}
	return resp, nil
}

// finish renders the reached destination as the cleaned final URL.
func finish(reached, orig *url.URL) string {
	final := *reached
	final.RawQuery = stripTracking(final.RawQuery)
	if orig.Fragment != "" && final.Fragment == "" {
		// Browsers carry the original fragment across redirects unless a
		// Location header supplies its own; do the same.
		final.Fragment = orig.Fragment
	}
	return final.String()
}

// jsRedirectPatterns match the unconditional location assignments tracker
// interstitials use: `window.location.href = "…"`, `location.replace('…')`,
// `document.location = "…"`, `top.location = '…'`, and the assign() form.
var jsRedirectPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:window\.|document\.|top\.)?location(?:\.href)?\s*=\s*["']([^"']+)["']`),
	regexp.MustCompile(`(?i)(?:window\.|document\.|top\.)?location\.(?:replace|assign)\(\s*["']([^"']+)["']\s*\)`),
}

// nextHop extracts a body-level redirect target from an HTML page: a
// <meta http-equiv="refresh"> URL with a small delay or — on small pages only
// (jsSniffLimit) — an inline-script location assignment. It returns nil when
// the page carries no such redirect, the target does not resolve to http(s),
// or the target is the page itself (a self-refresh, not a redirect).
func nextHop(body []byte, base *url.URL) *url.URL {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	sniffScripts := len(body) <= jsSniffLimit
	var target string
	var scripts []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if target != "" {
			return
		}
		if n.Type == html.ElementNode {
			switch n.Data {
			case "meta":
				if strings.EqualFold(attr(n, "http-equiv"), "refresh") {
					if u := parseRefreshContent(attr(n, "content")); u != "" {
						target = u
						return
					}
				}
			case "script":
				if sniffScripts && n.FirstChild != nil {
					scripts = append(scripts, n.FirstChild.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if target == "" {
		for _, src := range scripts {
			for _, re := range jsRedirectPatterns {
				if m := re.FindStringSubmatch(src); m != nil {
					target = m[1]
					break
				}
			}
			if target != "" {
				break
			}
		}
	}
	if target == "" {
		return nil
	}

	next, err := base.Parse(target)
	if err != nil {
		return nil
	}
	if next.Scheme != "http" && next.Scheme != "https" {
		return nil
	}
	if next.String() == base.String() {
		return nil
	}
	return next
}

// parseRefreshContent pulls the URL out of a meta-refresh content attribute
// ("0;url=https://…", "1; URL='/path'"). It returns "" when there is no URL
// part (a bare delay reloads the same page) or the delay exceeds
// maxRefreshDelay.
func parseRefreshContent(content string) string {
	delayPart, rest, ok := strings.Cut(content, ";")
	if !ok {
		return ""
	}
	delayPart = strings.TrimSpace(delayPart)
	if delayPart != "" {
		delay, err := strconv.ParseFloat(delayPart, 64)
		if err != nil || delay > maxRefreshDelay {
			return ""
		}
	}
	rest = strings.TrimSpace(rest)
	i := strings.Index(strings.ToLower(rest), "url=")
	if i < 0 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(rest[i+len("url="):]), `'"`)
}

// attr returns the named attribute's value, matching the name
// case-insensitively.
func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val
		}
	}
	return ""
}

// stripTracking removes known tracking parameters from a raw query string. It
// works on the raw string (split on "&") rather than url.Values so the
// surviving parameters keep their original order and encoding; only the key's
// *classification* uses the decoded form, so a percent-encoded tracking key
// (%75tm_source) can't sneak past the check.
func stripTracking(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	pairs := strings.Split(rawQuery, "&")
	kept := pairs[:0]
	for _, p := range pairs {
		key := p
		if i := strings.IndexByte(p, '='); i >= 0 {
			key = p[:i]
		}
		if dec, err := url.QueryUnescape(key); err == nil {
			key = dec
		}
		if !isTrackingParam(key) {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, "&")
}

// isTrackingParam reports whether a query key is a known tracking parameter.
// Matching is case-insensitive (newsletters write mcID, mcid, …). The list is
// deliberately conservative: an unknown parameter might be load-bearing, and
// the digest's display-time strip removes leftovers from the Slack link anyway.
func isTrackingParam(key string) bool {
	k := strings.ToLower(key)
	if strings.HasPrefix(k, "utm_") {
		return true
	}
	switch k {
	case "mcid", "linkid", "fbclid", "gclid", "msclkid", "mc_cid", "mc_eid", "igshid", "twclid":
		return true
	}
	return false
}

// cgnat is the RFC 6598 carrier-grade-NAT shared range. It is non-public
// space some clouds use internally, yet neither IsGlobalUnicast nor IsPrivate
// covers it, so it needs its own deny.
var cgnat = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// rejectNonPublic is a net.Dialer Control hook that refuses any address that
// is not public global unicast: loopback, link-local (which covers cloud
// metadata endpoints), multicast, unspecified, RFC 1918 / ULA private ranges
// (which cover Fly's 6PN fdaa::/16), and the RFC 6598 CGNAT range.
func rejectNonPublic(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("resolve: refusing unparseable address %q: %w", address, err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("resolve: refusing non-IP address %q", host)
	}
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || cgnat.Contains(ip) {
		return fmt.Errorf("resolve: refusing non-public address %v", ip)
	}
	return nil
}
