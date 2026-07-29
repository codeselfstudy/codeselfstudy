// Package expiry mines a deal page for its expiration date and price.
// Newsletters rarely print a deadline, but the deal's landing page usually
// carries one as structured data — schema.org offers embed it as
// priceValidUntil (or availabilityEnds) in JSON-LD, and some pages use
// itemprop microdata instead; the same offer blocks carry price and
// priceCurrency. Extraction is best-effort and purely local: malformed HTML,
// broken JSON, or an unparseable value all yield "", never an error. The
// package also owns the deadline plausibility check (OnOrAfter) that ingest
// and the digest use to catch confidently-wrong extracted dates.
package expiry

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// dateKeys are the schema.org offer fields that carry an end date, in
// preference order.
var dateKeys = []string{"priceValidUntil", "availabilityEnds"}

// priceKeys are the schema.org offer fields that carry a price, in preference
// order: a concrete price first, then an AggregateOffer's lower bound.
var priceKeys = []string{"price", "lowPrice"}

// Result holds what a deal page's structured data yielded. An empty field
// means the page stated nothing usable for it.
type Result struct {
	EndsAt string // "2006-01-02", ready for deals.ends_at
	Price  string // display-ready, e.g. "$25" or "from €9.99"
}

// FromHTML returns the page's offer expiration as a "2006-01-02" date string,
// or "" when none is found. It is Mine for callers that only want the date.
func FromHTML(page []byte) string { return Mine(page).EndsAt }

// Mine extracts the offer expiration and price from one parse of the page.
// For each field JSON-LD wins over itemprop microdata; within each, the first
// parseable value wins.
func Mine(page []byte) Result {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return Result{}
	}

	var jsonLD []string // raw <script type="application/ld+json"> payloads
	var metaDates []string
	var metaPrice, metaCurrency string
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
					if strings.EqualFold(prop, "price") && metaPrice == "" {
						metaPrice = attr(n, "content")
					}
					if strings.EqualFold(prop, "priceCurrency") && metaCurrency == "" {
						metaCurrency = attr(n, "content")
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	var res Result
	for _, raw := range jsonLD {
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			continue // one broken block must not spoil the others
		}
		if res.EndsAt == "" {
			res.EndsAt = findDate(v)
		}
		if res.Price == "" {
			res.Price = findPrice(v)
		}
		if res.EndsAt != "" && res.Price != "" {
			return res
		}
	}
	if res.EndsAt == "" {
		for _, c := range metaDates {
			if d := normalizeDate(c); d != "" {
				res.EndsAt = d
				break
			}
		}
	}
	if res.Price == "" {
		res.Price = renderPrice(metaPrice, metaCurrency, false)
	}
	return res
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

// findPrice walks decoded JSON-LD (objects, arrays, @graph nesting) for the
// first priceKeys value that renders as a price. Only a value sitting next to
// a priceCurrency is trusted — a bare number could be anything.
func findPrice(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range priceKeys {
			if raw, ok := t[k]; ok {
				if p := renderPrice(raw, t["priceCurrency"], k == "lowPrice"); p != "" {
					return p
				}
			}
		}
		for _, child := range t {
			if p := findPrice(child); p != "" {
				return p
			}
		}
	case []any:
		for _, child := range t {
			if p := findPrice(child); p != "" {
				return p
			}
		}
	}
	return ""
}

// currencySymbols maps the ISO 4217 codes worth abbreviating; anything else is
// rendered as "25 CAD".
var currencySymbols = map[string]string{"USD": "$", "EUR": "€", "GBP": "£"}

// renderPrice turns a schema.org price value (JSON string or number) and its
// priceCurrency into a display string like "$25", "€9.99", or "25 CAD" —
// "from …" when the value is an AggregateOffer's lowPrice. It returns "" for
// anything unusable: a non-numeric value, a missing currency, or a zero price
// (more often a placeholder than a real offer).
func renderPrice(raw, currency any, low bool) string {
	var f float64
	switch v := raw.(type) {
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return ""
		}
		f = parsed
	case float64:
		f = v
	default:
		return ""
	}
	if f <= 0 {
		return ""
	}
	cur, _ := currency.(string)
	cur = strings.ToUpper(strings.TrimSpace(cur))
	if cur == "" {
		return ""
	}

	amount := strconv.FormatFloat(f, 'f', 2, 64)
	if f == math.Trunc(f) {
		amount = strconv.FormatFloat(f, 'f', 0, 64)
	}
	price := amount + " " + cur
	if sym, ok := currencySymbols[cur]; ok {
		price = sym + amount
	}
	if low {
		price = "from " + price
	}
	return price
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

// Normalize renders any accepted date shape as a "2006-01-02" date string, or
// "" when s does not parse. Callers storing a model-supplied date use it to
// reject free text ("soon", "next week") before OnOrAfter — which deliberately
// lets unparseable values through — can bless it.
func Normalize(s string) string { return normalizeDate(s) }

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
