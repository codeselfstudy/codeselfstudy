package users_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/db"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/session"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/users"
)

var ctx = context.Background()

// newStore opens a fresh migrated temp SQLite database (the same harness the
// store package's own tests use), so these handler tests run against the real
// store rather than a mock.
func newStore(t *testing.T) *store.Store {
	t.Helper()
	d, err := db.Open("file:" + filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	if err := db.Migrate(ctx, d); err != nil {
		_ = d.Close()
		t.Fatalf("db.Migrate: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return store.New(d)
}

func prof(id, email string) session.Profile {
	return session.Profile{ID: id, Email: email, FirstName: "Ada", LastName: "Lovelace", ProfilePictureURL: "https://img/a.png"}
}

// mount builds an Echo with the account routes, gated by a stub middleware that
// injects p (nil → unauthenticated, so a handler's own session.User check
// produces the 401 the real middleware would). Faithful enough to exercise the
// handlers without the full cookie/JWKS flow, which api_me_test covers end-to-end.
func mount(h *users.Handlers, p *session.Profile) *echo.Echo {
	e := echo.New()
	api := e.Group("/api", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if p != nil {
				session.ContextWithUser(c, *p)
			}
			return next(c)
		}
	})
	h.Register(api)
	return e
}

func do(t *testing.T, e *echo.Echo, method, target, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
		r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, r)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return m
}

type fakePoster struct {
	mu sync.Mutex
	n  int
}

func (f *fakePoster) Post(context.Context, []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.n++
	return nil
}

func (f *fakePoster) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.n
}

func TestRoutesRequireSession(t *testing.T) {
	h := users.New(newStore(t), nil)
	e := mount(h, nil) // unauthenticated

	for _, tc := range []struct{ method, target string }{
		{http.MethodGet, "/api/me"},
		{http.MethodGet, "/api/settings"},
		{http.MethodPatch, "/api/settings"},
		{http.MethodPost, "/api/settings/delete-request"},
	} {
		rec := do(t, e, tc.method, tc.target, "", nil)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: want 401 got %d", tc.method, tc.target, rec.Code)
		}
	}
}

