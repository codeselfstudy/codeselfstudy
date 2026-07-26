// Package store owns the SQL for the deal-digest pipeline — every statement,
// including the correctness core of the whole app, the atomic once-per-interval
// digest claim.
//
// The schema lives in the internal/db goose migrations; a Store is constructed
// over an already-opened, already-migrated *sql.DB (see internal/db.Open and
// internal/db.Migrate), so the same handle is shared with the rest of the
// server. It is a concrete type (no interface): tests run the real store against
// a temporary SQLite file via the pure-Go modernc.org/sqlite driver, so the real
// store is its own test double. The Go backend currently supports local SQLite
// only; remote Turso (libsql://) support is added in a later change, where
// internal/db.Open picks the driver by URL scheme.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/publicsuffix"
)

// Email status values.
const (
	StatusReceived      = "received"
	StatusExtracted     = "extracted"
	StatusExtractFailed = "extract_failed"
)

// Digest status values.
const (
	DigestClaimed = "claimed"
	DigestPosted  = "posted"
	DigestFailed  = "failed"
)

// DefaultStaleWindow bounds how long a claimed-but-not-yet-completed digest
// blocks new claims (see ClaimDigest). It MUST exceed the worst-case time
// between winning a claim and calling MarkDigestPosted/MarkDigestFailed; if a
// claimant takes longer, a second caller could take over a still-in-flight claim
// and post a duplicate digest. Posting is a single synchronous Slack request, so
// five minutes is comfortably safe.
const DefaultStaleWindow = 5 * time.Minute

// Email is a stored inbound message. Only headers and normalized text are
// kept — the raw MIME is not persisted.
type Email struct {
	ID           int64
	MessageID    string
	From         string
	To           string
	Subject      string
	SentAt       *time.Time
	ReceivedAt   time.Time
	BodyText     string
	Status       string
	ExtractError string
}

// Deal is one extracted offer. Optional text fields are empty when absent;
// they map to SQL NULL so upserts can preserve an earlier non-empty value.
type Deal struct {
	ID               int64
	EmailID          int64
	DedupeKey        string
	Source           string
	Title            string
	URL              string
	Price            string
	EndsAt           string
	Description      string
	FirstSeenAt      time.Time
	LastSeenAt       time.Time
	SeenCount        int
	PostedInDigestID *int64

	// ClearEndsAt is a write-only flag for UpsertDeal (never read back from
	// the database): when set, an empty EndsAt overwrites the stored deadline
	// with NULL instead of preserving it. Ingest sets it after rejecting an
	// implausible extracted deadline, so a previously stored copy of the same
	// bad date cannot survive a re-sighting.
	ClearEndsAt bool
}

// Store is a concrete persistence layer over a *sql.DB.
//
// internal/db.Open caps the SQLite pool at one connection (SetMaxOpenConns(1)),
// so a method that has opened a transaction MUST route every further statement
// through that tx — issuing an s.db call while a tx is open would wait forever
// for the single connection the tx already holds. MarkDigestPosted is the one
// tx site today and follows this rule.
type Store struct {
	db *sql.DB
	// Now supplies the current time; overridable in tests for deterministic
	// claim-window behavior. Defaults to time.Now.
	Now func() time.Time
}

// New returns a Store over db, which must already be open and migrated (see
// internal/db.Open + internal/db.Migrate). The caller owns db's lifecycle.
func New(db *sql.DB) *Store {
	return &Store{db: db, Now: time.Now}
}

// InsertEmail inserts e (status "received", received_at set to now). If an email
// with the same message_id already exists, it is not modified; the existing row
// is returned with isNew=false. This makes re-POSTs of the same email idempotent.
func (s *Store) InsertEmail(ctx context.Context, e Email) (stored Email, isNew bool, err error) {
	now := s.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO emails (message_id, from_addr, to_addr, subject, sent_at, received_at, body_text, status)
		 VALUES (?,?,?,?,?,?,?,?)
		 ON CONFLICT(message_id) DO NOTHING`,
		e.MessageID, e.From, e.To, e.Subject, nullTime(e.SentAt), formatTime(now), e.BodyText, StatusReceived,
	)
	if err != nil {
		return Email{}, false, fmt.Errorf("insert email: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Email{}, false, fmt.Errorf("insert email rows: %w", err)
	}
	if n == 1 {
		id, err := res.LastInsertId()
		if err != nil {
			return Email{}, false, fmt.Errorf("insert email id: %w", err)
		}
		e.ID = id
		e.ReceivedAt = now
		e.Status = StatusReceived
		return e, true, nil
	}
	existing, err := s.GetEmailByMessageID(ctx, e.MessageID)
	if err != nil {
		return Email{}, false, err
	}
	return existing, false, nil
}

// GetEmailByMessageID returns the email with the given message_id, or
// sql.ErrNoRows if none exists.
func (s *Store) GetEmailByMessageID(ctx context.Context, messageID string) (Email, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, message_id, from_addr, to_addr, subject, sent_at, received_at, body_text, status, extract_error
		 FROM emails WHERE message_id = ?`, messageID)
	return scanEmail(row)
}

