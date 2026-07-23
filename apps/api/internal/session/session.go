// Package session implements a WorkOS login backed by a first-party, encrypted
// cookie set on this app's own origin.
//
// The previous browser-only flow (@workos-inc/authkit-react) kept the refresh
// token in a cookie on WorkOS's API domain. That cookie is third-party to
// codeselfstudy.com, so browsers block it — and because the site is a static,
// multi-page app that fully reloads on every navigation, the session could not
// be restored and the user appeared signed out. WorkOS's own fix for that is a
// paid custom auth domain.
//
// Instead, the Go server (which already serves the static site and /api on the
// same origin) performs the OAuth code exchange itself and seals
// {access token, refresh token, session id, user profile} into an AES-256-GCM
// cookie on this origin. Being first-party, it survives reloads. On each
// request the access token is validated against WorkOS's JWKS (reusing
// auth.Verifier); when it has expired the refresh token mints a new pair and the
// cookie is re-sealed transparently.
package session

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
	"golang.org/x/sync/singleflight"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/auth"
)

// cookieName is the session cookie. contextUserKey is where the authenticated
// profile is stashed for downstream handlers (HandleMe).
const (
	cookieName     = "wos_session"
	contextUserKey = "session:user"
	cookieMaxAge   = 30 * 24 * time.Hour

	// nonceCookieName holds the one-time CSRF nonce that binds an in-flight
	// OAuth login to the browser that started it. Scoped to /auth and short-lived.
	nonceCookieName = "wos_oauth"
	nonceMaxAge     = 10 * time.Minute
)

// Config is the server-side WorkOS configuration the session flow needs, on top
// of the client id / api hostname the JWKS Verifier already reads.
type Config struct {
	ClientID       string // WORKOS_CLIENT_ID
	APIKey         string // WORKOS_API_KEY (server secret, sk_...)
	CookiePassword string // WORKOS_COOKIE_PASSWORD (>= 32 chars; seals the cookie)
	BaseURL        string // APP_BASE_URL, e.g. https://codeselfstudy.com
}

// LoadConfig reads the session config from the environment. getenv is injected
// so tests and main share one code path.
func LoadConfig(getenv func(string) string) Config {
	return Config{
		ClientID:       firstNonEmpty(getenv("WORKOS_CLIENT_ID"), getenv("PUBLIC_WORKOS_CLIENT_ID")),
		APIKey:         getenv("WORKOS_API_KEY"),
		CookiePassword: getenv("WORKOS_COOKIE_PASSWORD"),
		BaseURL:        getenv("APP_BASE_URL"),
	}
}

// Enabled reports whether every value required to run the server-side session
// flow is present. A short cookie password is treated as absent: WorkOS requires
// at least 32 characters of entropy to seal a session safely.
func (c Config) Enabled() bool {
	return c.ClientID != "" && c.APIKey != "" && len(c.CookiePassword) >= 32 && c.BaseURL != ""
}

// Manager owns the sealed-cookie session flow: the /auth/* routes, the gate
// middleware, and the /api/me handler.
type Manager struct {
	clientID    string
	verifier    *auth.Verifier
	aead        cipher.AEAD
	stateKey    []byte
	origin      string // scheme://host, no trailing slash
	callbackURL string
	secure      bool

	// refresh exchanges a refresh token for a new access/refresh pair. A field
	// so tests can substitute the WorkOS call.
	refresh refreshFunc

	// refreshGroup coalesces concurrent refreshes for the same session so the
	// single-use refresh token is exchanged exactly once (see Middleware).
	refreshGroup singleflight.Group
}

type refreshFunc func(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)

