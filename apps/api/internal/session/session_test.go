package session

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/auth"
)

const (
	testIssuer         = "https://session.workos.local"
	testKID            = "session-test-kid"
	testCookiePassword = "0123456789abcdef0123456789abcdef" // exactly 32 chars
	altCookiePassword  = "ffffffffffffffffffffffffffffffff" // a different 32-char key
)

// jwksFixture serves a public JWKS and signs access tokens the verifier trusts,
// mirroring the fixtures in the auth package.
type jwksFixture struct {
	srv  *httptest.Server
	priv *rsa.PrivateKey
}

func newJWKSFixture(t *testing.T) *jwksFixture {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa: %v", err)
	}
	pub, err := jwk.FromRaw(&priv.PublicKey)
	if err != nil {
		t.Fatalf("jwk pub: %v", err)
	}
	_ = pub.Set(jwk.KeyIDKey, testKID)
	_ = pub.Set(jwk.AlgorithmKey, jwa.RS256)
	set := jwk.NewSet()
	_ = set.AddKey(pub)

	mux := http.NewServeMux()
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &jwksFixture{srv: srv, priv: priv}
}

func (f *jwksFixture) signAccess(t *testing.T, exp time.Time) string {
	t.Helper()
	tok, err := jwt.NewBuilder().
		Issuer(testIssuer).
		Subject("user_test").
		Claim("sid", "session_test_1").
		IssuedAt(time.Now().Add(-1 * time.Minute)).
		Expiration(exp).
		Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	key, err := jwk.FromRaw(f.priv)
	if err != nil {
		t.Fatalf("jwk priv: %v", err)
	}
	_ = key.Set(jwk.KeyIDKey, testKID)
	_ = key.Set(jwk.AlgorithmKey, jwa.RS256)
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return string(signed)
}

func startedVerifier(t *testing.T, f *jwksFixture) *auth.Verifier {
	t.Helper()
	v, err := auth.NewVerifier("client_test", "session.workos.local")
	if err != nil {
		t.Fatalf("verifier: %v", err)
	}
	v.WithJWKSURL(f.srv.URL + "/jwks").WithIssuer(testIssuer)
	if err := v.Start(t.Context()); err != nil {
		t.Fatalf("start: %v", err)
	}
	return v
}

func newTestManagerPw(t *testing.T, f *jwksFixture, pw string) *Manager {
	t.Helper()
	m, err := New(Config{
		ClientID:       "client_test",
		APIKey:         "sk_test",
		CookiePassword: pw,
		BaseURL:        "https://app.test",
	}, startedVerifier(t, f))
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	return m
}

func newTestManager(t *testing.T, f *jwksFixture) *Manager {
	return newTestManagerPw(t, f, testCookiePassword)
}

// --- helpers ---------------------------------------------------------------

func mustSeal(t *testing.T, m *Manager, d sessionData) string {
	t.Helper()
	s, err := m.seal(d)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	return s
}

func cookie(value string) *http.Cookie {
	return &http.Cookie{Name: cookieName, Value: value}
}

func serveMe(m *Manager, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	e := echo.New()
	e.GET("/api/me", m.HandleMe, m.Middleware())
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// cookieFrom returns the value of a freshly set session cookie, or "".
func cookieFrom(rec *httptest.ResponseRecorder) string {
	for _, c := range rec.Result().Cookies() {
		if c.Name == cookieName && c.Value != "" {
			return c.Value
		}
	}
	return ""
}

func flipLast(s string) string {
	if s == "" {
		return "A"
	}
	repl := byte('A')
	if s[len(s)-1] == 'A' {
		repl = 'B'
	}
	return s[:len(s)-1] + string(repl)
}

// --- tests -----------------------------------------------------------------

func TestConfigEnabled(t *testing.T) {
	base := Config{ClientID: "c", APIKey: "k", CookiePassword: testCookiePassword, BaseURL: "https://x"}
	if !base.Enabled() {
		t.Fatal("complete config should be enabled")
	}
	cases := map[string]Config{
		"no client":      {APIKey: "k", CookiePassword: testCookiePassword, BaseURL: "https://x"},
		"no key":         {ClientID: "c", CookiePassword: testCookiePassword, BaseURL: "https://x"},
		"no base":        {ClientID: "c", APIKey: "k", CookiePassword: testCookiePassword},
		"short password": {ClientID: "c", APIKey: "k", CookiePassword: "tooshort", BaseURL: "https://x"},
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			if cfg.Enabled() {
				t.Errorf("expected disabled for %q", name)
			}
		})
	}
}

