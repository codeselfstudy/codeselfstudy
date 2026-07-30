package digest

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/db"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

var ctx = context.Background()

const (
	interval = 24 * time.Hour
	stale    = store.DefaultStaleWindow
)

func newStore(t *testing.T) *store.Store {
	t.Helper()
	d, err := db.Open("file:" + filepath.Join(t.TempDir(), "d.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Migrate(ctx, d); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	return store.New(d)
}

func seedEmail(t *testing.T, s *store.Store, mid string) int64 {
	t.Helper()
	e, _, err := s.InsertEmail(ctx, store.Email{MessageID: mid, From: "deals@humblebundle.com", BodyText: "b"})
	if err != nil {
		t.Fatalf("InsertEmail: %v", err)
	}
	return e.ID
}

func addDeal(t *testing.T, s *store.Store, emailID int64, title string) {
	t.Helper()
	key := store.DedupeKey("deals@humblebundle.com", title)
	d := store.Deal{EmailID: emailID, DedupeKey: key, Source: "Humble", Title: title, URL: "https://h/" + key}
	if err := s.UpsertDeal(ctx, d, time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("UpsertDeal: %v", err)
	}
}

func addDeals(t *testing.T, s *store.Store, emailID int64, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		addDeal(t, s, emailID, fmt.Sprintf("Deal%d", i))
	}
}

type fakePoster struct {
	mu       sync.Mutex
	payloads [][]byte
	err      error
}

func (f *fakePoster) Post(_ context.Context, payload []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.payloads = append(f.payloads, append([]byte(nil), payload...))
	return nil
}

func (f *fakePoster) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.payloads)
}