// New builds a Manager. It requires a fully populated Config and a started
// Verifier (used to validate access tokens). It configures the WorkOS SDK's
// default client with the API key.
func New(cfg Config, verifier *auth.Verifier) (*Manager, error) {
	if !cfg.Enabled() {
		return nil, errors.New("session: incomplete config (need WORKOS_CLIENT_ID, WORKOS_API_KEY, WORKOS_COOKIE_PASSWORD>=32, APP_BASE_URL)")
	}
	if verifier == nil {
		return nil, errors.New("session: verifier is required")
	}
	base, err := url.Parse(cfg.BaseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("session: invalid APP_BASE_URL %q", cfg.BaseURL)
	}

	// Two independent keys derived from the one password: one encrypts the
	// cookie, the other signs the OAuth state. Distinct domain-separation
	// prefixes keep them from ever being the same bytes.
	encKey := sha256.Sum256([]byte("codeselfstudy:wos-seal:" + cfg.CookiePassword))
	stateKey := sha256.Sum256([]byte("codeselfstudy:wos-state:" + cfg.CookiePassword))

	block, err := aes.NewCipher(encKey[:])
	if err != nil {
		return nil, fmt.Errorf("session: aes: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("session: gcm: %w", err)
	}

	usermanagement.SetAPIKey(cfg.APIKey)

	origin := base.Scheme + "://" + base.Host
	m := &Manager{
		clientID:    cfg.ClientID,
		verifier:    verifier,
		aead:        aead,
		stateKey:    stateKey[:],
		origin:      origin,
		callbackURL: origin + "/auth/callback",
		secure:      base.Scheme == "https",
	}
	m.refresh = func(ctx context.Context, rt string) (string, string, error) {
		resp, err := usermanagement.AuthenticateWithRefreshToken(ctx, usermanagement.AuthenticateWithRefreshTokenOpts{
			ClientID:     m.clientID,
			RefreshToken: rt,
		})
		if err != nil {
			return "", "", err
		}
		return resp.AccessToken, resp.RefreshToken, nil
	}
	return m, nil
}

// Register mounts the auth routes. They are concrete GET paths, so Echo matches
// them ahead of the static "/*" catchall.
func (m *Manager) Register(e *echo.Echo) {
	e.GET("/auth/login", m.Login)
	e.GET("/auth/callback", m.Callback)
	e.GET("/auth/logout", m.Logout)
}

// Login sends the browser to the WorkOS hosted AuthKit page. The post-login
// destination (returnTo) is validated to a same-origin path and signed into the
// OAuth state. A fresh random nonce is signed into the state AND set in a
// short-lived HttpOnly cookie, so the callback can confirm the redirect came
// back to the same browser that started the flow (login-CSRF protection).
func (m *Manager) Login(c echo.Context) error {
	noStore(c)
	returnTo := safeReturnTo(c.QueryParam("returnTo"))
	nonce, err := randomToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	m.setNonceCookie(c, nonce)
	authURL, err := usermanagement.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		ClientID:    m.clientID,
		RedirectURI: m.callbackURL,
		Provider:    "authkit",
		State:       m.signState(returnTo, nonce),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.Redirect(http.StatusFound, authURL.String())
}

// Callback completes the OAuth exchange: it validates the signed state, trades
// the code for tokens, seals them into the cookie, and redirects to returnTo.
// Any failure lands the user back on the home page rather than an error screen.
func (m *Manager) Callback(c echo.Context) error {
	noStore(c)

	// The nonce cookie is one-time: read it, then always clear it, whatever the
	// outcome, so a stale nonce can't be reused by a later request.
	cookieNonce := ""
	if ck, err := c.Cookie(nonceCookieName); err == nil {
		cookieNonce = ck.Value
	}
	m.clearNonceCookie(c)

	if c.QueryParam("error") != "" {
		return c.Redirect(http.StatusFound, "/")
	}
	code := c.QueryParam("code")
	returnTo, stateNonce, ok := m.verifyState(c.QueryParam("state"))
	if code == "" || !ok {
		return c.Redirect(http.StatusFound, "/")
	}
	// Two-channel CSRF check: the nonce signed into the state must match the one
	// this browser was given at Login. A missing or mismatched cookie means the
	// callback did not originate from a login this browser started.
	if cookieNonce == "" || subtle.ConstantTimeCompare([]byte(cookieNonce), []byte(stateNonce)) != 1 {
		return c.Redirect(http.StatusFound, "/")
	}
	returnTo = safeReturnTo(returnTo)

	ctx := c.Request().Context()
	resp, err := usermanagement.AuthenticateWithCode(ctx, usermanagement.AuthenticateWithCodeOpts{
		ClientID: m.clientID,
		Code:     code,
	})
	if err != nil {
		return c.Redirect(http.StatusFound, "/")
	}

	sealed, err := m.seal(sessionData{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		SessionID:    m.sessionID(ctx, resp.AccessToken),
		User:         profileFrom(resp.User),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	m.setCookie(c, sealed)
	return c.Redirect(http.StatusFound, returnTo)
}

// Logout clears the local cookie and redirects through the WorkOS logout URL so
// the WorkOS session is ended too, then back to a same-origin returnTo.
func (m *Manager) Logout(c echo.Context) error {
	noStore(c)
	returnTo := safeReturnTo(c.QueryParam("returnTo"))

	var sid string
	if ck, err := c.Cookie(cookieName); err == nil && ck.Value != "" {
		if data, err := m.unseal(ck.Value); err == nil {
			sid = data.SessionID
		}
	}
	m.clearCookie(c)

	if sid == "" {
		return c.Redirect(http.StatusFound, returnTo)
	}
	logoutURL, err := usermanagement.GetLogoutURL(usermanagement.GetLogoutURLOpts{
		SessionID: sid,
		ReturnTo:  m.origin + returnTo,
	})
	if err != nil {
		return c.Redirect(http.StatusFound, returnTo)
	}
	return c.Redirect(http.StatusFound, logoutURL.String())
}

// Middleware gates a route on a valid session cookie. It validates the access
// token, silently refreshes and re-seals when it has expired, and puts the user
// profile on the context. On any unrecoverable failure it clears the cookie and
// returns 401.
func (m *Manager) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			noStore(c)
			ck, err := c.Cookie(cookieName)
			if err != nil || ck.Value == "" {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			data, err := m.unseal(ck.Value)
			if err != nil {
				m.clearCookie(c)
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			ctx := c.Request().Context()
			if _, err := m.verifier.ParseToken(ctx, data.AccessToken); err != nil {
				// Expired or otherwise unusable — try one refresh. A cookie an
				// attacker cannot forge (AES-GCM) means the refresh token here is
				// genuine; WorkOS rejects a stale one, which we turn into 401.
				//
				// WorkOS refresh tokens are single-use, so concurrent requests
				// carrying the same expired cookie must not each call refresh —
				// the second would present an already-spent token, fail, and
				// spuriously log the user out. singleflight collapses them to one
				// exchange keyed by the refresh token; all waiters share its result.
				access, refresh, rerr := m.refreshOnce(ctx, data.RefreshToken)
				if rerr != nil {
					m.clearCookie(c)
					return echo.NewHTTPError(http.StatusUnauthorized)
				}
				if _, verr := m.verifier.ParseToken(ctx, access); verr != nil {
					m.clearCookie(c)
					return echo.NewHTTPError(http.StatusUnauthorized)
				}
				data.AccessToken = access
				data.RefreshToken = refresh
				data.SessionID = m.sessionID(ctx, access)
				sealed, serr := m.seal(data)
				if serr != nil {
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				m.setCookie(c, sealed)
			}

			c.Set(contextUserKey, data.User)
			return next(c)
		}
	}
}

// HandleMe returns the signed-in user. It reads the profile the middleware
// placed on the context (sealed in the cookie at login), so it needs no WorkOS
// round-trip.
func (m *Manager) HandleMe(c echo.Context) error {
	p, ok := c.Get(contextUserKey).(profile)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":     p.ID,
		"email":  p.Email,
		"name":   strings.TrimSpace(p.FirstName + " " + p.LastName),
		"avatar": p.ProfilePictureURL,
	})
}

// sessionData is the payload sealed into the cookie.
type sessionData struct {
	AccessToken  string  `json:"at"`
	RefreshToken string  `json:"rt"`
	SessionID    string  `json:"sid"`
	User         profile `json:"u"`
}

// profile is the subset of the WorkOS user the frontend renders.
type profile struct {
	ID                string `json:"id"`
	Email             string `json:"email"`
	FirstName         string `json:"first_name,omitempty"`
	LastName          string `json:"last_name,omitempty"`
	ProfilePictureURL string `json:"avatar,omitempty"`
}

func profileFrom(u usermanagement.User) profile {
	return profile{
		ID:                u.ID,
		Email:             u.Email,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		ProfilePictureURL: u.ProfilePictureURL,
	}
}

// sessionID extracts the WorkOS session id (`sid` claim) from a validated
// access token, for use at logout. Best-effort: an unreadable token yields "".
func (m *Manager) sessionID(ctx context.Context, accessToken string) string {
	tok, err := m.verifier.ParseToken(ctx, accessToken)
	if err != nil {
		return ""
	}
	if sid, ok := tok.Get("sid"); ok {
		if s, ok := sid.(string); ok {
			return s
		}
	}
	return ""
}

// seal encrypts the session payload with AES-256-GCM and returns a URL-safe
// base64 string (nonce prepended) suitable for a cookie value.
func (m *Manager) seal(d sessionData) (string, error) {
	plain, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, m.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ct := m.aead.Seal(nonce, nonce, plain, nil)
	return base64.RawURLEncoding.EncodeToString(ct), nil
}

// unseal reverses seal: it authenticates and decrypts the cookie value, failing
// closed on any tamper, truncation, or wrong-key input.
func (m *Manager) unseal(s string) (sessionData, error) {
	var d sessionData
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return d, err
	}
	ns := m.aead.NonceSize()
	if len(raw) < ns {
		return d, errors.New("session: ciphertext too short")
	}
	plain, err := m.aead.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return d, err
	}
	if err := json.Unmarshal(plain, &d); err != nil {
		return d, err
	}
	return d, nil
}

