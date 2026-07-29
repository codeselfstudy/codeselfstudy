// Package extract turns a parsed email into structured deals. Extraction is
// LLM-based (via the Gemini API) rather than per-site parsing, so any newsletter
// works and non-deal emails yield an empty slice.
package extract

import (
	"context"
	"errors"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/mailparse"
)

// Deal is one offer extracted from an email. Optional fields may be empty.
type Deal struct {
	Source      string `json:"source"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Price       string `json:"price"`
	EndsAt      string `json:"ends_at"`
	Description string `json:"description"`
}

// Extractor extracts deals from a parsed email. Implementations must return an
// empty slice (not an error) for emails that contain no deals.
type Extractor interface {
	Extract(ctx context.Context, e mailparse.Email) ([]Deal, error)
}

// Enrichment is what a deal page's text yields for fields the email and the
// page's structured data both left empty. An empty field means the page stated
// nothing usable for it.
type Enrichment struct {
	EndsAt string `json:"ends_at"`
	Price  string `json:"price"`
}

// PageEnricher extracts a deal's still-missing fields from its landing page's
// text. sentAt anchors yearless deadlines the way Extract's Date line does.
// Implementations must report only what the page states — never invent.
type PageEnricher interface {
	EnrichFromPage(ctx context.Context, title, pageText string, sentAt *time.Time) (Enrichment, error)
}

// ErrNotConfigured is returned by Disabled.Extract.
var ErrNotConfigured = errors.New("gemini extractor not configured (set GEMINI_API_KEY)")

// Disabled is an Extractor used when no API key is configured. It lets the server
// boot (so /healthz stays up) while every extraction attempt fails loudly, rather
// than crash-looping at startup.
type Disabled struct{}

// Extract always returns ErrNotConfigured.
func (Disabled) Extract(context.Context, mailparse.Email) ([]Deal, error) {
	return nil, ErrNotConfigured
}
