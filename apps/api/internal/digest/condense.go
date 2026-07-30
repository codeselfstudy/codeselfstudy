package digest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

const (
	// suppressionWindow is how far back already-posted deals are shown to the
	// condenser as "already announced", so an ongoing promotion re-extracted
	// from a later email under new wording is not re-posted.
	suppressionWindow = 7 * 24 * time.Hour

	// maxRecentForCondense caps the already-announced list fed to the model;
	// a week of digests fits comfortably, and the prompt stays bounded.
	maxRecentForCondense = 25

	// condenseTimeout bounds the model call. Past it the digest posts in the
	// uncondensed format rather than stall the posting path.
	condenseTimeout = 20 * time.Second
)

// Bullet is one line of a condensed digest. DealID cites the queued deal the
// line is about — that deal's stored URL becomes the link — and Text is the
// short offer description ("50% off all books").
type Bullet struct {
	DealID int64
	Text   string
}

// CondensedGroup is one source's bullets.
type CondensedGroup struct {
	Source  string
	Bullets []Bullet
}

// Condenser condenses a batch of queued deals into per-source essentials:
// equivalent rewordings merge into one bullet, products merely carrying an
// umbrella discount drop, and offers equivalent to something in recent
// (already announced) are omitted entirely. Implementations must cite only
// deal ids from queued and must not invent prices or dates; Run drops any
// bullet citing an unknown id.
type Condenser interface {
	Condense(ctx context.Context, queued, recent []store.Deal) ([]CondensedGroup, error)
}

// condenseDeals loads the recently-announced context and runs the condenser
// over the queued deals, returning sanitized groups.
func condenseDeals(ctx context.Context, s *store.Store, c Condenser, queued []store.Deal) ([]CondensedGroup, error) {
	cctx, cancel := context.WithTimeout(ctx, condenseTimeout)
	defer cancel()
	recent, err := s.RecentlyPostedDeals(cctx, s.Now().Add(-suppressionWindow), maxRecentForCondense)
	if err != nil {
		return nil, fmt.Errorf("load recent deals: %w", err)
	}
	groups, err := c.Condense(cctx, queued, recent)
	if err != nil {
		return nil, fmt.Errorf("condense: %w", err)
	}
	return sanitizeGroups(groups, queued), nil
}

// sanitizeGroups enforces the Condenser contract on model output: bullets must
// cite a queued deal and carry text. Offending bullets are dropped, as are
// groups left empty.
func sanitizeGroups(groups []CondensedGroup, queued []store.Deal) []CondensedGroup {
	known := make(map[int64]bool, len(queued))
	for _, d := range queued {
		known[d.ID] = true
	}
	var out []CondensedGroup
	for _, g := range groups {
		var bullets []Bullet
		for _, b := range g.Bullets {
			if known[b.DealID] && strings.TrimSpace(b.Text) != "" {
				bullets = append(bullets, b)
			}
		}
		if len(bullets) > 0 {
			out = append(out, CondensedGroup{Source: g.Source, Bullets: bullets})
		}
	}
	return out
}

// countBullets returns the total bullets across groups.
func countBullets(groups []CondensedGroup) int {
	n := 0
	for _, g := range groups {
		n += len(g.Bullets)
	}
	return n
}