// SetEmailStatus updates an email's status and extract_error (pass "" to clear).
func (s *Store) SetEmailStatus(ctx context.Context, id int64, status, extractErr string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE emails SET status = ?, extract_error = ? WHERE id = ?`,
		status, nullString(extractErr), id)
	if err != nil {
		return fmt.Errorf("set email status: %w", err)
	}
	return nil
}

// UpsertDeal inserts a deal or, when its dedupe_key already exists, bumps
// last_seen_at and seen_count and refreshes each optional field from the new
// sighting when it supplies a value (the stored value is kept only when the
// new one is empty). d.ClearEndsAt is the exception: it makes an empty EndsAt
// overwrite the stored deadline with NULL rather than preserve it.
// If the existing deal was last seen before repostCutoff (i.e. unseen for longer
// than the repost window), its posted_in_digest_id is reset to NULL so the deal
// re-enters the next digest.
func (s *Store) UpsertDeal(ctx context.Context, d Deal, repostCutoff time.Time) error {
	now := formatTime(s.Now().UTC())
	// The driver may not convert Go bools; bind the flag as 0/1 explicitly.
	clearEndsAt := 0
	if d.ClearEndsAt {
		clearEndsAt = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO deals
		   (email_id, dedupe_key, source, title, url, price, ends_at, description, first_seen_at, last_seen_at, seen_count)
		 VALUES (?,?,?,?,?,?,?,?,?,?,1)
		 ON CONFLICT(dedupe_key) DO UPDATE SET
		   last_seen_at        = excluded.last_seen_at,
		   seen_count          = deals.seen_count + 1,
		   url                 = COALESCE(excluded.url, deals.url),
		   price               = COALESCE(excluded.price, deals.price),
		   ends_at             = CASE WHEN ? THEN NULL
		                              ELSE COALESCE(excluded.ends_at, deals.ends_at) END,
		   description         = COALESCE(excluded.description, deals.description),
		   posted_in_digest_id = CASE WHEN deals.last_seen_at < ?
		                              THEN NULL ELSE deals.posted_in_digest_id END`,
		d.EmailID, d.DedupeKey, d.Source, d.Title,
		nullString(d.URL), nullString(d.Price), nullString(d.EndsAt), nullString(d.Description),
		now, now, clearEndsAt, formatTime(repostCutoff.UTC()),
	)
	if err != nil {
		return fmt.Errorf("upsert deal: %w", err)
	}
	return nil
}

// ClaimDigest atomically reserves the digest slot. It succeeds (ok=true, with the
// new digest id) only when no digest was posted within interval and no live claim
// exists within staleWindow. When force is true the interval check is skipped
// (manual /admin/digest trigger) but the stale-claim check is kept, so it remains
// race-safe. Ownership is decided by the database, not the caller: the single
// INSERT…WHERE NOT EXISTS is atomic on SQLite (serialized via SetMaxOpenConns(1))
// and on Turso (server-side write serialization), so at most one caller inserts.
//
// staleWindow must exceed the worst-case claim→complete latency — see
// DefaultStaleWindow.
func (s *Store) ClaimDigest(ctx context.Context, interval, staleWindow time.Duration, force bool) (id int64, ok bool, err error) {
	now := s.Now().UTC()
	claimedAt := formatTime(now)
	staleCutoff := formatTime(now.Add(-staleWindow))

	var res sql.Result
	if force {
		res, err = s.db.ExecContext(ctx,
			`INSERT INTO digests (claimed_at, status)
			 SELECT ?, 'claimed'
			 WHERE NOT EXISTS (
			   SELECT 1 FROM digests WHERE status = 'claimed' AND claimed_at > ?)`,
			claimedAt, staleCutoff)
	} else {
		intervalCutoff := formatTime(now.Add(-interval))
		res, err = s.db.ExecContext(ctx,
			`INSERT INTO digests (claimed_at, status)
			 SELECT ?, 'claimed'
			 WHERE NOT EXISTS (
			   SELECT 1 FROM digests
			   WHERE (status = 'posted'  AND posted_at  > ?)
			      OR (status = 'claimed' AND claimed_at > ?))`,
			claimedAt, intervalCutoff, staleCutoff)
	}
	if err != nil {
		return 0, false, fmt.Errorf("claim digest: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, false, fmt.Errorf("claim digest rows: %w", err)
	}
	if n != 1 {
		return 0, false, nil
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, false, fmt.Errorf("claim digest id: %w", err)
	}
	return id, true, nil
}

// UnpostedDeals returns deals not yet included in any digest, oldest first by
// first_seen_at, capped at limit (limit <= 0 means no cap). A deal re-queued by
// the repost window keeps its original first_seen_at, so a resurfaced offer sorts
// ahead of newer ones — intended, since it is genuinely the older deal.
func (s *Store) UnpostedDeals(ctx context.Context, limit int) ([]Deal, error) {
	query := `SELECT id, email_id, dedupe_key, source, title, url, price, ends_at, description,
	                 first_seen_at, last_seen_at, seen_count, posted_in_digest_id
	          FROM deals WHERE posted_in_digest_id IS NULL ORDER BY first_seen_at`
	args := []any{}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unposted deals: %w", err)
	}
	defer rows.Close()

	var deals []Deal
	for rows.Next() {
		d, err := scanDeal(rows)
		if err != nil {
			return nil, err
		}
		deals = append(deals, d)
	}
	return deals, rows.Err()
}

