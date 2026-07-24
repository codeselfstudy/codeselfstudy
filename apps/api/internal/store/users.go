package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrUsernameTaken is returned by SetUsername (and surfaced from the upsert dedupe
// internally) when a username collides with an existing one. The comparison is
// case-insensitive: it is the users_username_unique index (COLLATE NOCASE) that
// rejects the write, so "JaneDoe" and "janedoe" collide. Callers map this to 409.
var ErrUsernameTaken = errors.New("store: username already taken")

// maxUsernameSuffix bounds the deterministic "-2", "-3", … dedupe attempts before
// falling back to a random suffix. Ten is far more than any real name collides in
// practice; the random fallback only exists so a pathological run can still insert.
const maxUsernameSuffix = 10

// usernameMaxLen mirrors users.MaxLen. It is duplicated (rather than imported) to
// keep the store free of a dependency on internal/users; the store only needs it
// to keep a suffixed candidate within the same width the app validates.
const usernameMaxLen = 30

// User is a local account row. WorkOS remains the identity provider; this row
// adds the site-owned fields (a unique username, timezone) plus a mirror of the
// email/avatar for display without a WorkOS round-trip. UsernameChangedAt is nil
// until the user first renames, and drives the change-rate limit.
type User struct {
	ID                int64
	WorkOSID          string
	Email             string
	Username          string
	Timezone          string
	AvatarURL         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	UsernameChangedAt *time.Time
}

// DeletionRequest is a user's request that an admin delete their account. Users
// never self-delete (data-integrity risk); this row is the durable record the
// admin actions manually.
type DeletionRequest struct {
	ID          int64
	UserID      int64
	RequestedAt time.Time
	Reason      string
	Status      string
	HandledAt   *time.Time
}

// UpsertUserByWorkOSID creates or refreshes the local row for a WorkOS user.
//
// On first sight it inserts a row seeded with candidate (u.Username, normally
// from users.Generate), deduping against the unique index so the stored username
// is unique even when two people generate the same one. On later logins it
// refreshes the WorkOS-owned fields (email, avatar) and leaves the user's chosen
// username and timezone untouched, returning isNew=false.
//
// u must carry WorkOSID, Email, Username (the candidate), and AvatarURL. The
// returned User is the stored row (with the possibly-suffixed username).
func (s *Store) UpsertUserByWorkOSID(ctx context.Context, u User) (stored User, isNew bool, err error) {
	existing, err := s.GetUserByWorkOSID(ctx, u.WorkOSID)
	switch {
	case err == nil:
		// Returning user: WorkOS is the source of truth for email/avatar, so
		// refresh those; never touch the username the user may have chosen.
		refreshed, rerr := s.refreshExistingUser(ctx, existing, u)
		if rerr != nil {
			return User{}, false, rerr
		}
		return refreshed, false, nil
	case errors.Is(err, sql.ErrNoRows):
		return s.insertUserDedup(ctx, u)
	default:
		return User{}, false, err
	}
}

// insertUserDedup inserts a new user, walking candidate → candidate-2 → … on a
// username collision, then a random suffix, so a first login always lands a row.
func (s *Store) insertUserDedup(ctx context.Context, u User) (User, bool, error) {
	now := s.Now().UTC()
	base := u.Username
	if base == "" {
		base = "user"
	}

	candidate := base
	for attempt := 1; attempt <= maxUsernameSuffix; attempt++ {
		id, err := s.tryInsertUser(ctx, u.WorkOSID, u.Email, candidate, u.AvatarURL, now)
		switch {
		case err == nil:
			return storedNewUser(u, id, candidate, now), true, nil
		case isUsernameConflict(err):
			candidate = withSuffix(base, "-"+strconv.Itoa(attempt+1)) // -2, -3, …
		case isWorkOSConflict(err):
			// A concurrent login for the same brand-new user won the insert
			// race; adopt its row rather than erroring the second login, and
			// apply this login's WorkOS email/avatar just like the normal path.
			existing, gerr := s.GetUserByWorkOSID(ctx, u.WorkOSID)
			if gerr != nil {
				return User{}, false, gerr
			}
			refreshed, rerr := s.refreshExistingUser(ctx, existing, u)
			if rerr != nil {
				return User{}, false, rerr
			}
			return refreshed, false, nil
		default:
			return User{}, false, fmt.Errorf("insert user: %w", err)
		}
	}

	// Deterministic suffixes exhausted — fall back to a random one.
	candidate = withSuffix(base, "-"+randomSuffix())
	id, err := s.tryInsertUser(ctx, u.WorkOSID, u.Email, candidate, u.AvatarURL, now)
	if err != nil {
		return User{}, false, fmt.Errorf("insert user (random suffix): %w", err)
	}
	return storedNewUser(u, id, candidate, now), true, nil
}

// tryInsertUser attempts a single INSERT, returning the new id or the raw driver
// error (so the caller can classify a UNIQUE violation).
func (s *Store) tryInsertUser(ctx context.Context, workosID, email, username, avatar string, now time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (workos_id, email, username, timezone, avatar_url, created_at, updated_at)
		 VALUES (?,?,?,'',?,?,?)`,
		workosID, email, username, avatar, formatTime(now), formatTime(now))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// refreshExistingUser applies the WorkOS-owned mirror fields (email, avatar) to an
// existing row and returns the in-memory copy kept in sync with the persisted row,
// including the new updated_at. Shared by the returning-login path and the
// concurrent-insert adopt path so they refresh identically.
func (s *Store) refreshExistingUser(ctx context.Context, existing, u User) (User, error) {
	now := s.Now().UTC()
	if err := s.refreshUserProfile(ctx, existing.ID, u.Email, u.AvatarURL, now); err != nil {
		return User{}, err
	}
	existing.Email = u.Email
	existing.AvatarURL = u.AvatarURL
	existing.UpdatedAt = now
	return existing, nil
}

// refreshUserProfile updates the WorkOS-owned mirror fields on each login. now is
// the timestamp written to updated_at (passed in so the caller can reflect it on
// the returned row).
func (s *Store) refreshUserProfile(ctx context.Context, id int64, email, avatar string, now time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET email = ?, avatar_url = ?, updated_at = ? WHERE id = ?`,
		email, avatar, formatTime(now), id)
	if err != nil {
		return fmt.Errorf("refresh user profile: %w", err)
	}
	return nil
}

