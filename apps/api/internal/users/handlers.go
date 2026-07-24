package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/digest"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/session"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

// renameCooldown is how long a user must wait between username changes. It
// doubles as anti-squatting — a claimed handle cannot be cycled rapidly. Timezone
// changes are unlimited.
const renameCooldown = 30 * 24 * time.Hour

// Handlers exposes the account routes (/api/me and /api/settings*) over the
// store. Construct with New and mount on a session-gated group with Register.
// poster is the admin-channel Slack webhook and may be nil (the ping is optional;
// a deletion request still records its durable row). Mirrors the shape of
// ingest.Handlers.
type Handlers struct {
	store  *store.Store
	poster digest.WebhookPoster
}

// New builds the account Handlers over an already-open store and an optional
// admin-channel webhook poster.
func New(st *store.Store, poster digest.WebhookPoster) *Handlers {
	return &Handlers{store: st, poster: poster}
}

// Register mounts the account routes on api, which MUST already be gated by the
// session middleware (each handler reads session.User). GET routes accept HEAD
// too; the mutating routes additionally pass through originCheck, a same-origin
// guard backing up the session cookie's SameSite=Lax.
func (h *Handlers) Register(api *echo.Group) {
	getOrHead := []string{http.MethodGet, http.MethodHead}
	api.Match(getOrHead, "/me", h.Me)
	api.Match(getOrHead, "/settings", h.GetSettings)
	api.PATCH("/settings", h.PatchSettings, originCheck)
	api.POST("/settings/delete-request", h.DeleteRequest, originCheck)
}

// Upsert creates or refreshes the account row for a signed-in WorkOS profile,
// generating a first username for a brand-new user. Returns the stored row and
// whether it was newly created. main wires this as session.OnLogin; the handlers
// reuse it to self-heal a row a failed login never created.
func Upsert(ctx context.Context, st *store.Store, p session.Profile) (store.User, bool, error) {
	return st.UpsertUserByWorkOSID(ctx, store.User{
		WorkOSID:  p.ID,
		Email:     p.Email,
		Username:  Generate(p.FirstName, p.LastName, p.Email),
		AvatarURL: p.ProfilePictureURL,
	})
}

type meResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Timezone string `json:"timezone"`
}

