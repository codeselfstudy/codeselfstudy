package expiry

import "testing"

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
