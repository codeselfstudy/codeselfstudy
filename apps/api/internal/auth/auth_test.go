package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	testIssuer = "https://test.workos.local"
	testKID    = "test-key-1"
)

// jwksFixture is a fully fledged httptest server that exposes a signing key
// pair plus its public JWKS. Tests sign tokens with priv, hand them to the
// verifier, and assert what gets accepted.
type jwksFixture struct {
	server *httptest.Server
	priv   *rsa.PrivateKey
	calls  *int32
}

func newJWKSFixture(t *testing.T) *jwksFixture {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa generate: %v", err)
	}
	pub, err := jwk.FromRaw(&priv.PublicKey)
	if err != nil {
		t.Fatalf("jwk from raw: %v", err)
	}
	if err := pub.Set(jwk.KeyIDKey, testKID); err != nil {
		t.Fatalf("set kid: %v", err)
	}
	if err := pub.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		t.Fatalf("set alg: %v", err)
	}
	set := jwk.NewSet()
	if err := set.AddKey(pub); err != nil {
		t.Fatalf("add key: %v", err)
	}

	var calls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return &jwksFixture{server: srv, priv: priv, calls: &calls}
}

func (f *jwksFixture) jwksURL() string {
	return f.server.URL + "/jwks"
}

func (f *jwksFixture) signToken(t *testing.T, mutate func(jwt.Token)) string {
	t.Helper()
	tok, err := jwt.NewBuilder().
		Issuer(testIssuer).
		Subject("user_01HW0EXAMPLE").
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(5 * time.Minute)).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if mutate != nil {
		mutate(tok)
	}
	key, err := jwk.FromRaw(f.priv)
	if err != nil {
		t.Fatalf("priv jwk: %v", err)
	}
	if err := key.Set(jwk.KeyIDKey, testKID); err != nil {
		t.Fatalf("priv kid: %v", err)
	}
	if err := key.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		t.Fatalf("priv alg: %v", err)
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return string(signed)
}

func startedVerifier(t *testing.T, f *jwksFixture) *Verifier {
	t.Helper()
	v, err := NewVerifier("test_client", "test.workos.local")
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}
	v.WithJWKSURL(f.jwksURL()).WithIssuer(testIssuer)
	if err := v.Start(t.Context()); err != nil {
		t.Fatalf("start: %v", err)
	}
	return v
}

func TestNewVerifierBuildsExpectedURLs(t *testing.T) {
	v, err := NewVerifier("client_123", "api.workos.com")
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}
	if got, want := v.jwksURL, "https://api.workos.com/sso/jwks/client_123"; got != want {
		t.Errorf("jwksURL: got %q want %q", got, want)
	}
	if got, want := v.issuer, "https://api.workos.com"; got != want {
		t.Errorf("issuer: got %q want %q", got, want)
	}
}

func TestNewVerifierRejectsEmptyArgs(t *testing.T) {
	cases := []struct {
		name        string
		clientID    string
		apiHostname string
	}{
		{"empty client", "", "api.workos.com"},
		{"empty hostname", "client_123", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewVerifier(tc.clientID, tc.apiHostname); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestMiddlewareAcceptsValidToken(t *testing.T) {
	f := newJWKSFixture(t)
	v := startedVerifier(t, f)

	e := echo.New()
	e.GET("/me", func(c echo.Context) error {
		tok := Claims(c)
		if tok == nil {
			t.Fatal("claims missing on accepted request")
		}
		if got, _ := tok.Get("sub"); got != "user_01HW0EXAMPLE" {
			t.Fatalf("sub: got %v", got)
		}
		return c.NoContent(http.StatusOK)
	}, Middleware(v))

	signed := f.signToken(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+signed)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMiddlewareRejects(t *testing.T) {
	f := newJWKSFixture(t)
	v := startedVerifier(t, f)

	e := echo.New()
	e.GET("/me", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, Middleware(v))

	cases := []struct {
		name   string
		header string
	}{
		{
			name: "missing header",
		},
		{
			name:   "non-bearer scheme",
			header: "Basic abc",
		},
		{
			name: "expired token",
			header: "Bearer " + f.signToken(t, func(tok jwt.Token) {
				_ = tok.Set(jwt.IssuedAtKey, time.Now().Add(-1*time.Hour))
				_ = tok.Set(jwt.ExpirationKey, time.Now().Add(-30*time.Minute))
			}),
		},
		{
			name: "wrong issuer",
			header: "Bearer " + f.signToken(t, func(tok jwt.Token) {
				_ = tok.Set(jwt.IssuerKey, "https://evil.example")
			}),
		},
		{
			name:   "garbage token",
			header: "Bearer not.a.jwt",
		},
		{
			name:   "wrong-key-signed token",
			header: "Bearer " + signWithStrangerKey(t),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tc.header != "" {
				req.Header.Set(echo.HeaderAuthorization, tc.header)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status: want 401 got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

// signWithStrangerKey signs a token with a fresh key the verifier has never
// seen. Used to assert wrong-signature rejection.
func signWithStrangerKey(t *testing.T) string {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa: %v", err)
	}
	tok, err := jwt.NewBuilder().
		Issuer(testIssuer).
		Subject("attacker").
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(5 * time.Minute)).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	key, err := jwk.FromRaw(priv)
	if err != nil {
		t.Fatalf("jwk: %v", err)
	}
	if err := key.Set(jwk.KeyIDKey, testKID); err != nil {
		t.Fatalf("kid: %v", err)
	}
	if err := key.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		t.Fatalf("alg: %v", err)
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return string(signed)
}

func TestJWKSCachePrimesOnStart(t *testing.T) {
	f := newJWKSFixture(t)
	_ = startedVerifier(t, f)
	if got := atomic.LoadInt32(f.calls); got == 0 {
		t.Fatal("expected JWKS endpoint to be hit on Start, was not")
	}
}

func TestStartFailsOnUnreachableJWKS(t *testing.T) {
	v, err := NewVerifier("test_client", "test.workos.local")
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}
	v.WithJWKSURL("http://127.0.0.1:1/does-not-exist").WithIssuer(testIssuer)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := v.Start(ctx); err == nil {
		t.Fatal("expected Start to fail when JWKS endpoint is unreachable")
	}
}

func TestBearerTokenParse(t *testing.T) {
	cases := []struct {
		header, want string
	}{
		{"", ""},
		{"Basic foo", ""},
		{"Bearer ", ""},
		{"Bearer  foo", "foo"},
		{"Bearer abc.def.ghi", "abc.def.ghi"},
	}
	for _, tc := range cases {
		t.Run(strings.ReplaceAll(tc.header, " ", "_"), func(t *testing.T) {
			if got := bearerToken(tc.header); got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}
