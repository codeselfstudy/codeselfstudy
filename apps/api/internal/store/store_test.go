package store_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/db"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

var ctx = context.Background()

const (
	testInterval = 24 * time.Hour
	testStale    = 5 * time.Minute
)

// openMigrated opens a fresh temporary SQLite database and applies the schema.
// A file (not :memory:) is used so TestMigrateIdempotent can reopen the same
// path; the caller is responsible for closing the returned handle.
func openMigrated(t *testing.T, path string) *sql.DB {
	t.Helper()
	d, err := db.Open("file:" + path)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	if err := db.Migrate(ctx, d); err != nil {
		d.Close()
		t.Fatalf("db.Migrate: %v", err)
	}
	return d
}

func newStore(t *testing.T) *store.Store {
	t.Helper()
	d := openMigrated(t, filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { d.Close() })
	return store.New(d)
}

// fixedClock installs a settable clock on the store and returns a setter.
func fixedClock(s *store.Store, t time.Time) func(time.Time) {
	var mu sync.RWMutex
	cur := t
	s.Now = func() time.Time {
		mu.RLock()
		defer mu.RUnlock()
		return cur
	}
	return func(nt time.Time) {
		mu.Lock()
		defer mu.Unlock()
		cur = nt
	}
}

func mustInsertEmail(t *testing.T, s *store.Store, messageID string) store.Email {
	t.Helper()
	e, isNew, err := s.InsertEmail(ctx, store.Email{
		MessageID: messageID,
		From:      "Humble Bundle <deals@humblebundle.com>",
		Subject:   "Deals",
		BodyText:  "body",
	})
	if err != nil {
		t.Fatalf("InsertEmail: %v", err)
	}
	if !isNew {
		t.Fatalf("expected new email for %s", messageID)
	}
	return e
}

func TestMigrateIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	d1 := openMigrated(t, path)
	d1.Close()

	// Re-opening the same file must re-run migrations as a no-op.
	d2 := openMigrated(t, path)
	defer d2.Close()

	// Schema is usable.
	mustInsertEmail(t, store.New(d2), "<a@x>")
}

func TestInsertEmailIdempotent(t *testing.T) {
	s := newStore(t)
	first := mustInsertEmail(t, s, "<dup@x>")

	again, isNew, err := s.InsertEmail(ctx, store.Email{MessageID: "<dup@x>", Subject: "changed", BodyText: "b"})
	if err != nil {
		t.Fatalf("InsertEmail: %v", err)
	}
	if isNew {
		t.Fatal("second insert of same message_id should not be new")
	}
	if again.ID != first.ID {
		t.Errorf("id changed: %d != %d", again.ID, first.ID)
	}
	if again.Subject != "Deals" {
		t.Errorf("existing row was overwritten: subject=%q", again.Subject)
	}
}

func TestSetEmailStatus(t *testing.T) {
	s := newStore(t)
	e := mustInsertEmail(t, s, "<s@x>")

	if err := s.SetEmailStatus(ctx, e.ID, store.StatusExtractFailed, "boom"); err != nil {
		t.Fatalf("SetEmailStatus: %v", err)
	}
	got, err := s.GetEmailByMessageID(ctx, "<s@x>")
	if err != nil {
		t.Fatalf("GetEmailByMessageID: %v", err)
	}
	if got.Status != store.StatusExtractFailed || got.ExtractError != "boom" {
		t.Errorf("status=%q err=%q", got.Status, got.ExtractError)
	}
}

func TestUpsertDealBumpsSeenCountAndCoalesces(t *testing.T) {
	s := newStore(t)
	e := mustInsertEmail(t, s, "<u@x>")
	key := store.DedupeKey("deals@humblebundle.com", "Programming Bundle")
	past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	// First sighting carries a URL but no price.
	d := store.Deal{EmailID: e.ID, DedupeKey: key, Source: "Humble", Title: "Programming Bundle", URL: "https://h/1"}
	if err := s.UpsertDeal(ctx, d, past); err != nil {
		t.Fatalf("UpsertDeal 1: %v", err)
	}
	// Second sighting has no URL (should keep the old one) but adds a price.
	d2 := store.Deal{EmailID: e.ID, DedupeKey: key, Source: "Humble", Title: "Programming Bundle", Price: "$25"}
	if err := s.UpsertDeal(ctx, d2, past); err != nil {
		t.Fatalf("UpsertDeal 2: %v", err)
	}

	got, err := s.GetDealByDedupeKey(ctx, key)
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if got.SeenCount != 2 {
		t.Errorf("SeenCount = %d, want 2", got.SeenCount)
	}
	if got.URL != "https://h/1" {
		t.Errorf("URL = %q, want preserved https://h/1", got.URL)
	}
	if got.Price != "$25" {
		t.Errorf("Price = %q, want $25", got.Price)
	}
}

