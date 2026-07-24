package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/auth"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/session"
)

// End-to-end check that /api/me wins over the /api/* catchall and that the
// auth middleware is the gate. Mirrors the unit-level coverage in
// internal/auth but via the same wiring main.go uses.

const (
	meTestIssuer = "https://test.workos.local"
	meTestKID    = "me-test-kid"
)

type meFixture struct {
	srv  *httptest.Server
	priv *rsa.PrivateKey
	hits *int32
}

func newMeFixture(t *testing.T) *meFixture {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa: %v", err)
	}
	pub, err := jwk.FromRaw(&priv.PublicKey)
	if err != nil {
		t.Fatalf("jwk pub: %v", err)
	}
	_ = pub.Set(jwk.KeyIDKey, meTestKID)
	_ = pub.Set(jwk.AlgorithmKey, jwa.RS256)
	set := jwk.NewSet()
	_ = set.AddKey(pub)

	var hits int32
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &meFixture{srv: srv, priv: priv, hits: &hits}
}

func (f *meFixture) sign(t *testing.T) string {
	t.Helper()
	tok, err := jwt.NewBuilder().
		Issuer(meTestIssuer).
		Subject("user_test_42").
		Claim("email", "alice@example.com").
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(5 * time.Minute)).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	priv, err := jwk.FromRaw(f.priv)
	if err != nil {
		t.Fatalf("jwk priv: %v", err)
	}
	_ = priv.Set(jwk.KeyIDKey, meTestKID)
	_ = priv.Set(jwk.AlgorithmKey, jwa.RS256)
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, priv))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return string(signed)
}

func startedVerifierFor(t *testing.T, f *meFixture) *auth.Verifier {
	t.Helper()
	v, err := auth.NewVerifier("test_client", "test.workos.local")
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	v.WithJWKSURL(f.srv.URL + "/jwks").WithIssuer(meTestIssuer)
	if err := v.Start(t.Context()); err != nil {
		t.Fatalf("start: %v", err)
	}
	return v
}

func TestApiMeReturnsClaims(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	e := newServer(fixtureDir(t), v, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+f.sign(t))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v body=%q", err, rec.Body.String())
	}
	if body["sub"] != "user_test_42" {
		t.Errorf("sub: got %v want user_test_42", body["sub"])
	}
	if body["email"] != "alice@example.com" {
		t.Errorf("email: got %v want alice@example.com", body["email"])
	}
	if body["iss"] != meTestIssuer {
		t.Errorf("iss: got %v want %s", body["iss"], meTestIssuer)
	}
}

func TestApiMeRejectsMissingToken(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	e := newServer(fixtureDir(t), v, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status: want 401 got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestApiMeHeadHonorsAuth(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	e := newServer(fixtureDir(t), v, nil, nil, nil)

	cases := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{"missing token", "", http.StatusUnauthorized},
		{"valid token", "Bearer " + f.sign(t), http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodHead, "/api/me", nil)
			if tc.header != "" {
				req.Header.Set(echo.HeaderAuthorization, tc.header)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status: want %d got %d body=%s", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

// When a session Manager is wired, newServer must mount the /auth/* routes and
// gate /api/me on the session cookie (not the Bearer path). This covers the
// production branch of newServer that the other tests, which pass sess=nil,
// don't reach.
func TestSessionWiringMountsAuthAndGatesMe(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	sess, err := session.New(session.Config{
		ClientID:       "client_test",
		APIKey:         "sk_test",
		CookiePassword: "0123456789abcdef0123456789abcdef",
		BaseURL:        "https://app.test",
	}, v)
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	e := newServer(fixtureDir(t), v, sess, nil, nil)

	t.Run("auth login redirects", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/login?returnTo=/events/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusFound {
			t.Fatalf("/auth/login: want 302 got %d", rec.Code)
		}
	})

	t.Run("api me needs the cookie", func(t *testing.T) {
		// A Bearer token must no longer satisfy /api/me — the session cookie is
		// the only accepted credential once the Manager is wired.
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+f.sign(t))
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("/api/me without cookie: want 401 got %d", rec.Code)
		}
	})
}

func TestApiMeIsDisabledWithoutVerifier(t *testing.T) {
	// /api/me without a verifier should fall through to the /api/* JSON 404,
	// not return 401 — that's the smoke-test ergonomics we lean on for
	// runs without WorkOS env config.
	e := newServer(fixtureDir(t), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: want 404 got %d body=%s", rec.Code, rec.Body.String())
	}
}
