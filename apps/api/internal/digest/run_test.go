package digest

import (
	"context"
	"errors"
	"fmt"
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

	posted, err := Run(ctx, s, fp, interval, stale, false)
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

	posted, err := Run(ctx, s, fp, interval, stale, false)
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
	posted, err = Run(ctx, s, fp2, interval, stale, false)
	if err != nil || !posted {
		t.Fatalf("retry after failure: posted=%v err=%v", posted, err)
	}
}

func TestRunEmptyReleasesClaim(t *testing.T) {
	s := newStore(t)
	fp := &fakePoster{}

	posted, err := Run(ctx, s, fp, interval, stale, false)
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

	posted, err := Run(ctx, s, fp, interval, stale, false)
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

	if posted, err := Run(ctx, s, fp, interval, stale, false); err != nil || !posted {
		t.Fatalf("first run: posted=%v err=%v", posted, err)
	}

	// A new deal arrives within the interval.
	addDeal(t, s, id, "Latecomer")

	// Non-force is suppressed (a digest was posted < interval ago).
	if posted, _ := Run(ctx, s, fp, interval, stale, false); posted {
		t.Fatal("expected suppression within interval")
	}
	if fp.count() != 1 {
		t.Fatalf("poster called %d times, want 1", fp.count())
	}
	// Force posts the new deal despite the interval.
	if posted, err := Run(ctx, s, fp, interval, stale, true); err != nil || !posted {
		t.Fatalf("force run: posted=%v err=%v", posted, err)
	}
	if fp.count() != 2 {
		t.Errorf("poster called %d times after force, want 2", fp.count())
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
			posted, err := Run(ctx, s, fp, interval, stale, false)
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