func TestRunHappyPath(t *testing.T) {
	s := newStore(t)
	id := seedEmail(t, s, "<h@x>")
	addDeals(t, s, id, 2)
	fp := &fakePoster{}

	posted, err := Run(ctx, s, fp, interval, stale, false, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !posted {
		t.Fatal("expected posted=true")
	}
	if fp.count() != 1 {
		t.Fatalf("poster called %d times, want 1", fp.count())
	}
	if left, _ := s.UnpostedDeals(ctx, 0); len(left) != 0 {
		t.Errorf("%d deals still unposted after digest", len(left))
	}
}

func TestRunPostFailureKeepsDealsQueued(t *testing.T) {
	s := newStore(t)
	id := seedEmail(t, s, "<f@x>")
	addDeals(t, s, id, 2)
	fp := &fakePoster{err: errors.New("slack down")}

	posted, err := Run(ctx, s, fp, interval, stale, false, nil)
	if err == nil {
		t.Fatal("expected error on post failure")
	}
	if posted {
		t.Fatal("expected posted=false")
	}
	if left, _ := s.UnpostedDeals(ctx, 0); len(left) != 2 {
		t.Errorf("deals should stay queued, got %d unposted", len(left))
	}
	// A failed digest must not suppress the next attempt.
	fp2 := &fakePoster{}
	posted, err = Run(ctx, s, fp2, interval, stale, false, nil)
	if err != nil || !posted {
		t.Fatalf("retry after failure: posted=%v err=%v", posted, err)
	}
}

func TestRunEmptyReleasesClaim(t *testing.T) {
	s := newStore(t)
	fp := &fakePoster{}

	posted, err := Run(ctx, s, fp, interval, stale, false, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if posted {
		t.Fatal("expected posted=false for empty digest")
	}
	if fp.count() != 0 {
		t.Errorf("poster should not be called for empty digest")
	}
	// The empty claim must have been released: a fresh claim can be won.
	if _, ok, _ := s.ClaimDigest(ctx, interval, stale, false); !ok {
		t.Error("empty run did not release its claim")
	}
}

func TestRunOverflowMarksOnlyShown(t *testing.T) {
	s := newStore(t)
	id := seedEmail(t, s, "<o@x>")
	addDeals(t, s, id, MaxDealsPerDigest+1)
	fp := &fakePoster{}

	posted, err := Run(ctx, s, fp, interval, stale, false, nil)
	if err != nil || !posted {
		t.Fatalf("Run: posted=%v err=%v", posted, err)
	}
	left, _ := s.UnpostedDeals(ctx, 0)
	if len(left) != 1 {
		t.Errorf("expected 1 deal left queued, got %d", len(left))
	}
}

func TestRunSuppressedThenForce(t *testing.T) {
	s := newStore(t)
	id := seedEmail(t, s, "<s@x>")
	addDeals(t, s, id, 2)
	fp := &fakePoster{}

	if posted, err := Run(ctx, s, fp, interval, stale, false, nil); err != nil || !posted {
		t.Fatalf("first run: posted=%v err=%v", posted, err)
	}

	// A new deal arrives within the interval.
	addDeal(t, s, id, "Latecomer")

	// Non-force is suppressed (a digest was posted < interval ago).
	if posted, _ := Run(ctx, s, fp, interval, stale, false, nil); posted {
		t.Fatal("expected suppression within interval")
	}
	if fp.count() != 1 {
		t.Fatalf("poster called %d times, want 1", fp.count())
	}
	// Force posts the new deal despite the interval.
	if posted, err := Run(ctx, s, fp, interval, stale, true, nil); err != nil || !posted {
		t.Fatalf("force run: posted=%v err=%v", posted, err)
	}
	if fp.count() != 2 {
		t.Errorf("poster called %d times after force, want 2", fp.count())
	}
}

// fakeCondenser returns canned groups (or an error), and records what it saw.
type fakeCondenser struct {
	mu     sync.Mutex
	groups []CondensedGroup
	err    error
	queued []store.Deal
	recent []store.Deal
	calls  int
}

func (f *fakeCondenser) Condense(_ context.Context, queued, recent []store.Deal) ([]CondensedGroup, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.queued = queued
	f.recent = recent
	return f.groups, f.err
}

func TestRunCondensedDigest(t *testing.T) {
	// With a condenser, the digest posts per-source bullets: bold source name,
	// "• <url|text>" lines linked via the cited deal's stored URL, and all
	// queued deals — including ones condensed away — marked posted.
	s := newStore(t)
	id := seedEmail(t, s, "<cd@x>")
	addDeals(t, s, id, 4)
	queued, _ := s.UnpostedDeals(ctx, 0)
	fp := &fakePoster{}
	fc := &fakeCondenser{groups: []CondensedGroup{{
		Source: "Humble",
		Bullets: []Bullet{
			{DealID: queued[0].ID, Text: "50% off all books"},
			{DealID: queued[3].ID, Text: "Free webinar: The Rust Journey"},
		},
	}}}

	posted, err := Run(ctx, s, fp, interval, stale, false, fc)
	if err != nil || !posted {
		t.Fatalf("Run: posted=%v err=%v", posted, err)
	}
	if fp.count() != 1 {
		t.Fatalf("poster called %d times, want 1", fp.count())
	}
	got := string(fp.payloads[0])
	for _, want := range []string{
		"2 new deals", // header counts bullets, not consumed deals
		"*Humble*",
		"• ",
		// The bullet links deal 0's stored URL (its "|" percent-encoded for
		// Slack link syntax) with the condensed text as the label.
		"https://h/humblebundle.com%7Cdeal0|50% off all books",
		"Free webinar: The Rust Journey",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("payload missing %q:\n%s", want, got)
		}
	}
	if left, _ := s.UnpostedDeals(ctx, 0); len(left) != 0 {
		t.Errorf("%d deals still unposted — condensed-away deals must be consumed", len(left))
	}
}

func TestRunCondenserSeesRecentDeals(t *testing.T) {
	// The condenser gets recently-posted deals as already-announced context.
	s := newStore(t)
	id := seedEmail(t, s, "<cr@x>")
	addDeals(t, s, id, 1)
	fp := &fakePoster{}
	if posted, err := Run(ctx, s, fp, interval, stale, false, nil); err != nil || !posted {
		t.Fatalf("seed run: posted=%v err=%v", posted, err)
	}

	addDeal(t, s, id, "Latecomer")
	fc := &fakeCondenser{} // no groups: everything suppressed
	posted, err := Run(ctx, s, fp, interval, stale, true, fc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if posted {
		t.Fatal("expected posted=false when the condenser returns no bullets")
	}
	if fp.count() != 1 {
		t.Fatalf("poster called %d times, want 1 (suppressed cycle must not post)", fp.count())
	}
	if len(fc.recent) != 1 || fc.recent[0].Title != "Deal0" {
		t.Errorf("condenser recent = %+v, want the previously posted Deal0", fc.recent)
	}
	if len(fc.queued) != 1 || fc.queued[0].Title != "Latecomer" {
		t.Errorf("condenser queued = %+v, want just Latecomer", fc.queued)
	}
	// The suppressed deal is consumed: nothing left to re-announce.
	if left, _ := s.UnpostedDeals(ctx, 0); len(left) != 0 {
		t.Errorf("%d deals still unposted after suppressed cycle", len(left))
	}
}

func TestRunCondenserErrorFallsBack(t *testing.T) {
	// A condenser failure must not lose the digest: the plain format posts and
	// the error is surfaced alongside posted=true for logging.
	s := newStore(t)
	id := seedEmail(t, s, "<ce@x>")
	addDeals(t, s, id, 2)
	fp := &fakePoster{}
	fc := &fakeCondenser{err: errors.New("model down")}

	posted, err := Run(ctx, s, fp, interval, stale, false, fc)
	if !posted {
		t.Fatal("expected posted=true via fallback")
	}
	if err == nil || !strings.Contains(err.Error(), "model down") {
		t.Errorf("err = %v, want the condense failure surfaced", err)
	}
	if fp.count() != 1 {
		t.Fatalf("poster called %d times, want 1", fp.count())
	}
	if got := string(fp.payloads[0]); !strings.Contains(got, "2 new deals") || strings.Contains(got, "• ") {
		t.Errorf("fallback payload should be the plain per-deal format:\n%s", got)
	}
}

func TestRunCondenserUnknownIDDropped(t *testing.T) {
	// Bullets citing ids outside the queued set are the model inventing —
	// dropped; a group left empty disappears. Here that empties everything, so
	// nothing posts.
	s := newStore(t)
	id := seedEmail(t, s, "<cu@x>")
	addDeals(t, s, id, 1)
	fp := &fakePoster{}
	fc := &fakeCondenser{groups: []CondensedGroup{{
		Source:  "Humble",
		Bullets: []Bullet{{DealID: 99999, Text: "invented offer"}, {DealID: 0, Text: "   "}},
	}}}

	posted, err := Run(ctx, s, fp, interval, stale, false, fc)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if posted || fp.count() != 0 {
		t.Fatalf("posted=%v count=%d, want no post when every bullet is invalid", posted, fp.count())
	}
	if left, _ := s.UnpostedDeals(ctx, 0); len(left) != 0 {
		t.Errorf("%d deals still unposted", len(left))
	}
}

func TestRunConcurrentSingleWinner(t *testing.T) {
	s := newStore(t)
	id := seedEmail(t, s, "<c@x>")
	addDeals(t, s, id, 3)
	fp := &fakePoster{}

	var wins int64
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			posted, err := Run(ctx, s, fp, interval, stale, false, nil)
			if err != nil {
				t.Errorf("Run: %v", err)
			}
			if posted {
				atomic.AddInt64(&wins, 1)
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("winners = %d, want 1", wins)
	}
	if fp.count() != 1 {
		t.Fatalf("poster called %d times, want 1", fp.count())
	}
}