// signState HMACs the return path together with the flow nonce, so neither can
// be tampered with at the callback; the whole thing is base64'd for safe
// transport through WorkOS. Layout (before the outer base64): returnTo|nonce|sig.
// nonce and sig are base64url (no "|"), so the callback splits from the right,
// which keeps a "|" inside returnTo intact.
func (m *Manager) signState(returnTo, nonce string) string {
	payload := returnTo + "|" + nonce + "|" + base64.RawURLEncoding.EncodeToString(m.stateMAC(returnTo, nonce))
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func (m *Manager) verifyState(state string) (returnTo, nonce string, ok bool) {
	raw, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "", "", false
	}
	s := string(raw)
	sigAt := strings.LastIndex(s, "|")
	if sigAt < 0 {
		return "", "", false
	}
	nonceAt := strings.LastIndex(s[:sigAt], "|")
	if nonceAt < 0 {
		return "", "", false
	}
	returnTo, nonce, sigB64 := s[:nonceAt], s[nonceAt+1:sigAt], s[sigAt+1:]
	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return "", "", false
	}
	if !hmac.Equal(sig, m.stateMAC(returnTo, nonce)) {
		return "", "", false
	}
	return returnTo, nonce, true
}

// stateMAC is the HMAC over returnTo and nonce, domain-separated by a NUL so
// distinct (returnTo, nonce) pairs can't collide by shifting the boundary.
func (m *Manager) stateMAC(returnTo, nonce string) []byte {
	mac := hmac.New(sha256.New, m.stateKey)
	mac.Write([]byte(returnTo))
	mac.Write([]byte{0})
	mac.Write([]byte(nonce))
	return mac.Sum(nil)
}

