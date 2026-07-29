package expiry

import (
	"testing"
	"time"
)

func page(inner string) []byte {
	return []byte("<!doctype html><html><head>" + inner + "</head><body><p>deal</p></body></html>")
}

func TestFromHTMLJSONLDObject(t *testing.T) {
	got := FromHTML(page(`<script type="application/ld+json">
		{"@type":"Product","offers":{"@type":"Offer","price":"25","priceValidUntil":"2026-08-01"}}
	</script>`))
	if got != "2026-08-01" {
		t.Errorf("FromHTML = %q, want 2026-08-01", got)
	}
}

func TestFromHTMLJSONLDArrayAndGraph(t *testing.T) {
	got := FromHTML(page(`<script type="application/ld+json">
		{"@graph":[{"@type":"WebSite"},{"@type":"Product","offers":[{"availabilityEnds":"2026-08-15T23:59:00Z"}]}]}
	</script>`))
	if got != "2026-08-15" {
		t.Errorf("FromHTML = %q, want 2026-08-15 (datetime reduced to its date)", got)
	}
}

func TestFromHTMLPriceValidUntilBeatsAvailabilityEnds(t *testing.T) {
	got := FromHTML(page(`<script type="application/ld+json">
		{"availabilityEnds":"2026-09-01","priceValidUntil":"2026-08-01"}
	</script>`))
	if got != "2026-08-01" {
		t.Errorf("FromHTML = %q, want priceValidUntil to win", got)
	}
}

func TestFromHTMLBrokenBlockDoesNotSpoilNext(t *testing.T) {
	got := FromHTML(page(`<script type="application/ld+json">{not json</script>` +
		`<script type="application/ld+json">{"priceValidUntil":"2026-08-02"}</script>`))
	if got != "2026-08-02" {
		t.Errorf("FromHTML = %q, want the second (valid) block's date", got)
	}
}

func TestFromHTMLMetaItemprop(t *testing.T) {
	got := FromHTML(page(`<meta itemprop="priceValidUntil" content="2026-08-03">`))
	if got != "2026-08-03" {
		t.Errorf("FromHTML = %q, want 2026-08-03", got)
	}
}

func TestFromHTMLJSONLDWinsOverMeta(t *testing.T) {
	got := FromHTML(page(`<meta itemprop="priceValidUntil" content="2026-09-09">` +
		`<script type="application/ld+json">{"priceValidUntil":"2026-08-04"}</script>`))
	if got != "2026-08-04" {
		t.Errorf("FromHTML = %q, want the JSON-LD date", got)
	}
}

func TestFromHTMLUnparseableDateSkipped(t *testing.T) {
	got := FromHTML(page(`<script type="application/ld+json">{"priceValidUntil":"next Tuesday"}</script>`))
	if got != "" {
		t.Errorf("FromHTML = %q, want empty for an unparseable date", got)
	}
}

func TestFromHTMLNoMarkers(t *testing.T) {
	if got := FromHTML(page(`<title>Just a page</title>`)); got != "" {
		t.Errorf("FromHTML = %q, want empty", got)
	}
	if got := FromHTML(nil); got != "" {
		t.Errorf("FromHTML(nil) = %q, want empty", got)
	}
}

func TestMinePrice(t *testing.T) {
	cases := []struct{ name, inner, want string }{
		{
			"string price with USD",
			`<script type="application/ld+json">{"offers":{"@type":"Offer","price":"25.00","priceCurrency":"USD"}}</script>`,
			"$25",
		},
		{
			"number price with EUR",
			`<script type="application/ld+json">{"offers":{"price":9.99,"priceCurrency":"EUR"}}</script>`,
			"€9.99",
		},
		{
			"aggregate lowPrice",
			`<script type="application/ld+json">{"offers":{"@type":"AggregateOffer","lowPrice":"9.99","highPrice":"49.99","priceCurrency":"USD"}}</script>`,
			"from $9.99",
		},
		{
			"unabbreviated currency",
			`<script type="application/ld+json">{"offers":{"price":"30","priceCurrency":"CAD"}}</script>`,
			"30 CAD",
		},
		{
			"missing currency rejected",
			`<script type="application/ld+json">{"offers":{"price":"25"}}</script>`,
			"",
		},
		{
			"zero price rejected",
			`<script type="application/ld+json">{"offers":{"price":"0","priceCurrency":"USD"}}</script>`,
			"",
		},
		{
			"non-numeric price rejected",
			`<script type="application/ld+json">{"offers":{"price":"cheap","priceCurrency":"USD"}}</script>`,
			"",
		},
		{
			"itemprop microdata fallback",
			`<meta itemprop="price" content="12.50"><meta itemprop="priceCurrency" content="GBP">`,
			"£12.50",
		},
		{
			"json-ld wins over itemprop",
			`<meta itemprop="price" content="99"><meta itemprop="priceCurrency" content="USD">` +
				`<script type="application/ld+json">{"offers":{"price":"25","priceCurrency":"USD"}}</script>`,
			"$25",
		},
		{"no markers", `<title>Just a page</title>`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Mine(page(tc.inner)).Price; got != tc.want {
				t.Errorf("Mine().Price = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMineBothFields(t *testing.T) {
	// One offer block carrying both a deadline and a price fills both fields
	// from a single parse.
	got := Mine(page(`<script type="application/ld+json">
		{"@type":"Product","offers":{"price":"25","priceCurrency":"USD","priceValidUntil":"2026-08-01"}}
	</script>`))
	if got.EndsAt != "2026-08-01" || got.Price != "$25" {
		t.Errorf("Mine = %+v, want EndsAt 2026-08-01 and Price $25", got)
	}
}

func TestMineFieldsFromSeparateBlocks(t *testing.T) {
	// A price-only block must not stop the date search in a later block.
	got := Mine(page(`<script type="application/ld+json">{"offers":{"price":"25","priceCurrency":"USD"}}</script>` +
		`<script type="application/ld+json">{"priceValidUntil":"2026-08-05"}</script>`))
	if got.EndsAt != "2026-08-05" || got.Price != "$25" {
		t.Errorf("Mine = %+v, want both fields filled across blocks", got)
	}
}

func TestNormalizeDate(t *testing.T) {
	cases := []struct{ in, want string }{
		{"2026-08-01", "2026-08-01"},
		{"2026-08-01T10:00:00Z", "2026-08-01"},
		{"2026-08-01T10:00:00-07:00", "2026-08-01"},
		{"2026-08-01T10:00:00", "2026-08-01"},
		{"2026-08-01 10:00:00", "2026-08-01"},
		{"  2026-08-01  ", "2026-08-01"},
		{"08/01/2026", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := normalizeDate(tc.in); got != tc.want {
			t.Errorf("normalizeDate(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestOnOrAfter(t *testing.T) {
	ref := time.Date(2026, 7, 26, 15, 30, 0, 0, time.UTC)
	cases := []struct {
		name, in string
		want     bool
	}{
		{"past day", "2025-11-27", false},
		{"day before", "2026-07-25", false},
		{"same day", "2026-07-26", true},
		{"same day despite earlier timestamp", "2026-07-26T01:00:00Z", true},
		{"future day", "2026-11-27", true},
		{"unparseable is not rejected", "while supplies last", true},
		{"empty is not rejected", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := OnOrAfter(tc.in, ref); got != tc.want {
				t.Errorf("OnOrAfter(%q, %s) = %v, want %v", tc.in, ref, got, tc.want)
			}
		})
	}
}