// GetUserByWorkOSID returns the user with the given WorkOS id, or sql.ErrNoRows.
func (s *Store) GetUserByWorkOSID(ctx context.Context, workosID string) (User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, workos_id, email, username, timezone, avatar_url, created_at, updated_at, username_changed_at
		 FROM users WHERE workos_id = ?`, workosID)
	return scanUser(row)
}

// SetUsername changes a user's username and stamps username_changed_at (which the
// handler's rate limit reads). Returns ErrUsernameTaken when the new name
// collides case-insensitively with another user's.
func (s *Store) SetUsername(ctx context.Context, id int64, username string) error {
	now := formatTime(s.Now().UTC())
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET username = ?, username_changed_at = ?, updated_at = ? WHERE id = ?`,
		username, now, now, id)
	if err != nil {
		if isUsernameConflict(err) {
			return ErrUsernameTaken
		}
		return fmt.Errorf("set username: %w", err)
	}
	return nil
}

// SetTimezone changes a user's timezone. Unlimited (unlike username changes).
func (s *Store) SetTimezone(ctx context.Context, id int64, tz string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET timezone = ?, updated_at = ? WHERE id = ?`,
		tz, formatTime(s.Now().UTC()), id)
	if err != nil {
		return fmt.Errorf("set timezone: %w", err)
	}
	return nil
}

// CreateDeletionRequest files a pending account-deletion request for the user. It
// is idempotent: the deletion_requests_pending UNIQUE index allows only one open
// request per user, so a second (even concurrent) call — a double-click — hits the
// constraint and returns created=false rather than filing a duplicate. The first
// request's reason is preserved. A UNIQUE violation is the only way this INSERT can
// fail on that index; other failures (e.g. the foreign key) surface as errors.
func (s *Store) CreateDeletionRequest(ctx context.Context, userID int64, reason string) (created bool, err error) {
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO account_deletion_requests (user_id, requested_at, reason, status)
		 VALUES (?,?,?,'pending')`,
		userID, formatTime(s.Now().UTC()), reason)
	if err != nil {
		if isUniqueViolation(err) {
			return false, nil // a pending request already exists — idempotent
		}
		return false, fmt.Errorf("create deletion request: %w", err)
	}
	return true, nil
}

