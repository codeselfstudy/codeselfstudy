// Package auth verifies WorkOS-issued JWTs server-side. The web frontend
// uses @workos-inc/authkit-react for the user-facing flow; the Go backend
// only needs to confirm that an incoming access token was actually signed by
// WorkOS and isn't expired before it lets a request reach a protected route.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Verifier holds the cached JWKS and the expected issuer for token
// validation. Construct one per process via NewVerifier; the cache refreshes
// itself on a goroutine in the background.
type Verifier struct {
	jwksURL string
	issuer  string
	keys    *jwk.Cache
	refresh time.Duration
}

// NewVerifier builds a Verifier that pulls keys from the WorkOS JWKS endpoint
// for clientID and accepts tokens whose `iss` claim equals
// `https://<apiHostname>`. apiHostname is the same value the web app reads
// from PUBLIC_WORKOS_API_HOSTNAME (e.g. "api.workos.com"); clientID is
// PUBLIC_WORKOS_CLIENT_ID. The cache refreshes every 15 minutes by default —
// override with WithRefreshInterval before calling Start.
func NewVerifier(clientID, apiHostname string) (*Verifier, error) {
	if clientID == "" {
		return nil, errors.New("auth: clientID is empty")
	}
	if apiHostname == "" {
		return nil, errors.New("auth: apiHostname is empty")
	}
	host := strings.TrimPrefix(apiHostname, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimSuffix(host, "/")
	jwksURL, err := url.JoinPath("https://"+host, "/sso/jwks", clientID)
	if err != nil {
		return nil, fmt.Errorf("auth: build jwks url: %w", err)
	}
	return &Verifier{
		jwksURL: jwksURL,
		issuer:  "https://" + host,
		refresh: 15 * time.Minute,
	}, nil
}

// WithJWKSURL overrides the JWKS endpoint. Tests use this to point the
// verifier at an httptest server.
func (v *Verifier) WithJWKSURL(u string) *Verifier {
	v.jwksURL = u
	return v
}

// WithIssuer overrides the expected `iss` claim. Tests use this when the
// httptest server's URL is the issuer in test tokens.
func (v *Verifier) WithIssuer(iss string) *Verifier {
	v.issuer = iss
	return v
}

// WithRefreshInterval overrides the JWKS refresh cadence. Must be called
// before Start.
func (v *Verifier) WithRefreshInterval(d time.Duration) *Verifier {
	v.refresh = d
	return v
}

// Start primes the JWKS cache and registers the URL for periodic refresh.
// Call once at server startup; the returned context is the one to use for
// the cache's lifetime — cancel it on shutdown to stop refreshing.
func (v *Verifier) Start(ctx context.Context) error {
	cache := jwk.NewCache(ctx)
	if err := cache.Register(v.jwksURL, jwk.WithMinRefreshInterval(v.refresh)); err != nil {
		return fmt.Errorf("auth: register jwks: %w", err)
	}
	if _, err := cache.Refresh(ctx, v.jwksURL); err != nil {
		return fmt.Errorf("auth: prime jwks: %w", err)
	}
	v.keys = cache
	return nil
}

// keySet returns the current cached JWKS, refreshing on demand if stale.
func (v *Verifier) keySet(ctx context.Context) (jwk.Set, error) {
	if v.keys == nil {
		return nil, errors.New("auth: verifier not started")
	}
	return v.keys.Get(ctx, v.jwksURL)
}

// ParseToken validates a raw JWT against the cached JWKS, the expected issuer,
// and expiry, returning the parsed token. Used by the Bearer Middleware, where
// the token arrives unwrapped from an untrusted client and the issuer must be
// pinned.
func (v *Verifier) ParseToken(ctx context.Context, raw string) (jwt.Token, error) {
	return v.parse(ctx, raw, jwt.WithIssuer(v.issuer))
}

// ParseTokenNoIssuer validates signature and expiry but NOT the issuer. It is
// for the cookie-backed session, where the token has already been sealed into a
// first-party AES-256-GCM cookie (tamper-proof) before we ever re-read it — so
// signature + expiry are the meaningful checks, and pinning the exact issuer
// only adds fragility. WorkOS AuthKit access tokens use
// `iss = "https://api.workos.com/"` (trailing slash) or
// `.../user_management/<client_id>` depending on setup, none of which match the
// bare host the Bearer path pins; requiring an exact match here made every
// /api/me reject the session and clear the cookie.
func (v *Verifier) ParseTokenNoIssuer(ctx context.Context, raw string) (jwt.Token, error) {
	return v.parse(ctx, raw)
}

func (v *Verifier) parse(ctx context.Context, raw string, opts ...jwt.ParseOption) (jwt.Token, error) {
	keys, err := v.keySet(ctx)
	if err != nil {
		return nil, err
	}
	return jwt.Parse(
		[]byte(raw),
		append([]jwt.ParseOption{jwt.WithKeySet(keys), jwt.WithValidate(true)}, opts...)...,
	)
}