// seedPostedDeal inserts a deal at seenAt, posts it in a digest, and returns its key.
func seedPostedDeal(t *testing.T, s *store.Store, set func(time.Time), emailID int64, title string, seenAt time.Time) string {
	t.Helper()
	set(seenAt)
	key := store.DedupeKey("deals@humblebundle.com", title)
	if err := s.UpsertDeal(ctx, store.Deal{EmailID: emailID, DedupeKey: key, Source: "Humble", Title: title}, seenAt); err != nil {
		t.Fatalf("seed UpsertDeal: %v", err)
	}
	row, _ := s.GetDealByDedupeKey(ctx, key)
	id, ok, err := s.ClaimDigest(ctx, testInterval, testStale, true)
	if err != nil || !ok {
		t.Fatalf("seed ClaimDigest: ok=%v err=%v", ok, err)
	}
	if err := s.MarkDigestPosted(ctx, id, []int64{row.ID}); err != nil {
		t.Fatalf("seed MarkDigestPosted: %v", err)
	}
	return key
}

func isPosted(t *testing.T, s *store.Store, key string) bool {
	t.Helper()
	d, err := s.GetDealByDedupeKey(ctx, key)
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	return d.PostedInDigestID != nil
}

func TestUpsertDealRepostWindow(t *testing.T) {
	s := newStore(t)
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	set := fixedClock(s, t0)
	e := mustInsertEmail(t, s, "<r@x>")

	// Two independent deals, each first seen and posted at T0.
	staleKey := seedPostedDeal(t, s, set, e.ID, "Stale Bundle", t0)
	freshKey := seedPostedDeal(t, s, set, e.ID, "Fresh Bundle", t0)

	// Re-seen 50 days later. cutoff = now-45d = T0+5d. Old last_seen (T0) < cutoff → re-queue.
	after := t0.Add(50 * 24 * time.Hour)
	set(after)
	if err := s.UpsertDeal(ctx, store.Deal{EmailID: e.ID, DedupeKey: staleKey, Source: "Humble", Title: "Stale Bundle"}, after.Add(-45*24*time.Hour)); err != nil {
		t.Fatalf("UpsertDeal stale: %v", err)
	}
	if isPosted(t, s, staleKey) {
		t.Error("deal unseen past the repost window should be re-queued (posted_in_digest_id NULL)")
	}

	// Re-seen only 10 days later. cutoff = now-45d = T0-35d. Old last_seen (T0) not < cutoff → stays posted.
	within := t0.Add(10 * 24 * time.Hour)
	set(within)
	if err := s.UpsertDeal(ctx, store.Deal{EmailID: e.ID, DedupeKey: freshKey, Source: "Humble", Title: "Fresh Bundle"}, within.Add(-45*24*time.Hour)); err != nil {
		t.Fatalf("UpsertDeal fresh: %v", err)
	}
	if !isPosted(t, s, freshKey) {
		t.Error("deal re-seen within the repost window should stay posted")
	}
}

func TestClaimDigestConcurrentSingleWinner(t *testing.T) {
	s := newStore(t)
	fixedClock(s, time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC))

	const n = 8
	var wins int64
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, ok, err := s.ClaimDigest(ctx, testInterval, testStale, false)
			if err != nil {
				t.Errorf("ClaimDigest: %v", err)
			}
			if ok {
				atomic.AddInt64(&wins, 1)
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("winners = %d, want exactly 1", wins)
	}
}

func TestClaimDigestStaleTakeover(t *testing.T) {
	s := newStore(t)
	t0 := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	set := fixedClock(s, t0)

	// A claim made 10 minutes ago, never completed (crashed claimant).
	set(t0.Add(-10 * time.Minute))
	if _, ok, err := s.ClaimDigest(ctx, testInterval, testStale, false); err != nil || !ok {
		t.Fatalf("initial claim: ok=%v err=%v", ok, err)
	}

	// Now: the stale claim (>5m old) must not block a fresh claim.
	set(t0)
	if _, ok, err := s.ClaimDigest(ctx, testInterval, testStale, false); err != nil || !ok {
		t.Fatalf("stale claim should not block: ok=%v err=%v", ok, err)
	}
	// But the just-made fresh claim must block the next one.
	if _, ok, err := s.ClaimDigest(ctx, testInterval, testStale, false); err != nil || ok {
		t.Fatalf("fresh claim should block: ok=%v err=%v", ok, err)
	}
}

