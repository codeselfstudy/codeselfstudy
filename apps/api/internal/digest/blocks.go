// Package digest builds the Slack digest message and drives the once-per-interval
// posting state machine over the store. It is framework-free.
package digest

import (
	"encoding/json"
	"fmt"
	"strings"

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
// No timestamp is rendered — Slack already stamps every message.
func BuildBlocks(deals []store.Deal) ([]byte, error) {
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
		lines = append(lines, fmt.Sprintf("*<%s|%s>*", escapeLinkURL(stripQueryParams(d.URL)), title))
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

// stripQueryParams drops the query string from a deal URL so the digest posts a
// clean canonical link instead of the long tracking URL newsletters attach
// (utm_*, mcID, linkID, …). Any fragment is preserved: per RFC 3986 the query
// only exists when a "?" precedes the "#", so a "?" that sits inside the
// fragment is left alone and a URL with no query is returned unchanged. This is
// a Slack-display transform only — the raw URL is left intact on the stored
// deal. A plain string split (rather than net/url) is deliberate: the query
// we're discarding often holds unescaped junk like "linkID={$linkID}" that
// url.Parse would reject.
func stripQueryParams(u string) string {
	end := len(u)
	if h := strings.IndexByte(u, '#'); h >= 0 {
		end = h // only the part before the fragment can hold a query
	}
	if i := strings.IndexByte(u[:end], '?'); i >= 0 {
		return u[:i] + u[end:] // drop the query, keep any fragment
	}
	return u
}

// escapeLinkURL percent-encodes the characters that would break Slack's
// <url|text> link syntax. The entity-escaping used for display text is wrong for
// a URL (it would corrupt the path/fragment), so only these three delimiters are
// encoded; everything else in the link is left intact.
var linkURLEscaper = strings.NewReplacer("<", "%3C", ">", "%3E", "|", "%7C")

func escapeLinkURL(u string) string {
	return linkURLEscaper.Replace(u)
}