// GetDealByDedupeKey returns the deal with the given dedupe_key, or
// sql.ErrNoRows if none exists.
func (s *Store) GetDealByDedupeKey(ctx context.Context, key string) (Deal, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, email_id, dedupe_key, source, title, url, price, ends_at, description,
		        first_seen_at, last_seen_at, seen_count, posted_in_digest_id
		 FROM deals WHERE dedupe_key = ?`, key)
	return scanDeal(row)
}

// DeleteDigest removes a digest row. Used to release a claim that produced no
// deals, so an empty run does not suppress the next interval.
func (s *Store) DeleteDigest(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM digests WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete digest: %w", err)
	}
	return nil
}

// MarkDigestPosted marks a digest posted and stamps the given deals as included
// in it, in a single transaction.
func (s *Store) MarkDigestPosted(ctx context.Context, digestID int64, dealIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE digests SET status = 'posted', posted_at = ?, deal_count = ? WHERE id = ?`,
		formatTime(s.Now().UTC()), len(dealIDs), digestID); err != nil {
		return fmt.Errorf("mark digest posted: %w", err)
	}
	if len(dealIDs) > 0 {
		// One statement rather than N — matters over the network to Turso.
		placeholders := make([]string, len(dealIDs))
		args := make([]any, 0, len(dealIDs)+1)
		args = append(args, digestID)
		for i, id := range dealIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		q := `UPDATE deals SET posted_in_digest_id = ? WHERE id IN (` + strings.Join(placeholders, ",") + `)`
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return fmt.Errorf("stamp deals: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// MarkDigestFailed marks a digest failed. Its deals stay unposted and a later
// ingest retries; failed rows never block a future claim.
func (s *Store) MarkDigestFailed(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx,
		`UPDATE digests SET status = 'failed' WHERE id = ?`, id); err != nil {
		return fmt.Errorf("mark digest failed: %w", err)
	}
	return nil
}

// DedupeKey builds the stable identity of a deal: the sender's registrable
// domain plus the normalized title. Newsletter URLs are per-recipient tracking
// redirects, so they are deliberately excluded.
func DedupeKey(fromAddr, title string) string {
	return registrableDomain(fromAddr) + "|" + normalizeTitle(title)
}

func registrableDomain(fromAddr string) string {
	addr := strings.TrimSpace(fromAddr)
	if a, err := mail.ParseAddress(fromAddr); err == nil {
		addr = a.Address
	}
	at := strings.LastIndex(addr, "@")
	if at < 0 {
		return strings.ToLower(addr)
	}
	domain := strings.ToLower(strings.TrimSpace(addr[at+1:]))
	if reg, err := publicsuffix.EffectiveTLDPlusOne(domain); err == nil {
		return reg
	}
	return domain
}

func normalizeTitle(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// --- scanning & null helpers ---

type scanner interface {
	Scan(dest ...any) error
}

func scanEmail(sc scanner) (Email, error) {
	var (
		e         Email
		sentAt    sql.NullString
		received  string
		extractEr sql.NullString
	)
	if err := sc.Scan(&e.ID, &e.MessageID, &e.From, &e.To, &e.Subject,
		&sentAt, &received, &e.BodyText, &e.Status, &extractEr); err != nil {
		return Email{}, err
	}
	if sentAt.Valid {
		if t, err := parseTime(sentAt.String); err == nil {
			e.SentAt = &t
		}
	}
	if t, err := parseTime(received); err == nil {
		e.ReceivedAt = t
	}
	e.ExtractError = extractEr.String
	return e, nil
}

func scanDeal(sc scanner) (Deal, error) {
	var (
		d                               Deal
		url, price, endsAt, description sql.NullString
		firstSeen, lastSeen             string
		postedIn                        sql.NullInt64
	)
	if err := sc.Scan(&d.ID, &d.EmailID, &d.DedupeKey, &d.Source, &d.Title,
		&url, &price, &endsAt, &description, &firstSeen, &lastSeen, &d.SeenCount, &postedIn); err != nil {
		return Deal{}, err
	}
	d.URL = url.String
	d.Price = price.String
	d.EndsAt = endsAt.String
	d.Description = description.String
	if t, err := parseTime(firstSeen); err == nil {
		d.FirstSeenAt = t
	}
	if t, err := parseTime(lastSeen); err == nil {
		d.LastSeenAt = t
	}
	if postedIn.Valid {
		id := postedIn.Int64
		d.PostedInDigestID = &id
	}
	return d, nil
}

// formatTime renders a timestamp as fixed-width RFC3339 in UTC, so stored values
// sort lexicographically in chronological order (all end in "Z").
func formatTime(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func parseTime(s string) (time.Time, error) { return time.Parse(time.RFC3339, s) }

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTime(*t)
}