// randomToken returns 32 bytes of URL-safe base64 randomness, used for the
// one-time OAuth nonce.
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// refreshOnce runs the WorkOS refresh through singleflight so that concurrent
// requests for the same refresh token trigger exactly one exchange and share
// its result.
func (m *Manager) refreshOnce(ctx context.Context, refreshToken string) (access, newRefresh string, err error) {
	key := refreshKey(refreshToken)
	v, err, _ := m.refreshGroup.Do(key, func() (any, error) {
		a, r, e := m.refresh(ctx, refreshToken)
		if e != nil {
			return nil, e
		}
		return [2]string{a, r}, nil
	})
	if err != nil {
		return "", "", err
	}
	pair := v.([2]string)
	return pair[0], pair[1], nil
}

// refreshKey derives a stable singleflight key from a refresh token without
// keeping the raw secret around as a map key.
func refreshKey(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (m *Manager) setCookie(c echo.Context, sealed string) {
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     cookieName,
		Value:    sealed,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(cookieMaxAge.Seconds()),
	})
}

func (m *Manager) clearCookie(c echo.Context) {
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// setNonceCookie / clearNonceCookie manage the one-time CSRF nonce. It is scoped
// to /auth so it never rides along with page or /api requests, and SameSite=Lax
// still lets it return on the top-level GET redirect from WorkOS to /auth/callback.
func (m *Manager) setNonceCookie(c echo.Context, nonce string) {
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     nonceCookieName,
		Value:    nonce,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(nonceMaxAge.Seconds()),
	})
}

func (m *Manager) clearNonceCookie(c echo.Context) {
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     nonceCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// safeReturnTo accepts only a same-origin, rooted relative path, mirroring the
// web AuthProvider's safeReturnTo. Anything absolute, protocol-relative, or
// malformed collapses to "/", which blocks open redirects through returnTo.
func safeReturnTo(raw string) string {
	if raw == "" || !strings.HasPrefix(raw, "/") {
		return "/"
	}
	// "//host" is protocol-relative; "/\" is a browser-normalized variant.
	if strings.HasPrefix(raw, "//") || strings.HasPrefix(raw, "/\\") {
		return "/"
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "" || u.Host != "" {
		return "/"
	}
	out := u.EscapedPath()
	if out == "" {
		out = "/"
	}
	if u.RawQuery != "" {
		out += "?" + u.RawQuery
	}
	// Preserve the fragment (e.g. an on-page anchor) — browsers honor it in a
	// redirect Location, so keeping it round-trips the user's exact spot.
	if u.Fragment != "" {
		out += "#" + u.EscapedFragment()
	}
	return out
}

// noStore keeps sessioned responses out of any shared cache (Cloudflare sits in
// front of this origin), so one user's /api/me or Set-Cookie can never be served
// to another. Mirrors the no-cache handling the static handler already applies
// to service-worker scripts.
func noStore(c echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-store")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
