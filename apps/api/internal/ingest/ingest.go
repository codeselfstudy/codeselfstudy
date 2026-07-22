package ingest

import (
	"crypto/subtle"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/digest"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/extract"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/mailparse"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

// defaultMaxIngestBytes caps the raw email body accepted at /api/ingest.
// Cloudflare Email Routing tops out well under this; the Worker also guards on
// size.
const defaultMaxIngestBytes = 25 * 1024 * 1024

// Handlers holds the email-ingest pipeline dependencies and exposes the Echo
// handlers + middleware. Construct with New and mount with Register.
type Handlers struct {
	cfg       Config
	store     *store.Store
	extractor extract.Extractor
	poster    digest.WebhookPoster

	// MaxIngestBytes bounds the /api/ingest body; overridable in tests.
	MaxIngestBytes int64
}

// New builds the ingest Handlers over an already-open store, an extractor, and a
// Slack poster.
func New(cfg Config, st *store.Store, extractor extract.Extractor, poster digest.WebhookPoster) *Handlers {
	return &Handlers{
		cfg:            cfg,
		store:          st,
		extractor:      extractor,
		poster:         poster,
		MaxIngestBytes: defaultMaxIngestBytes,
	}
}

// Register mounts the ingest routes on e, each behind the bearer-token
// middleware. Both are POST, so they never collide with the GET/HEAD static
// catch-all; mount before the /api/* JSON-404 reservation so the concrete paths
// win over the wildcard.
func (h *Handlers) Register(e *echo.Echo) {
	e.POST("/api/ingest", h.Ingest, h.BearerAuth)
	e.POST("/api/admin/digest", h.AdminDigest, h.BearerAuth)
}

// BearerAuth checks the Authorization header against the configured token using
// a constant-time comparison. An unset token rejects everything.
func (h *Handlers) BearerAuth(next echo.HandlerFunc) echo.HandlerFunc {
	const prefix = "Bearer "
	return func(c echo.Context) error {
		hdr := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(hdr, prefix) {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
		}
		token := strings.TrimPrefix(hdr, prefix)
		if h.cfg.IngestToken == "" ||
			subtle.ConstantTimeCompare([]byte(token), []byte(h.cfg.IngestToken)) != 1 {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
		return next(c)
	}
}

type ingestResult struct {
	EmailID        int64 `json:"email_id"`
	DealsExtracted int   `json:"deals_extracted"`
	DigestPosted   bool  `json:"digest_posted"`
	Duplicate      bool  `json:"duplicate,omitempty"`
}

// Ingest parses a raw email, stores it, extracts deals, and (best-effort) posts
// a digest. It is idempotent: a re-POST of an already-extracted email does
// nothing.
func (h *Handlers) Ingest(c echo.Context) error {
	ctx := c.Request().Context()

	raw, err := io.ReadAll(http.MaxBytesReader(c.Response(), c.Request().Body, h.MaxIngestBytes))
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "email too large")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "read body")
	}

	email, err := mailparse.Parse(raw)
	if err != nil {
		c.Logger().Errorf("parse email: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email")
	}

	stored, isNew, err := h.store.InsertEmail(ctx, store.Email{
		MessageID: email.MessageID,
		From:      email.From,
		To:        email.To,
		Subject:   email.Subject,
		SentAt:    email.SentAt,
		BodyText:  email.Text,
	})
	if err != nil {
		return err
	}

	// Idempotent re-POST: an already-extracted email is done. A prior
	// extract_failed (or an interrupted attempt) falls through and retries.
	if !isNew && stored.Status == store.StatusExtracted {
		return c.JSON(http.StatusOK, ingestResult{EmailID: stored.ID, Duplicate: true})
	}

	deals, err := h.extractor.Extract(ctx, email)
	if err != nil {
		_ = h.store.SetEmailStatus(ctx, stored.ID, store.StatusExtractFailed, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "extract deals")
	}

	repostCutoff := h.store.Now().Add(-h.cfg.RepostAfter)
	for _, d := range deals {
		if err := h.store.UpsertDeal(ctx, store.Deal{
			EmailID:     stored.ID,
			DedupeKey:   store.DedupeKey(email.From, d.Title),
			Source:      d.Source,
			Title:       d.Title,
			URL:         d.URL,
			Price:       d.Price,
			EndsAt:      d.EndsAt,
			Description: d.Description,
		}, repostCutoff); err != nil {
			return err
		}
	}
	if err := h.store.SetEmailStatus(ctx, stored.ID, store.StatusExtracted, ""); err != nil {
		return err
	}

	// The email is safely ingested; a digest failure must not fail the request
	// (the deals are queued and a later ingest retries).
	posted, derr := digest.Run(ctx, h.store, h.poster, h.cfg.DigestInterval, store.DefaultStaleWindow, false)
	if derr != nil {
		c.Logger().Errorf("digest run: %v", derr)
	}

	return c.JSON(http.StatusOK, ingestResult{
		EmailID:        stored.ID,
		DealsExtracted: len(deals),
		DigestPosted:   posted,
	})
}

// AdminDigest forces a digest post (skips the interval check, still race-safe).
func (h *Handlers) AdminDigest(c echo.Context) error {
	posted, err := digest.Run(c.Request().Context(), h.store, h.poster, h.cfg.DigestInterval, store.DefaultStaleWindow, true)
	if err != nil {
		// Do not echo err: it can contain the Slack webhook URL (a secret).
		c.Logger().Errorf("admin digest: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "digest failed")
	}
	return c.JSON(http.StatusOK, map[string]bool{"posted": posted})
}