func TestClaimDigestIntervalSuppressionAndForce(t *testing.T) {
	s := newStore(t)
	t0 := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	set := fixedClock(s, t0)

	// Claim and post a digest at T0.
	id, ok, err := s.ClaimDigest(ctx, testInterval, testStale, false)
	if err != nil || !ok {
		t.Fatalf("claim: ok=%v err=%v", ok, err)
	}
	if err := s.MarkDigestPosted(ctx, id, nil); err != nil {
		t.Fatalf("MarkDigestPosted: %v", err)
	}

	// 1h later: within the 24h interval → suppressed.
	set(t0.Add(time.Hour))
	if _, ok, _ := s.ClaimDigest(ctx, testInterval, testStale, false); ok {
		t.Fatal("claim within interval should be suppressed")
	}
	// force=true skips the interval check.
	if _, ok, _ := s.ClaimDigest(ctx, testInterval, testStale, true); !ok {
		t.Fatal("force claim should succeed within interval")
	}
	// force still respects a live claim (the one we just forced).
	if _, ok, _ := s.ClaimDigest(ctx, testInterval, testStale, true); ok {
		t.Fatal("force claim should still respect a live claim")
	}
}

func TestFailedAndEmptyDigestsDoNotBlock(t *testing.T) {
	s := newStore(t)
	fixedClock(s, time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC))

	// A failed digest must not block a future claim.
	id, ok, _ := s.ClaimDigest(ctx, testInterval, testStale, false)
	if !ok {
		t.Fatal("first claim failed")
	}
	if err := s.MarkDigestFailed(ctx, id); err != nil {
		t.Fatalf("MarkDigestFailed: %v", err)
	}
	id2, ok, _ := s.ClaimDigest(ctx, testInterval, testStale, false)
	if !ok {
		t.Fatal("claim should succeed after a failed digest")
	}

	// Releasing an empty claim must free the slot again.
	if err := s.DeleteDigest(ctx, id2); err != nil {
		t.Fatalf("DeleteDigest: %v", err)
	}
	if _, ok, _ := s.ClaimDigest(ctx, testInterval, testStale, false); !ok {
		t.Fatal("claim should succeed after releasing an empty claim")
	}
}

func TestUnpostedDealsOrderAndLimit(t *testing.T) {
	s := newStore(t)
	set := fixedClock(s, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	e := mustInsertEmail(t, s, "<o@x>")

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, title := range []string{"First", "Second", "Third"} {
		set(base.Add(time.Duration(i) * time.Hour))
		key := store.DedupeKey("deals@humblebundle.com", title)
		if err := s.UpsertDeal(ctx, store.Deal{EmailID: e.ID, DedupeKey: key, Source: "H", Title: title}, base.Add(-time.Hour)); err != nil {
			t.Fatalf("UpsertDeal %s: %v", title, err)
		}
	}

	all, err := s.UnpostedDeals(ctx, 0)
	if err != nil {
		t.Fatalf("UnpostedDeals: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("got %d deals, want 3", len(all))
	}
	if all[0].Title != "First" || all[2].Title != "Third" {
		t.Errorf("order wrong: %s ... %s", all[0].Title, all[2].Title)
	}
	limited, _ := s.UnpostedDeals(ctx, 2)
	if len(limited) != 2 {
		t.Errorf("limit 2 returned %d", len(limited))
	}
}

func TestMarkDigestPostedStampsAllDeals(t *testing.T) {
	s := newStore(t)
	e := mustInsertEmail(t, s, "<m@x>")

	var ids []int64
	for _, title := range []string{"A", "B", "C"} {
		key := store.DedupeKey("deals@humblebundle.com", title)
		if err := s.UpsertDeal(ctx, store.Deal{EmailID: e.ID, DedupeKey: key, Source: "H", Title: title}, time.Time{}); err != nil {
			t.Fatalf("UpsertDeal %s: %v", title, err)
		}
		d, _ := s.GetDealByDedupeKey(ctx, key)
		ids = append(ids, d.ID)
	}

	id, ok, err := s.ClaimDigest(ctx, testInterval, testStale, false)
	if err != nil || !ok {
		t.Fatalf("ClaimDigest: ok=%v err=%v", ok, err)
	}
	if err := s.MarkDigestPosted(ctx, id, ids); err != nil {
		t.Fatalf("MarkDigestPosted: %v", err)
	}
	if left, _ := s.UnpostedDeals(ctx, 0); len(left) != 0 {
		t.Fatalf("all deals should be stamped, %d still unposted", len(left))
	}
}

func TestDedupeKey(t *testing.T) {
	tests := []struct {
		from, title, want string
	}{
		{"Humble Bundle <deals@humblebundle.com>", "Programming Bundle!", "humblebundle.com|programmingbundle"},
		{"deals@mail.humblebundle.com", "Programming Bundle!", "humblebundle.com|programmingbundle"},
		{"news@example.co.uk", "Weekly  Deals", "example.co.uk|weeklydeals"},
		{"deals@humblebundle.com", "PROGRAMMING bundle", "humblebundle.com|programmingbundle"},
	}
	for _, tt := range tests {
		if got := store.DedupeKey(tt.from, tt.title); got != tt.want {
			t.Errorf("DedupeKey(%q, %q) = %q, want %q", tt.from, tt.title, got, tt.want)
		}
	}
}
