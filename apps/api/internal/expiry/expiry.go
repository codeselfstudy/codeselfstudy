// Package expiry mines a deal page for its expiration date. Newsletters rarely
// print a deadline, but the deal's landing page usually carries one as
// structured data — schema.org offers embed it as priceValidUntil (or
// availabilityEnds) in JSON-LD, and some pages use itemprop microdata instead.
// Extraction is best-effort and purely local: malformed HTML, broken JSON, or
// an unparseable date all yield "", never an error. The package also owns the
// deadline plausibility check (OnOrAfter) that ingest and the digest use to
// catch confidently-wrong extracted dates.
package expiry

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// dateKeys are the schema.org offer fields that carry an end date, in
// preference order.
var dateKeys = []string{"priceValidUntil", "availabilityEnds"}

// FromHTML returns the page's offer expiration as a "2006-01-02" date string,
// or "" when none is found. JSON-LD wins over itemprop microdata; within each,
// the first parseable date wins.
func FromHTML(page []byte) string {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return ""
	}

	var jsonLD []string // raw <script type="application/ld+json"> payloads
	var metaDates []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script":
				if strings.EqualFold(attr(n, "type"), "application/ld+json") && n.FirstChild != nil {
					jsonLD = append(jsonLD, n.FirstChild.Data)
				}
			case "meta":
				if prop := attr(n, "itemprop"); prop != "" {
					for _, k := range dateKeys {
						if strings.EqualFold(prop, k) {
							metaDates = append(metaDates, attr(n, "content"))
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	for _, raw := range jsonLD {
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			continue // one broken block must not spoil the others
		}
		if d := findDate(v); d != "" {
			return d
		}
	}
	for _, c := range metaDates {
		if d := normalizeDate(c); d != "" {
			return d
		}
	}
	return ""
}

// findDate walks decoded JSON-LD (objects, arrays, @graph nesting) for the
// first dateKeys value that parses as a date.
func findDate(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range dateKeys {
			if s, ok := t[k].(string); ok {
				if d := normalizeDate(s); d != "" {
					return d
				}
			}
		}
		for _, child := range t {
			if d := findDate(child); d != "" {
				return d
			}
		}
	case []any:
		for _, child := range t {
			if d := findDate(child); d != "" {
				return d
			}
		}
	}
	return ""
}

// dateLayouts are the shapes schema.org dates show up in, tried in order.
var dateLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// normalizeDate parses a candidate and renders it date-only. The digest shows
// this string verbatim as "ends <date>", so a bare date beats a full RFC 3339
// timestamp; the timestamp's own calendar day is kept, without time-zone
// conversion — close enough for a deal deadline.
func normalizeDate(s string) string {
	t, ok := parseDate(s)
	if !ok {
		return ""
	}
	return t.Format("2006-01-02")
}

// parseDate tries the dateLayouts in order on a trimmed candidate.
func parseDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// OnOrAfter reports whether s names a calendar day on or after ref's. It is the
// guard against impossible deadlines — a stale year in the vendor's own copy,
// or an extractor that guesses a year for a yearless date — either way the
// deadline lands in the past: only a value that parses and falls strictly
// before ref's own day returns false. Empty or unparseable strings return true —
// they carry nothing confidently wrong to reject. Days are compared without
// time-zone conversion, matching normalizeDate's stance.
func OnOrAfter(s string, ref time.Time) bool {
	t, ok := parseDate(s)
	if !ok {
		return true
	}
	ty, tm, td := t.Date()
	ry, rm, rd := ref.Date()
	day := time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	refDay := time.Date(ry, rm, rd, 0, 0, 0, 0, time.UTC)
	return !day.Before(refDay)
}

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val
		}
	}
	return ""
}
