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

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/auth"
)

// cookieName is the session cookie. contextUserKey is where the authenticated
// profile is stashed for downstream handlers (HandleMe).
const (
	cookieName     = "wos_session"
	contextUserKey = "session:user"
	cookieMaxAge   = 30 * 24 * time.Hour
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
// OAuth state so the callback can trust it.
func (m *Manager) Login(c echo.Context) error {
	noStore(c)
	returnTo := safeReturnTo(c.QueryParam("returnTo"))
	authURL, err := usermanagement.GetAuthorizationURL(usermanagement.GetAuthorizationURLOpts{
		ClientID:    m.clientID,
		RedirectURI: m.callbackURL,
		Provider:    "authkit",
		State:       m.signState(returnTo),
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
	if c.QueryParam("error") != "" {
		return c.Redirect(http.StatusFound, "/")
	}
	code := c.QueryParam("code")
	returnTo, ok := m.verifyState(c.QueryParam("state"))
	if code == "" || !ok {
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
				access, refresh, rerr := m.refresh(ctx, data.RefreshToken)
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

// signState HMACs the return path so a tampered state is rejected at the
// callback; the whole thing is base64'd for safe transport through WorkOS.
func (m *Manager) signState(returnTo string) string {
	mac := hmac.New(sha256.New, m.stateKey)
	mac.Write([]byte(returnTo))
	payload := returnTo + "|" + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func (m *Manager) verifyState(state string) (string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "", false
	}
	i := strings.LastIndex(string(raw), "|")
	if i < 0 {
		return "", false
	}
	returnTo, sigB64 := string(raw)[:i], string(raw)[i+1:]
	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, m.stateKey)
	mac.Write([]byte(returnTo))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return "", false
	}
	return returnTo, true
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