func TestSafeReturnTo(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "/"},
		{"/", "/"},
		{"/events/", "/events/"},
		{"/events/?tab=upcoming", "/events/?tab=upcoming"},
		{"//evil.com", "/"},
		{"https://evil.com", "/"},
		{"http://evil.com/path", "/"},
		{"/\\evil.com", "/"},
		{"relative/path", "/"},
		{"javascript:alert(1)", "/"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := safeReturnTo(tc.in); got != tc.want {
				t.Errorf("safeReturnTo(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSealUnsealRoundTrip(t *testing.T) {
	m := newTestManager(t, newJWKSFixture(t))
	in := sessionData{AccessToken: "at", RefreshToken: "rt", SessionID: "sid", User: profile{ID: "u1", Email: "a@b.com"}}
	sealed := mustSeal(t, m, in)
	out, err := m.unseal(sealed)
	if err != nil {
		t.Fatalf("unseal: %v", err)
	}
	if out != in {
		t.Fatalf("round-trip mismatch: %+v vs %+v", out, in)
	}
}

func TestUnsealRejectsTamperAndWrongKey(t *testing.T) {
	f := newJWKSFixture(t)
	m := newTestManager(t, f)
	sealed := mustSeal(t, m, sessionData{AccessToken: "at"})

	if _, err := m.unseal(flipLast(sealed)); err == nil {
		t.Error("expected unseal to reject tampered ciphertext")
	}

	other := newTestManagerPw(t, f, altCookiePassword)
	if _, err := m.unseal(mustSeal(t, other, sessionData{AccessToken: "x"})); err == nil {
		t.Error("expected cross-password unseal to fail")
	}
}

func TestStateSignVerify(t *testing.T) {
	f := newJWKSFixture(t)
	m := newTestManager(t, f)
	state := m.signState("/events/")
	got, ok := m.verifyState(state)
	if !ok || got != "/events/" {
		t.Fatalf("verifyState round-trip: got %q ok=%v", got, ok)
	}
	if _, ok := m.verifyState(flipLast(state)); ok {
		t.Error("tampered state should not verify")
	}
	if _, ok := m.verifyState("not-base64!!"); ok {
		t.Error("garbage state should not verify")
	}
	other := newTestManagerPw(t, f, altCookiePassword)
	if _, ok := m.verifyState(other.signState("/events/")); ok {
		t.Error("state signed with a different key should not verify")
	}
}

func TestMiddlewareValidCookie(t *testing.T) {
	f := newJWKSFixture(t)
	m := newTestManager(t, f)
	sealed := mustSeal(t, m, sessionData{
		AccessToken:  f.signAccess(t, time.Now().Add(5*time.Minute)),
		RefreshToken: "rt",
		SessionID:    "sid",
		User:         profile{ID: "u1", Email: "ada@example.com", FirstName: "Ada", LastName: "Lovelace"},
	})

	rec := serveMe(m, cookie(sealed))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("cache-control: got %q want no-store", cc)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["email"] != "ada@example.com" {
		t.Errorf("email: got %v", body["email"])
	}
	if body["name"] != "Ada Lovelace" {
		t.Errorf("name: got %v", body["name"])
	}
}

func TestMiddlewareRejectsMissingAndTampered(t *testing.T) {
	f := newJWKSFixture(t)
	m := newTestManager(t, f)
	good := mustSeal(t, m, sessionData{AccessToken: f.signAccess(t, time.Now().Add(5*time.Minute))})

	t.Run("missing", func(t *testing.T) {
		if rec := serveMe(m); rec.Code != http.StatusUnauthorized {
			t.Fatalf("want 401 got %d", rec.Code)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if rec := serveMe(m, cookie("")); rec.Code != http.StatusUnauthorized {
			t.Fatalf("want 401 got %d", rec.Code)
		}
	})
	t.Run("tampered", func(t *testing.T) {
		if rec := serveMe(m, cookie(flipLast(good))); rec.Code != http.StatusUnauthorized {
			t.Fatalf("want 401 got %d", rec.Code)
		}
	})
}

func TestMiddlewareRefreshesExpired(t *testing.T) {
	f := newJWKSFixture(t)
	m := newTestManager(t, f)
	fresh := f.signAccess(t, time.Now().Add(5*time.Minute))
	var gotRefreshToken string
	m.refresh = func(_ context.Context, rt string) (string, string, error) {
		gotRefreshToken = rt
		return fresh, "rt_new", nil
	}

	sealed := mustSeal(t, m, sessionData{
		AccessToken:  f.signAccess(t, time.Now().Add(-1*time.Minute)), // expired
		RefreshToken: "rt_old",
		User:         profile{Email: "a@b.com"},
	})

	rec := serveMe(m, cookie(sealed))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotRefreshToken != "rt_old" {
		t.Errorf("refresh called with %q, want rt_old", gotRefreshToken)
	}
	set := cookieFrom(rec)
	if set == "" {
		t.Fatal("expected a refreshed Set-Cookie")
	}
	out, err := m.unseal(set)
	if err != nil {
		t.Fatalf("unseal refreshed: %v", err)
	}
	if out.AccessToken != fresh || out.RefreshToken != "rt_new" {
		t.Errorf("refreshed cookie tokens = %q/%q", out.AccessToken, out.RefreshToken)
	}
}

func TestMiddlewareRefreshFailureClears(t *testing.T) {
	f := newJWKSFixture(t)
	m := newTestManager(t, f)
	m.refresh = func(_ context.Context, _ string) (string, string, error) {
		return "", "", errors.New("refresh rejected")
	}
	sealed := mustSeal(t, m, sessionData{
		AccessToken:  f.signAccess(t, time.Now().Add(-1*time.Minute)),
		RefreshToken: "rt_old",
	})
	rec := serveMe(m, cookie(sealed))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status: want 401 got %d", rec.Code)
	}
}

func TestLoginRedirectsToWorkOS(t *testing.T) {
	m := newTestManager(t, newJWKSFixture(t))
	e := echo.New()
	e.GET("/auth/login", m.Login)
	req := httptest.NewRequest(http.MethodGet, "/auth/login?returnTo=/events/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status: want 302 got %d", rec.Code)
	}
	u, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("bad Location: %v", err)
	}
	if u.Query().Get("provider") != "authkit" {
		t.Errorf("provider: got %q", u.Query().Get("provider"))
	}
	if u.Query().Get("client_id") != "client_test" {
		t.Errorf("client_id: got %q", u.Query().Get("client_id"))
	}
	if u.Query().Get("redirect_uri") != "https://app.test/auth/callback" {
		t.Errorf("redirect_uri: got %q", u.Query().Get("redirect_uri"))
	}
	if got, ok := m.verifyState(u.Query().Get("state")); !ok || got != "/events/" {
		t.Errorf("state did not round-trip: got %q ok=%v", got, ok)
	}
}
