package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

func TestUpsertUserByWorkOSID_CreatesThenRefreshes(t *testing.T) {
	s := newStore(t)

	first, isNew, err := s.UpsertUserByWorkOSID(ctx, store.User{
		WorkOSID:  "wos_1",
		Email:     "jane@example.com",
		Username:  "janedoe",
		AvatarURL: "https://img/1.png",
	})
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if !isNew {
		t.Fatalf("first upsert isNew = false, want true")
	}
	if first.ID == 0 || first.Username != "janedoe" {
		t.Fatalf("first upsert = %+v, want id!=0 and username janedoe", first)
	}

	// A returning login: same workos_id, new email/avatar. The row is refreshed,
	// not duplicated, and the username is preserved.
	second, isNew, err := s.UpsertUserByWorkOSID(ctx, store.User{
		WorkOSID:  "wos_1",
		Email:     "jane.new@example.com",
		Username:  "IGNORED-candidate",
		AvatarURL: "https://img/2.png",
	})
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if isNew {
		t.Fatalf("second upsert isNew = true, want false")
	}
	if second.ID != first.ID {
		t.Fatalf("second upsert id = %d, want %d (same row)", second.ID, first.ID)
	}
	if second.Username != "janedoe" {
		t.Errorf("username changed on refresh: got %q, want janedoe", second.Username)
	}
	if second.Email != "jane.new@example.com" || second.AvatarURL != "https://img/2.png" {
		t.Errorf("email/avatar not refreshed: %+v", second)
	}
}

func TestUpsertUserByWorkOSID_DedupesUsername(t *testing.T) {
	s := newStore(t)

	if _, _, err := s.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_a", Email: "a@x.com", Username: "janedoe"}); err != nil {
		t.Fatalf("insert A: %v", err)
	}
	// Different WorkOS user, colliding candidate in DIFFERENT case: the NOCASE
	// index treats "JaneDoe" as the same handle, so it must be suffixed.
	b, _, err := s.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_b", Email: "b@x.com", Username: "JaneDoe"})
	if err != nil {
		t.Fatalf("insert B: %v", err)
	}
	if b.Username != "JaneDoe-2" {
		t.Fatalf("second janedoe = %q, want JaneDoe-2", b.Username)
	}
	// A third collision walks to -3.
	c, _, err := s.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_c", Email: "c@x.com", Username: "janedoe"})
	if err != nil {
		t.Fatalf("insert C: %v", err)
	}
	if c.Username != "janedoe-3" {
		t.Fatalf("third janedoe = %q, want janedoe-3", c.Username)
	}
}

func TestUpsertUserByWorkOSID_EmptyCandidateBecomesUser(t *testing.T) {
	s := newStore(t)
	u, _, err := s.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_e", Email: "e@x.com", Username: ""})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if u.Username != "user" {
		t.Errorf("empty candidate = %q, want user", u.Username)
	}
}

func TestSetUsername_StampsAndConflicts(t *testing.T) {
	s := newStore(t)
	clock := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	setNow := fixedClock(s, clock)

	a, _, _ := s.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_a", Email: "a@x.com", Username: "alice"})
	b, _, _ := s.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_b", Email: "b@x.com", Username: "bob"})

	// A fresh account has no change stamp.
	if a.UsernameChangedAt != nil {
		t.Fatalf("new user UsernameChangedAt = %v, want nil", a.UsernameChangedAt)
	}

	setNow(clock.Add(48 * time.Hour))
	if err := s.SetUsername(ctx, a.ID, "alice2"); err != nil {
		t.Fatalf("SetUsername: %v", err)
	}
	got, err := s.GetUserByWorkOSID(ctx, "wos_a")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Username != "alice2" {
		t.Errorf("username = %q, want alice2", got.Username)
	}
	if got.UsernameChangedAt == nil || !got.UsernameChangedAt.Equal(clock.Add(48*time.Hour)) {
		t.Errorf("UsernameChangedAt = %v, want %v", got.UsernameChangedAt, clock.Add(48*time.Hour))
	}

	// Case-insensitive collision with bob's name → ErrUsernameTaken.
	if err := s.SetUsername(ctx, a.ID, "BOB"); !errors.Is(err, store.ErrUsernameTaken) {
		t.Fatalf("SetUsername to taken name = %v, want ErrUsernameTaken", err)
	}
	_ = b
}

func TestSetTimezone(t *testing.T) {
	s := newStore(t)
	u, _, _ := s.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_tz", Email: "t@x.com", Username: "tzuser"})
	if err := s.SetTimezone(ctx, u.ID, "America/Los_Angeles"); err != nil {
		t.Fatalf("SetTimezone: %v", err)
	}
	got, _ := s.GetUserByWorkOSID(ctx, "wos_tz")
	if got.Timezone != "America/Los_Angeles" {
		t.Errorf("timezone = %q, want America/Los_Angeles", got.Timezone)
	}
}

func TestGetUserByWorkOSID_NotFound(t *testing.T) {
	s := newStore(t)
	_, err := s.GetUserByWorkOSID(ctx, "nobody")
	if err == nil {
		t.Fatal("GetUserByWorkOSID(nobody) = nil error, want sql.ErrNoRows")
	}
}

func TestDeletionRequest_Idempotent(t *testing.T) {
	s := newStore(t)
	u, _, _ := s.UpsertUserByWorkOSID(ctx, store.User{WorkOSID: "wos_d", Email: "d@x.com", Username: "delme"})

	if pending, err := s.PendingDeletionRequest(ctx, u.ID); err != nil || pending != nil {
		t.Fatalf("initial pending = (%v, %v), want (nil, nil)", pending, err)
	}

	created, err := s.CreateDeletionRequest(ctx, u.ID, "moving on")
	if err != nil || !created {
		t.Fatalf("first CreateDeletionRequest = (%v, %v), want (true, nil)", created, err)
	}
	created, err = s.CreateDeletionRequest(ctx, u.ID, "again")
	if err != nil || created {
		t.Fatalf("second CreateDeletionRequest = (%v, %v), want (false, nil)", created, err)
	}

	pending, err := s.PendingDeletionRequest(ctx, u.ID)
	if err != nil {
		t.Fatalf("PendingDeletionRequest: %v", err)
	}
	if pending == nil || pending.Status != "pending" || pending.Reason != "moving on" {
		t.Errorf("pending = %+v, want the first request (reason 'moving on')", pending)
	}
}