func TestMeReturnsUsername(t *testing.T) {
	st := newStore(t)
	if _, _, err := st.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_1", Email: "a@x.com", Username: "ada"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	p := prof("wos_1", "a@x.com")
	e := mount(users.New(st, nil), &p)

	rec := do(t, e, http.MethodGet, "/api/me", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	if body["username"] != "ada" {
		t.Errorf("username = %v, want ada", body["username"])
	}
	if body["email"] != "a@x.com" {
		t.Errorf("email = %v, want a@x.com", body["email"])
	}
}

func TestMeDegradesWithoutRow(t *testing.T) {
	// No row seeded: /api/me must still 200 with the session profile, never error.
	p := prof("wos_missing", "ghost@x.com")
	e := mount(users.New(newStore(t), nil), &p)

	rec := do(t, e, http.MethodGet, "/api/me", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	if body["username"] != "" {
		t.Errorf("username = %v, want empty (degraded)", body["username"])
	}
	if body["email"] != "ghost@x.com" {
		t.Errorf("email = %v, want the session email", body["email"])
	}
}

func TestGetSettings(t *testing.T) {
	st := newStore(t)
	if _, _, err := st.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_1", Email: "a@x.com", Username: "ada"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := st.SetTimezone(ctx, 1, "America/Los_Angeles"); err != nil {
		t.Fatalf("tz: %v", err)
	}
	p := prof("wos_1", "a@x.com")
	e := mount(users.New(st, nil), &p)

	rec := do(t, e, http.MethodGet, "/api/settings", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	if body["username"] != "ada" || body["timezone"] != "America/Los_Angeles" {
		t.Errorf("settings = %v, want ada / America/Los_Angeles", body)
	}
	if body["deletion_requested_at"] != nil {
		t.Errorf("deletion_requested_at = %v, want null", body["deletion_requested_at"])
	}
}

func TestPatchTimezoneAndUsername(t *testing.T) {
	st := newStore(t)
	if _, _, err := st.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_1", Email: "a@x.com", Username: "ada"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	p := prof("wos_1", "a@x.com")
	e := mount(users.New(st, nil), &p)

	rec := do(t, e, http.MethodPatch, "/api/settings", `{"username":"ada-lovelace","timezone":"Europe/London"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	if body["username"] != "ada-lovelace" || body["timezone"] != "Europe/London" {
		t.Errorf("patched = %v, want ada-lovelace / Europe/London", body)
	}
}

func TestPatchUsernameValidation(t *testing.T) {
	st := newStore(t)
	if _, _, err := st.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_1", Email: "a@x.com", Username: "ada"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	p := prof("wos_1", "a@x.com")
	e := mount(users.New(st, nil), &p)

	cases := []struct {
		name     string
		body     string
		wantCode int
		wantErr  string
	}{
		{"too short", `{"username":"ab"}`, http.StatusBadRequest, "username_invalid"},
		{"bad char", `{"username":"has space"}`, http.StatusBadRequest, "username_invalid"},
		{"reserved", `{"username":"admin"}`, http.StatusBadRequest, "username_reserved"},
		{"bad timezone", `{"timezone":"Mars/Phobos"}`, http.StatusBadRequest, "timezone_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(t, e, http.MethodPatch, "/api/settings", tc.body, nil)
			if rec.Code != tc.wantCode {
				t.Fatalf("want %d got %d body=%s", tc.wantCode, rec.Code, rec.Body.String())
			}
			if got := decode(t, rec)["error"]; got != tc.wantErr {
				t.Errorf("error = %v, want %s", got, tc.wantErr)
			}
		})
	}
}

func TestPatchIsAtomicOnInvalidField(t *testing.T) {
	st := newStore(t)
	for _, u := range []store.User{
		{WorkOSID: "wos_1", Email: "a@x.com", Username: "ada"},
		{WorkOSID: "wos_2", Email: "b@x.com", Username: "bob"},
	} {
		if _, _, err := st.UpsertUserByWorkOSID(ctx, u); err != nil {
			t.Fatalf("seed %s: %v", u.Username, err)
		}
	}
	p := prof("wos_1", "a@x.com")
	e := mount(users.New(st, nil), &p)

	// A valid timezone paired with a failing username must not commit the timezone:
	// all fields are validated, and the username writes first, before any timezone
	// write.
	cases := []struct {
		name, body string
		wantCode   int
	}{
		{"reserved username", `{"timezone":"Europe/London","username":"admin"}`, http.StatusBadRequest},
		{"taken username", `{"timezone":"Europe/London","username":"bob"}`, http.StatusConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(t, e, http.MethodPatch, "/api/settings", tc.body, nil)
			if rec.Code != tc.wantCode {
				t.Fatalf("want %d got %d body=%s", tc.wantCode, rec.Code, rec.Body.String())
			}
			if u, _ := st.GetUserByWorkOSID(ctx, "wos_1"); u.Timezone != "" {
				t.Errorf("timezone = %q, want empty (a rejected username must not commit the timezone)", u.Timezone)
			}
		})
	}
}

func TestPatchUsernameTaken(t *testing.T) {
	st := newStore(t)
	for _, u := range []store.User{
		{WorkOSID: "wos_a", Email: "a@x.com", Username: "alice"},
		{WorkOSID: "wos_b", Email: "b@x.com", Username: "bob"},
	} {
		if _, _, err := st.UpsertUserByWorkOSID(ctx, u); err != nil {
			t.Fatalf("seed %s: %v", u.Username, err)
		}
	}
	p := prof("wos_a", "a@x.com")
	e := mount(users.New(st, nil), &p)

	// Case-insensitive collision with bob's handle.
	rec := do(t, e, http.MethodPatch, "/api/settings", `{"username":"BOB"}`, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("want 409 got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := decode(t, rec)["error"]; got != "username_taken" {
		t.Errorf("error = %v, want username_taken", got)
	}
}

func TestPatchUsernameRateLimited(t *testing.T) {
	st := newStore(t)
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	st.Now = func() time.Time { return clock }

	if _, _, err := st.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_1", Email: "a@x.com", Username: "ada"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// First change stamps username_changed_at at the current clock (no prior stamp,
	// so the cutoff is irrelevant here).
	if err := st.SetUsername(ctx, 1, "ada1", clock); err != nil {
		t.Fatalf("SetUsername: %v", err)
	}
	// One day later — inside the 30-day cooldown.
	clock = clock.Add(24 * time.Hour)

	p := prof("wos_1", "a@x.com")
	e := mount(users.New(st, nil), &p)

	rec := do(t, e, http.MethodPatch, "/api/settings", `{"username":"ada2"}`, nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429 got %d body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	if body["error"] != "rate_limited" {
		t.Errorf("error = %v, want rate_limited", body["error"])
	}
	if got, _ := body["retry_after_days"].(float64); got != 29 {
		t.Errorf("retry_after_days = %v, want 29", body["retry_after_days"])
	}
	// The username must be unchanged.
	if u, _ := st.GetUserByWorkOSID(ctx, "wos_1"); u.Username != "ada1" {
		t.Errorf("username = %q, want ada1 (rejected change)", u.Username)
	}
}

func TestDeleteRequestIdempotentAndNotifies(t *testing.T) {
	st := newStore(t)
	if _, _, err := st.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_1", Email: "a@x.com", Username: "ada"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	fp := &fakePoster{}
	p := prof("wos_1", "a@x.com")
	e := mount(users.New(st, fp), &p)

	rec := do(t, e, http.MethodPost, "/api/settings/delete-request", `{"reason":"moving on"}`, nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("first: want 202 got %d body=%s", rec.Code, rec.Body.String())
	}
	if decode(t, rec)["deletion_requested_at"] == nil {
		t.Errorf("first: deletion_requested_at should be set")
	}

	// A second click is idempotent: still 202, but no second Slack ping (nothing
	// new was created).
	rec = do(t, e, http.MethodPost, "/api/settings/delete-request", `{"reason":"again"}`, nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("second: want 202 got %d", rec.Code)
	}
	if fp.count() != 1 {
		t.Errorf("slack posts = %d, want exactly 1", fp.count())
	}

	// The pending row now surfaces on GET /settings.
	rec = do(t, e, http.MethodGet, "/api/settings", "", nil)
	if decode(t, rec)["deletion_requested_at"] == nil {
		t.Errorf("GET settings: deletion_requested_at should be set")
	}
}

func TestPatchRejectsCrossOrigin(t *testing.T) {
	st := newStore(t)
	if _, _, err := st.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_1", Email: "a@x.com", Username: "ada"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	p := prof("wos_1", "a@x.com")
	e := mount(users.New(st, nil), &p)

	cases := []struct {
		name    string
		headers map[string]string
	}{
		{"sec-fetch-site cross-site", map[string]string{"Sec-Fetch-Site": "cross-site"}},
		{"foreign origin", map[string]string{"Origin": "https://evil.example"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(t, e, http.MethodPatch, "/api/settings", `{"timezone":"Europe/London"}`, tc.headers)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("want 403 got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}

	// A same-origin write (Origin host == request host) is allowed through.
	rec := do(t, e, http.MethodPatch, "/api/settings", `{"timezone":"Europe/London"}`, map[string]string{
		"Origin": "http://example.com", "Sec-Fetch-Site": "same-origin",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("same-origin: want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
}
