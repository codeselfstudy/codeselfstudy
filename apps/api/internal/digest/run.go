package digest

import (
	"context"
	"fmt"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

// Run posts a digest of unposted deals to Slack, at most once per interval.
//
// It atomically claims the digest slot (see store.ClaimDigest); if another caller
// holds a live claim or a digest was posted within interval (unless force), it
// returns posted=false with no error. An empty deal set releases the claim so it
// does not suppress the next interval. On a successful Slack post the shown deals
// and the digest row are marked posted; on a failed post the digest is marked
// failed and the deals stay queued for a later attempt (the error is returned).
//
// Residual limitation: if the Slack post succeeds but marking the digest posted
// fails every retry, Run returns (true, err) with the row still "claimed". Once
// staleWindow elapses a later run can re-post the same deals (a duplicate, never
// a loss). The retry below shrinks that window to a transient-DB-error edge case.
func Run(ctx context.Context, s *store.Store, poster WebhookPoster, interval, staleWindow time.Duration, force bool) (posted bool, err error) {
	digestID, ok, err := s.ClaimDigest(ctx, interval, staleWindow, force)
	if err != nil {
		return false, fmt.Errorf("claim digest: %w", err)
	}
	if !ok {
		return false, nil
	}

	// Load all unposted deals so we can report overflow accurately; only the first
	// MaxDealsPerDigest are shown and marked posted this round.
	deals, err := s.UnpostedDeals(ctx, 0)
	if err != nil {
		_ = s.DeleteDigest(ctx, digestID)
		return false, fmt.Errorf("load unposted deals: %w", err)
	}
	if len(deals) == 0 {
		if err := s.DeleteDigest(ctx, digestID); err != nil {
			return false, fmt.Errorf("release empty claim: %w", err)
		}
		return false, nil
	}

	body, err := BuildBlocks(deals, s.Now())
	if err != nil {
		_ = s.DeleteDigest(ctx, digestID)
		return false, fmt.Errorf("build blocks: %w", err)
	}

	if err := poster.Post(ctx, body); err != nil {
		if ferr := s.MarkDigestFailed(ctx, digestID); ferr != nil {
			return false, fmt.Errorf("post failed (%v) and mark failed: %w", err, ferr)
		}
		return false, fmt.Errorf("post digest: %w", err)
	}

	shown := deals
	if len(shown) > MaxDealsPerDigest {
		shown = shown[:MaxDealsPerDigest]
	}
	ids := make([]int64, len(shown))
	for i, d := range shown {
		ids[i] = d.ID
	}
	// The post already succeeded, so retry the bookkeeping write a few times
	// before giving up — a transient DB blip here would otherwise risk a
	// duplicate digest on the next run (see the doc comment).
	const markAttempts = 3
	for attempt := 1; ; attempt++ {
		if err := s.MarkDigestPosted(ctx, digestID, ids); err == nil {
			return true, nil
		} else if attempt == markAttempts {
			return true, fmt.Errorf("posted to slack but failed to mark digest posted after %d attempts: %w", markAttempts, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
