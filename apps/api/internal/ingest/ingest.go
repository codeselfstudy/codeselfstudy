package ingest

import (
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/digest"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/expiry"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/extract"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/mailparse"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

// defaultMaxIngestBytes caps the raw email body accepted at /api/ingest.
// Cloudflare Email Routing tops out well under this; the Worker also guards on
// size.
const defaultMaxIngestBytes = 25 * 1024 * 1024

// resolveBudget bounds URL resolution and page mining for one email's whole
// batch of deals, so a pathological email full of slow links cannot stall the
// ingest request. It scales with the batch — a page fetch costs real time per
// deal — but stays capped so /api/ingest remains responsive. Deals that miss
// the budget keep their extracted values.
func resolveBudget(deals int) time.Duration {
	const (
		floor   = 15 * time.Second
		perDeal = 3 * time.Second
		ceiling = 45 * time.Second
	)
	budget := floor + time.Duration(deals)*perDeal
	if budget > ceiling {
		return ceiling
	}
	return budget
}

// URLResolver cleans one deal URL (following tracking redirects, stripping
// tracking parameters). Implementations must return a usable URL — the input
// unchanged on any failure — plus the failure for logging. ResolvePage
// additionally returns the destination page (nil when skipped or failed) so
// the caller can mine it. See internal/resolve.
type URLResolver interface {
	Resolve(ctx context.Context, rawURL string) (string, error)
	ResolvePage(ctx context.Context, rawURL string) (string, []byte, error)
}

// Handlers holds the email-ingest pipeline dependencies and exposes the Echo
// handlers + middleware. Construct with New and mount with Register.
type Handlers struct {
	cfg       Config
	store     *store.Store
	extractor extract.Extractor
	poster    digest.WebhookPoster

	// MaxIngestBytes bounds the /api/ingest body; overridable in tests.
	MaxIngestBytes int64

	// Resolver, when set, cleans each extracted deal URL before it is stored.
	// nil skips resolution (tests, or a deliberately offline setup).
	Resolver URLResolver
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
	// Forced reports that the sender was on the APPROVED_FORWARDING_EMAILS
	// allowlist, so the digest skipped the interval wait.
	Forced bool `json:"forced,omitempty"`
}

// Ingest parses a raw email, stores it, extracts deals, and (best-effort) posts
// a digest. It is idempotent: a re-POST of an already-extracted email does
// nothing. Mail from an approved sender forces the digest, posting queued deals
// immediately instead of waiting out DigestInterval.
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

	// An extracted deadline already past on the email's own send date (the
	// extractor interprets deadlines relative to it; fall back to the clock
	// for emails without a parseable Date) carries a stale or guessed year
	// — vendors ship last year's copy, extractors guess a year for yearless
	// dates — not a real deadline. Drop it before resolution — even with no
	// resolver wired an impossible date must not be stored — and remember the
	// drop so the upsert below can erase a previously stored copy of the same
	// bad date instead of coalescing it back.
	ref := h.store.Now()
	if email.SentAt != nil {
		ref = *email.SentAt
	}
	cleared := make([]bool, len(deals))
	for i := range deals {
		d := &deals[i]
		if d.EndsAt != "" && !expiry.OnOrAfter(d.EndsAt, ref) {
			c.Logger().Warnf("dropping implausible deal deadline %q (email date %s)", d.EndsAt, ref.Format("2006-01-02"))
			d.EndsAt = ""
			cleared[i] = true
		}
	}

	// Resolve tracking-redirect URLs to their clean destinations before
	// storing (see internal/resolve). When the email stated no usable
	// deadline, the same fetch also brings back the deal page so its
	// structured data can fill ends_at (see internal/expiry) — an email-stated
	// deadline is never overwritten. Strictly best-effort under one shared
	// budget: a failed or skipped step keeps the extracted values, and a
	// resolver problem never fails the ingest.
	if h.Resolver != nil && len(deals) > 0 {
		rctx, cancel := context.WithTimeout(ctx, resolveBudget(len(deals)))
		for i := range deals {
			d := &deals[i]
			if d.EndsAt != "" {
				resolved, rerr := h.Resolver.Resolve(rctx, d.URL)
				if rerr != nil {
					c.Logger().Warnf("resolve deal url: %v", rerr)
				}
				d.URL = resolved
				continue
			}
			resolved, page, rerr := h.Resolver.ResolvePage(rctx, d.URL)
			if rerr != nil {
				c.Logger().Warnf("resolve deal url: %v", rerr)
			}
			d.URL = resolved
			if len(page) > 0 {
				// The page's structured data gets the same skepticism as the
				// extractor: shop pages routinely carry a stale
				// priceValidUntil from an earlier promotion, and a past date
				// is useless to store (the digest would hide it anyway).
				// Leaving it empty also keeps ClearEndsAt effective, so a
				// previously stored bad deadline still gets erased.
				if candidate := expiry.FromHTML(page); candidate == "" || expiry.OnOrAfter(candidate, ref) {
					d.EndsAt = candidate
				} else {
					c.Logger().Warnf("dropping implausible page deadline %q (email date %s)", candidate, ref.Format("2006-01-02"))
				}
			}
		}
		cancel()
	}

	repostCutoff := h.store.Now().Add(-h.cfg.RepostAfter)
	for i, d := range deals {
		if err := h.store.UpsertDeal(ctx, store.Deal{
			EmailID:     stored.ID,
			DedupeKey:   store.DedupeKey(email.From, d.Title),
			Source:      d.Source,
			Title:       d.Title,
			URL:         d.URL,
			Price:       d.Price,
			EndsAt:      d.EndsAt,
			Description: d.Description,
			// A rejected deadline that nothing refilled must overwrite a
			// stored one — the stored copy is the same bad date.
			ClearEndsAt: cleared[i] && d.EndsAt == "",
		}, repostCutoff); err != nil {
			return err
		}
	}
	if err := h.store.SetEmailStatus(ctx, stored.ID, store.StatusExtracted, ""); err != nil {
		return err
	}

	// An approved sender skips the once-per-interval wait. Forcing only bypasses
	// the interval check — the stale-claim check still runs, so this stays
	// race-safe (see digest.Run).
	force := h.cfg.IsApprovedSender(email.From, c.Request().Header.Get("X-Envelope-From"))

	// The email is safely ingested; a digest failure must not fail the request
	// (the deals are queued and a later ingest retries).
	posted, derr := digest.Run(ctx, h.store, h.poster, h.cfg.DigestInterval, store.DefaultStaleWindow, force)
	if derr != nil {
		c.Logger().Errorf("digest run: %v", derr)
	}

	return c.JSON(http.StatusOK, ingestResult{
		EmailID:        stored.ID,
		DealsExtracted: len(deals),
		DigestPosted:   posted,
		Forced:         force,
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