// Me returns the signed-in user's identity joined with the account row — a single
// indexed lookup per page load, with the DB as the source of truth for the
// username, so a rename is never stale. If the row is somehow absent (a login
// whose OnLogin upsert failed), it degrades to the session profile rather than
// erroring, so the navbar still renders.
func (h *Handlers) Me(c echo.Context) error {
	p, ok := session.User(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	resp := meResponse{Email: p.Email, Avatar: p.ProfilePictureURL}
	u, err := h.store.GetUserByWorkOSID(c.Request().Context(), p.ID)
	switch {
	case err == nil:
		resp.ID, resp.Email, resp.Username = u.ID, u.Email, u.Username
		resp.Avatar, resp.Timezone = u.AvatarURL, u.Timezone
	case errors.Is(err, sql.ErrNoRows):
		// degrade to the session profile (username stays empty)
	default:
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

type settingsResponse struct {
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	Timezone            string     `json:"timezone"`
	DeletionRequestedAt *time.Time `json:"deletion_requested_at"`
}

// GetSettings returns the caller's editable settings plus any pending deletion
// request (so the form can show the "deletion pending" state).
func (h *Handlers) GetSettings(c echo.Context) error {
	p, ok := session.User(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	u, err := h.userRow(c.Request().Context(), p)
	if err != nil {
		return err
	}
	return h.writeSettings(c, u)
}

type patchRequest struct {
	Username *string `json:"username"`
	Timezone *string `json:"timezone"`
}

// PatchSettings updates the username and/or timezone. Each field is optional
// (a nil pointer leaves it untouched). Status codes the web form maps to inline
// messages: 400 invalid/reserved username or invalid timezone, 409 taken, 429
// within the rename cooldown (with retry_after_days), 200 with the new settings.
func (h *Handlers) PatchSettings(c echo.Context) error {
	p, ok := session.User(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	var req patchRequest
	if err := c.Bind(&req); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid_body")
	}
	ctx := c.Request().Context()
	u, err := h.userRow(ctx, p)
	if err != nil {
		return err
	}

	// Timezone: validated against the tz database so an arbitrary string can't be
	// stored. Unlimited.
	if req.Timezone != nil && *req.Timezone != u.Timezone {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			return errJSON(c, http.StatusBadRequest, "timezone_invalid")
		}
		if err := h.store.SetTimezone(ctx, u.ID, *req.Timezone); err != nil {
			return err
		}
	}

	// Username: format + reserved validation, the 30-day rename cooldown, then the
	// unique-index collision mapped to 409.
	if req.Username != nil && *req.Username != u.Username {
		switch err := Validate(*req.Username); {
		case errors.Is(err, ErrUsernameReserved):
			return errJSON(c, http.StatusBadRequest, "username_reserved")
		case err != nil:
			return errJSON(c, http.StatusBadRequest, "username_invalid")
		}
		if u.UsernameChangedAt != nil {
			if elapsed := h.store.Now().UTC().Sub(*u.UsernameChangedAt); elapsed < renameCooldown {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error":            "rate_limited",
					"retry_after_days": daysUntil(renameCooldown - elapsed),
				})
			}
		}
		if err := h.store.SetUsername(ctx, u.ID, *req.Username); err != nil {
			if errors.Is(err, store.ErrUsernameTaken) {
				return errJSON(c, http.StatusConflict, "username_taken")
			}
			return err
		}
	}

	// Reflect the persisted state back.
	u, err = h.store.GetUserByWorkOSID(ctx, p.ID)
	if err != nil {
		return err
	}
	return h.writeSettings(c, u)
}

type deleteRequestBody struct {
	Reason string `json:"reason"`
}

type deleteResponse struct {
	DeletionRequestedAt *time.Time `json:"deletion_requested_at"`
}

// DeleteRequest records a pending account-deletion request (idempotent — a second
// call never files a duplicate) and best-effort pings the admin Slack channel.
// Users never self-delete; an admin actions the row manually. A webhook failure
// is logged but still returns 202, because the row is the durable record.
func (h *Handlers) DeleteRequest(c echo.Context) error {
	p, ok := session.User(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	ctx := c.Request().Context()
	u, err := h.userRow(ctx, p)
	if err != nil {
		return err
	}

	var body deleteRequestBody
	_ = c.Bind(&body) // reason is optional; an empty/absent body just means none

	created, err := h.store.CreateDeletionRequest(ctx, u.ID, body.Reason)
	if err != nil {
		return err
	}
	if created {
		h.notifyAdmin(ctx, c, u)
	}

	dr, err := h.store.PendingDeletionRequest(ctx, u.ID)
	if err != nil {
		return err
	}
	resp := deleteResponse{}
	if dr != nil {
		resp.DeletionRequestedAt = &dr.RequestedAt
	}
	return c.JSON(http.StatusAccepted, resp)
}

// writeSettings serializes a user's current settings, including any pending
// deletion request.
func (h *Handlers) writeSettings(c echo.Context, u store.User) error {
	resp := settingsResponse{Username: u.Username, Email: u.Email, Timezone: u.Timezone}
	dr, err := h.store.PendingDeletionRequest(c.Request().Context(), u.ID)
	if err != nil {
		return err
	}
	if dr != nil {
		resp.DeletionRequestedAt = &dr.RequestedAt
	}
	return c.JSON(http.StatusOK, resp)
}

// userRow fetches the caller's account row, self-healing a missing one by
// upserting from the session profile — the row a login whose OnLogin upsert
// didn't land would otherwise lack.
func (h *Handlers) userRow(ctx context.Context, p session.Profile) (store.User, error) {
	u, err := h.store.GetUserByWorkOSID(ctx, p.ID)
	if errors.Is(err, sql.ErrNoRows) {
		u, _, err = Upsert(ctx, h.store, p)
	}
	return u, err
}

// notifyAdmin posts a best-effort Slack message about a new deletion request. A
// nil poster (webhook unset) or a post failure is logged and swallowed.
func (h *Handlers) notifyAdmin(ctx context.Context, c echo.Context, u store.User) {
	if h.poster == nil {
		return
	}
	payload, err := json.Marshal(map[string]string{
		"text": fmt.Sprintf("Account deletion requested by %s (%s), user id %d — action manually.", u.Username, u.Email, u.ID),
	})
	if err != nil {
		c.Logger().Errorf("delete-request: marshal slack payload: %v", err)
		return
	}
	if err := h.poster.Post(ctx, payload); err != nil {
		c.Logger().Errorf("delete-request: slack notify: %v", err)
	}
}

// originCheck rejects a cross-origin state-changing request. The session cookie is
// SameSite=Lax, which already stops a cross-site POST/PATCH from carrying it (so
// this is defense in depth), but adding the check means every future write
// endpoint mounted here inherits it. It trusts Sec-Fetch-Site when the browser
// sends it, otherwise compares the Origin header's host to the request host; a
// request with neither header (a same-origin server call, curl, a test) is
// allowed.
func originCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		if sfs := r.Header.Get("Sec-Fetch-Site"); sfs != "" {
			if sfs != "same-origin" && sfs != "none" {
				return echo.NewHTTPError(http.StatusForbidden, "cross-origin request rejected")
			}
			return next(c)
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || u.Host != r.Host {
				return echo.NewHTTPError(http.StatusForbidden, "cross-origin request rejected")
			}
		}
		return next(c)
	}
}

// errJSON writes a structured error the web form maps to a message.
func errJSON(c echo.Context, status int, code string) error {
	return c.JSON(status, map[string]string{"error": code})
}

// daysUntil rounds a remaining duration up to whole days, so a user 29.1 days
// into the cooldown is told "1 day", never "0".
func daysUntil(d time.Duration) int {
	days := int(d / (24 * time.Hour))
	if d%(24*time.Hour) > 0 {
		days++
	}
	return days
}
