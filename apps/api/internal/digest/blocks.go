// Package digest builds the Slack digest message and drives the once-per-interval
// posting state machine over the store. It is framework-free.
package digest

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

// MaxDealsPerDigest bounds how many deals a single digest renders; the rest stay
// queued for the next one. Slack also limits a message to 50 blocks, and each
// deal uses two (section + divider), so 25 keeps us well under that.
const MaxDealsPerDigest = 25

// Slack Block Kit payload types (only the subset we emit).
type payload struct {
	Blocks []any `json:"blocks"`
}

type textObject struct {
	Type string `json:"type"` // "plain_text" | "mrkdwn"
	Text string `json:"text"`
}

type headerBlock struct {
	Type string     `json:"type"` // "header"
	Text textObject `json:"text"`
}

type sectionBlock struct {
	Type string     `json:"type"` // "section"
	Text textObject `json:"text"`
}

type dividerBlock struct {
	Type string `json:"type"` // "divider"
}

type contextBlock struct {
	Type     string       `json:"type"` // "context"
	Elements []textObject `json:"elements"`
}

// BuildBlocks renders the deals as a Slack incoming-webhook payload. At most
// MaxDealsPerDigest deals are shown; if more are supplied, a context line notes
// how many remain queued. The header count reflects the full set of new deals.
func BuildBlocks(deals []store.Deal, now time.Time) ([]byte, error) {
	shown := deals
	overflow := 0
	if len(shown) > MaxDealsPerDigest {
		overflow = len(shown) - MaxDealsPerDigest
		shown = shown[:MaxDealsPerDigest]
	}

	blocks := []any{
		headerBlock{Type: "header", Text: textObject{Type: "plain_text", Text: headerText(len(deals))}},
	}
	for _, d := range shown {
		blocks = append(blocks,
			sectionBlock{Type: "section", Text: textObject{Type: "mrkdwn", Text: dealText(d)}},
			dividerBlock{Type: "divider"},
		)
	}
	if overflow > 0 {
		blocks = append(blocks, contextBlock{
			Type:     "context",
			Elements: []textObject{{Type: "mrkdwn", Text: fmt.Sprintf("+ %d more — they'll stay queued for the next digest", overflow)}},
		})
	}
	blocks = append(blocks, contextBlock{
		Type:     "context",
		Elements: []textObject{{Type: "mrkdwn", Text: "deal-digest · " + now.UTC().Format("2006-01-02 15:04 MST")}},
	})

	return json.MarshalIndent(payload{Blocks: blocks}, "", "  ")
}

func headerText(n int) string {
	if n == 1 {
		return "1 new deal"
	}
	return fmt.Sprintf("%d new deals", n)
}

// dealText renders one deal as Slack mrkdwn: a bold linked title, a middot-joined
// meta line (price · ends <date> · source), and an italic description. Empty
// fields are omitted.
func dealText(d store.Deal) string {
	var lines []string

	title := escapeMrkdwn(d.Title)
	if d.URL != "" {
		lines = append(lines, fmt.Sprintf("*<%s|%s>*", escapeLinkURL(d.URL), title))
	} else {
		lines = append(lines, "*"+title+"*")
	}

	var meta []string
	if d.Price != "" {
		meta = append(meta, escapeMrkdwn(d.Price))
	}
	if d.EndsAt != "" {
		meta = append(meta, "ends "+escapeMrkdwn(d.EndsAt))
	}
	if d.Source != "" {
		meta = append(meta, escapeMrkdwn(d.Source))
	}
	if len(meta) > 0 {
		lines = append(lines, strings.Join(meta, " · "))
	}

	if d.Description != "" {
		lines = append(lines, "_"+escapeMrkdwn(d.Description)+"_")
	}
	return strings.Join(lines, "\n")
}

// escapeMrkdwn escapes the three characters Slack treats specially in mrkdwn text.
// Ampersand must be replaced first to avoid double-escaping.
func escapeMrkdwn(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// escapeLinkURL percent-encodes the characters that would break Slack's
// <url|text> link syntax. The entity-escaping used for display text is wrong for
// a URL (it would corrupt query strings), so only these three delimiters are
// encoded; everything else in the tracking URL is left intact.
var linkURLEscaper = strings.NewReplacer("<", "%3C", ">", "%3E", "|", "%7C")

func escapeLinkURL(u string) string {
	return linkURLEscaper.Replace(u)
}