// PendingDeletionRequest returns the user's open deletion request, or (nil, nil)
// when there is none.
func (s *Store) PendingDeletionRequest(ctx context.Context, userID int64) (*DeletionRequest, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, requested_at, reason, status, handled_at
		 FROM account_deletion_requests
		 WHERE user_id = ? AND status = 'pending'
		 ORDER BY id DESC LIMIT 1`, userID)
	dr, err := scanDeletionRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dr, nil
}

// --- helpers ---

// storedNewUser fills in the fields set at insert time on the returned row.
func storedNewUser(u User, id int64, username string, now time.Time) User {
	u.ID = id
	u.Username = username
	u.Timezone = ""
	u.CreatedAt = now
	u.UpdatedAt = now
	u.UsernameChangedAt = nil
	return u
}

func scanUser(sc scanner) (User, error) {
	var (
		u                User
		created, updated string
		changedAt        sql.NullString
	)
	if err := sc.Scan(&u.ID, &u.WorkOSID, &u.Email, &u.Username, &u.Timezone,
		&u.AvatarURL, &created, &updated, &changedAt); err != nil {
		return User{}, err
	}
	if t, err := parseTime(created); err == nil {
		u.CreatedAt = t
	}
	if t, err := parseTime(updated); err == nil {
		u.UpdatedAt = t
	}
	if changedAt.Valid {
		if t, err := parseTime(changedAt.String); err == nil {
			u.UsernameChangedAt = &t
		}
	}
	return u, nil
}

func scanDeletionRequest(sc scanner) (DeletionRequest, error) {
	var (
		dr        DeletionRequest
		requested string
		handled   sql.NullString
	)
	if err := sc.Scan(&dr.ID, &dr.UserID, &requested, &dr.Reason, &dr.Status, &handled); err != nil {
		return DeletionRequest{}, err
	}
	if t, err := parseTime(requested); err == nil {
		dr.RequestedAt = t
	}
	if handled.Valid {
		if t, err := parseTime(handled.String); err == nil {
			dr.HandledAt = &t
		}
	}
	return dr, nil
}

// withSuffix appends suffix to base, trimming base so the whole stays within
// usernameMaxLen (the width the app validates).
func withSuffix(base, suffix string) string {
	if len(base)+len(suffix) > usernameMaxLen {
		base = base[:usernameMaxLen-len(suffix)]
	}
	return base + suffix
}

// randomSuffix returns 6 hex characters of randomness for the last-resort
// username dedupe.
func randomSuffix() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is catastrophic and near-impossible; a fixed
		// marker still yields a syntactically valid, retry-able candidate.
		return "000000"
	}
	return hex.EncodeToString(b)
}

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint failure.
// Both drivers (modernc.org/sqlite and libsql) surface SQLite's canonical
// "UNIQUE constraint failed: <table>.<column>" message, which the column-specific
// helpers below then narrow.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func isUsernameConflict(err error) bool {
	return isUniqueViolation(err) && strings.Contains(err.Error(), "username")
}

func isWorkOSConflict(err error) bool {
	return isUniqueViolation(err) && strings.Contains(err.Error(), "workos_id")
}
