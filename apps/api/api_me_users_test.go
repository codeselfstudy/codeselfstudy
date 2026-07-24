package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/auth"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/db"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/session"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/users"
)

// These end-to-end tests prove newServer's /api/me wiring: with the account DB
// present the users handlers own /api/me and return the DB username; without it,
// the session's cookie-profile handler still answers. Reaching the session-gated
// handler needs a real authenticated request, so the tests mint a valid session
// cookie the same way session.Manager.Callback does.

const meUsersCookiePw = "0123456789abcdef0123456789abcdef" // >= 32 chars

// sealSession builds a session cookie exactly as internal/session seals one
// (SHA-256 key derivation, AES-256-GCM with the nonce prepended, and the
// sessionData JSON tags "at"/"u"). It is deliberately coupled to that format; if
// the seal changes, this helper must change with it.
func sealSession(t *testing.T, cookiePw, accessToken string, p session.Profile) string {
	t.Helper()
	key := sha256.Sum256([]byte("codeselfstudy:wos-seal:" + cookiePw))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatalf("aes: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("gcm: %v", err)
	}
	plain, err := json.Marshal(map[string]any{"at": accessToken, "rt": "", "sid": "", "u": p})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("nonce: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(aead.Seal(nonce, nonce, plain, nil))
}

func migratedStore(t *testing.T) *store.Store {
	t.Helper()
	d, err := db.Open("file:" + filepath.Join(t.TempDir(), "me.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	if err := db.Migrate(t.Context(), d); err != nil {
		_ = d.Close()
		t.Fatalf("db.Migrate: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return store.New(d)
}

func meSessionManager(t *testing.T, v *auth.Verifier) *session.Manager {
	t.Helper()
	m, err := session.New(session.Config{
		ClientID:       "client_test",
		APIKey:         "sk_test",
		CookiePassword: meUsersCookiePw,
		BaseURL:        "https://app.test",
	}, v)
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	return m
}

// authedMeRequest issues GET /api/me with a freshly-sealed cookie for the given
// profile and returns the decoded JSON body.
func authedMeRequest(t *testing.T, e http.Handler, token string, p session.Profile) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	// The session cookie is named "wos_session" (internal/session.cookieName).
	req.AddCookie(&http.Cookie{Name: "wos_session", Value: sealSession(t, meUsersCookiePw, token, p)})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	var body map[string]any
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
	}
	return rec.Code, body
}

func TestApiMeReturnsUsernameWhenUsersWired(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	sess := meSessionManager(t, v)

	st := migratedStore(t)
	if _, _, err := st.UpsertUserByWorkOSID(t.Context(), store.User{
		WorkOSID: "user_test_42", Email: "alice@example.com", Username: "ada",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	e := newServer(fixtureDir(t), v, sess, nil, users.New(st, nil))

	code, body := authedMeRequest(t, e, f.sign(t), session.Profile{ID: "user_test_42", Email: "alice@example.com"})
	if code != http.StatusOK {
		t.Fatalf("want 200 got %d", code)
	}
	if body["username"] != "ada" {
		t.Errorf("username = %v, want ada (the DB-backed handler)", body["username"])
	}
}

func TestApiMeFallsBackToCookieProfileWithoutUsers(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	sess := meSessionManager(t, v)
	e := newServer(fixtureDir(t), v, sess, nil, nil) // no users handlers

	code, body := authedMeRequest(t, e, f.sign(t), session.Profile{ID: "user_test_42", Email: "alice@example.com"})
	if code != http.StatusOK {
		t.Fatalf("want 200 got %d", code)
	}
	if body["email"] != "alice@example.com" {
		t.Errorf("email = %v, want alice@example.com (cookie profile)", body["email"])
	}
	if _, has := body["username"]; has {
		t.Errorf("cookie-profile /api/me should carry no username, got %v", body["username"])
	}
}

// TestAccountsWorkWithoutIngest locks in the decoupling: with the account DB
// present but the ingest pipeline absent (INGEST_TOKEN unset -> ing == nil), the
// account routes are mounted and gated while /api/ingest is not mounted at all.
func TestAccountsWorkWithoutIngest(t *testing.T) {
	f := newMeFixture(t)
	v := startedVerifierFor(t, f)
	sess := meSessionManager(t, v)
	e := newServer(fixtureDir(t), v, sess, nil, users.New(migratedStore(t), nil))

	// /api/ingest is not mounted -> JSON 404 (the /api/* reservation), not 401.
	req := httptest.NewRequest(http.MethodPost, "/api/ingest", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("/api/ingest without the pipeline: want 404 got %d", rec.Code)
	}

	// The account routes ARE mounted, so they gate (401 without a cookie) rather
	// than falling through to the 404 reservation.
	req = httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("/api/settings: want 401 (mounted and gated) got %d", rec.Code)
	}
}
