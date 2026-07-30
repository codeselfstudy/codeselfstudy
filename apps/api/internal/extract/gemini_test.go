package extract

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/digest"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/mailparse"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

var ctx = context.Background()

var sampleEmail = mailparse.Email{
	MessageID: "<x@humblebundle.com>",
	From:      "Humble Bundle <deals@humblebundle.com>",
	Subject:   "This week's deals",
	Text:      "Humble Programming Bundle — pay what you want.",
}

// geminiEnvelope wraps model output text in the API's response shape.
func geminiEnvelope(text string) []byte {
	b, _ := json.Marshal(map[string]any{
		"candidates": []map[string]any{
			{
				"content":      map[string]any{"role": "model", "parts": []map[string]any{{"text": text}}},
				"finishReason": "STOP",
			},
		},
	})
	return b
}

// newTestGemini points a Gemini extractor at an httptest server with fast retries.
func newTestGemini(t *testing.T, handler http.HandlerFunc) *Gemini {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	g, err := NewGemini(ctx, "test-key", "gemini-3.5-flash-lite", srv.URL)
	if err != nil {
		t.Fatalf("NewGemini: %v", err)
	}
	g.RetryDelays = []time.Duration{time.Millisecond, time.Millisecond}
	return g
}

func TestExtractRecordedResponse(t *testing.T) {
	recorded, err := os.ReadFile(filepath.Join("testdata", "gemini_response.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(geminiEnvelope(string(recorded)))
	})

	deals, err := g.Extract(ctx, sampleEmail)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(deals) != 2 {
		t.Fatalf("got %d deals, want 2", len(deals))
	}
	if deals[0].Title != "Humble Programming Bundle" || deals[0].URL != "https://humblebundle.com/books/prog" {
		t.Errorf("deal[0] = %+v", deals[0])
	}
	if deals[0].Price != "$25 (96% off)" || deals[0].EndsAt != "2026-07-27" {
		t.Errorf("deal[0] price/ends = %+v", deals[0])
	}
	if deals[1].URL != "" {
		t.Errorf("deal[1] should have empty URL, got %q", deals[1].URL)
	}
}

func TestExtractEmptyArray(t *testing.T) {
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(geminiEnvelope("[]"))
	})
	deals, err := g.Extract(ctx, sampleEmail)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(deals) != 0 {
		t.Errorf("got %d deals, want 0", len(deals))
	}
}

func TestExtractRetriesThenSucceeds(t *testing.T) {
	var calls int32
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":429,"message":"rate limited","status":"RESOURCE_EXHAUSTED"}}`))
			return
		}
		w.Write(geminiEnvelope("[]"))
	})

	if _, err := g.Extract(ctx, sampleEmail); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("made %d calls, want 3 (2 retries)", got)
	}
}

func TestExtractExhaustsRetries(t *testing.T) {
	var calls int32
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":{"code":503,"message":"unavailable"}}`))
	})

	if _, err := g.Extract(ctx, sampleEmail); err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if got := atomic.LoadInt32(&calls); got != int32(g.MaxAttempts) {
		t.Errorf("made %d calls, want %d", got, g.MaxAttempts)
	}
}

func TestExtractMalformedJSONNotRetried(t *testing.T) {
	var calls int32
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Write(geminiEnvelope("this is not json"))
	})

	if _, err := g.Extract(ctx, sampleEmail); err == nil {
		t.Fatal("expected parse error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("made %d calls, want 1 (no retry on parse error)", got)
	}
}

func TestExtractEmptyCandidates(t *testing.T) {
	// A safety-blocked or zero-candidate response must not panic and yields no deals.
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"candidates":[]}`))
	})
	deals, err := g.Extract(ctx, sampleEmail)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(deals) != 0 {
		t.Errorf("got %d deals, want 0", len(deals))
	}
}

func TestExtractSendsSchemaAndSystemInstruction(t *testing.T) {
	var body []byte
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Write(geminiEnvelope("[]"))
	})
	if _, err := g.Extract(ctx, sampleEmail); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	s := string(body)
	for _, want := range []string{"responseSchema", "systemInstruction", "responseMimeType"} {
		if !strings.Contains(s, want) {
			t.Errorf("request body missing %q\n%s", want, s)
		}
	}
}

func TestEnrichFromPage(t *testing.T) {
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(geminiEnvelope(`{"ends_at":"2026-08-20","price":"$19"}`))
	})
	sent := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	en, err := g.EnrichFromPage(ctx, "Bundle A", "Sale ends August 20. Get it for $19.", &sent)
	if err != nil {
		t.Fatalf("EnrichFromPage: %v", err)
	}
	if en.EndsAt != "2026-08-20" || en.Price != "$19" {
		t.Errorf("Enrichment = %+v, want ends 2026-08-20 and $19", en)
	}
}

func TestEnrichFromPageNulls(t *testing.T) {
	// The schema's nullable fields come back as JSON nulls when the page
	// states nothing; they must land as empty strings, not an error.
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(geminiEnvelope(`{"ends_at":null,"price":null}`))
	})
	en, err := g.EnrichFromPage(ctx, "Bundle A", "A page with no offer facts.", nil)
	if err != nil {
		t.Fatalf("EnrichFromPage: %v", err)
	}
	if en != (Enrichment{}) {
		t.Errorf("Enrichment = %+v, want zero", en)
	}
}

func TestEnrichFromPageMalformedJSON(t *testing.T) {
	var calls int32
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Write(geminiEnvelope("not json"))
	})
	if _, err := g.EnrichFromPage(ctx, "Bundle A", "text", nil); err == nil {
		t.Fatal("expected parse error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("made %d calls, want 1 (no retry on parse error)", got)
	}
}

func TestBuildEnrichPrompt(t *testing.T) {
	sent := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	got := buildEnrichPrompt("Bundle A", "page text here", &sent)
	want := "Date: 2026-07-20\nDeal: Bundle A\n\npage text here"
	if got != want {
		t.Errorf("buildEnrichPrompt = %q, want %q", got, want)
	}
	if withoutDate := buildEnrichPrompt("Bundle A", "x", nil); strings.Contains(withoutDate, "Date:") {
		t.Errorf("prompt without sentAt = %q, want no Date line", withoutDate)
	}
}

func TestBuildEnrichPromptTruncates(t *testing.T) {
	long := strings.Repeat("x", maxPromptChars+500)
	got := buildEnrichPrompt("Bundle A", long, nil)
	if len([]rune(got)) > maxPromptChars+100 {
		t.Errorf("prompt length = %d runes, want the page text capped at %d", len([]rune(got)), maxPromptChars)
	}
}

func TestDisabledExtractor(t *testing.T) {
	deals, err := Disabled{}.Extract(ctx, sampleEmail)
	if !errors.Is(err, ErrNotConfigured) {
		t.Errorf("err = %v, want ErrNotConfigured", err)
	}
	if deals != nil {
		t.Errorf("deals = %v, want nil", deals)
	}
}

func TestParseDeals(t *testing.T) {
	if d, err := parseDeals("  "); err != nil || len(d) != 0 {
		t.Errorf("blank: deals=%v err=%v", d, err)
	}
	if d, err := parseDeals("[]"); err != nil || len(d) != 0 {
		t.Errorf("empty array: deals=%v err=%v", d, err)
	}
	if _, err := parseDeals("{bad"); err == nil {
		t.Error("expected error for malformed json")
	}
}

func TestBuildUserPromptDateLine(t *testing.T) {
	sent := time.Date(2026, 7, 26, 15, 29, 0, 0, time.UTC)
	e := mailparse.Email{From: "a@b.com", Subject: "S", Text: "body", SentAt: &sent}
	got := buildUserPrompt(e)
	if !strings.HasPrefix(got, "Date: 2026-07-26\n") {
		t.Errorf("prompt missing Date line:\n%s", got)
	}

	// Without a parseable Date header there is no line to anchor on — the
	// prompt must simply omit it rather than invent one.
	e.SentAt = nil
	if got := buildUserPrompt(e); strings.Contains(got, "Date:") {
		t.Errorf("prompt has a Date line without SentAt:\n%s", got)
	}
}

func TestBuildUserPromptTruncates(t *testing.T) {
	long := make([]rune, maxPromptChars+5000)
	for i := range long {
		long[i] = 'a'
	}
	e := mailparse.Email{From: "a@b.com", Subject: "S", Text: string(long)}
	got := buildUserPrompt(e)
	// Header lines + at most maxPromptChars of body.
	if len([]rune(got)) > maxPromptChars+100 {
		t.Errorf("prompt not truncated: %d runes", len([]rune(got)))
	}
}

// TestExtractLive hits the real Gemini API; run with GEMINI_LIVE_TEST=1 and a key.
// It exercises the software-only filter: a programming bundle yields deals, a game
// bundle yields none.
func TestExtractLive(t *testing.T) {
	if os.Getenv("GEMINI_LIVE_TEST") != "1" {
		t.Skip("set GEMINI_LIVE_TEST=1 and GEMINI_API_KEY to run")
	}
	g, err := NewGemini(ctx, os.Getenv("GEMINI_API_KEY"), "gemini-3.5-flash-lite", "")
	if err != nil {
		t.Fatalf("NewGemini: %v", err)
	}

	tests := []struct {
		name     string
		email    mailparse.Email
		wantDeal bool // true: expect >=1 deal; false: expect 0
	}{
		{
			name: "software book bundle kept",
			email: mailparse.Email{
				From:    "Humble Bundle <deals@humblebundle.com>",
				Subject: "New bundle",
				Text:    "Humble Python Bundle: 12 ebooks, pay what you want, ends July 30. https://humblebundle.com/books/python",
			},
			wantDeal: true,
		},
		{
			name: "game/comic bundle dropped",
			email: mailparse.Email{
				From:    "Humble Bundle <deals@humblebundle.com>",
				Subject: "New bundle",
				Text:    "Teenage Mutant Ninja Turtles Humble Bundle: 20 games and comics, pay what you want, ends July 30. https://humblebundle.com/games/tmnt",
			},
			wantDeal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deals, err := g.Extract(ctx, tt.email)
			if err != nil {
				t.Fatalf("Extract: %v", err)
			}
			t.Logf("got %d deals: %+v", len(deals), deals)
			if tt.wantDeal && len(deals) == 0 {
				t.Errorf("expected at least one software deal, got none")
			}
			if !tt.wantDeal && len(deals) != 0 {
				t.Errorf("expected no deals, got %d: %+v", len(deals), deals)
			}
		})
	}
}

// TestSystemInstructionScopesToSoftware locks the filter's intent — software in,
// games out — so the prompt can't silently regress. The model-based behaviour
// itself is verified by TestExtractLive.
func TestSystemInstructionScopesToSoftware(t *testing.T) {
	lower := strings.ToLower(systemInstruction)
	for _, want := range []string{"software", "game"} {
		if !strings.Contains(lower, want) {
			t.Errorf("systemInstruction missing %q", want)
		}
	}
}

// TestSystemInstructionCollapsesPromotions locks the overlap rule — one
// promotion spanning categories yields one umbrella entry, and a product
// merely carrying that promotion's discount gets no entry of its own — so the
// prompt can't silently regress. The model-based behaviour itself is verified
// by TestExtractLiveCollapsesPromotions.
func TestSystemInstructionCollapsesPromotions(t *testing.T) {
	lower := strings.ToLower(systemInstruction)
	for _, want := range []string{"promotion", "category or tier", "genuinely distinct"} {
		if !strings.Contains(lower, want) {
			t.Errorf("systemInstruction missing %q", want)
		}
	}
}

// TestCondenseSystemInstruction locks the condense rules — merge rewordings,
// drop sale-priced products, suppress already-announced offers, cite only
// given ids — so the prompt can't silently regress. The model behaviour is
// verified by TestCondenseLive.
func TestCondenseSystemInstruction(t *testing.T) {
	lower := strings.ToLower(condenseSystemInstruction)
	for _, want := range []string{"queued", "announced", "merge", "cite only ids", "never invent"} {
		if !strings.Contains(lower, want) {
			t.Errorf("condenseSystemInstruction missing %q", want)
		}
	}
}

func TestBuildCondensePrompt(t *testing.T) {
	queued := []store.Deal{{ID: 7, Source: "Manning", Title: "Sale A", Price: "50% off", EndsAt: "2026-07-31", Description: "line one\nline two"}}
	recent := []store.Deal{{ID: 3, Source: "Manning", Title: "Old Sale"}}
	got := buildCondensePrompt(queued, recent)
	for _, want := range []string{
		"QUEUED:\n7 | Manning | Sale A | 50% off | 2026-07-31 | line one line two\n",
		"ANNOUNCED:\n3 | Manning | Old Sale |  |  | \n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt missing %q:\n%s", want, got)
		}
	}
}

func TestCondenseParsesResponse(t *testing.T) {
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(geminiEnvelope(`[{"source":"Manning","bullets":[{"deal_id":7,"text":"50% off all books"},{"deal_id":9,"text":"Free webinar: The Rust Journey"}]}]`))
	})
	groups, err := g.Condense(ctx, []store.Deal{{ID: 7}, {ID: 9}}, nil)
	if err != nil {
		t.Fatalf("Condense: %v", err)
	}
	want := []digest.CondensedGroup{{
		Source: "Manning",
		Bullets: []digest.Bullet{
			{DealID: 7, Text: "50% off all books"},
			{DealID: 9, Text: "Free webinar: The Rust Journey"},
		},
	}}
	if !reflect.DeepEqual(groups, want) {
		t.Errorf("Condense = %+v, want %+v", groups, want)
	}
}

func TestCondenseMalformedJSONNotRetried(t *testing.T) {
	var calls int32
	g := newTestGemini(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Write(geminiEnvelope("not json"))
	})
	if _, err := g.Condense(ctx, []store.Deal{{ID: 1}}, nil); err == nil {
		t.Fatal("expected parse error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("made %d calls, want 1 (no retry on parse error)", got)
	}
}

// TestExtractLiveCollapsesPromotions hits the real Gemini API; run with
// GEMINI_LIVE_TEST=1 and a key. A Manning-shaped email — one sitewide sale
// restated per category, a book merely at the sale's own discount, and one
// genuinely distinct free webinar — must collapse to exactly two deals: the
// umbrella promotion and the webinar. The sale-priced book must NOT get its
// own entry.
func TestExtractLiveCollapsesPromotions(t *testing.T) {
	if os.Getenv("GEMINI_LIVE_TEST") != "1" {
		t.Skip("set GEMINI_LIVE_TEST=1 and GEMINI_API_KEY to run")
	}
	g, err := NewGemini(ctx, os.Getenv("GEMINI_API_KEY"), "gemini-3.5-flash-lite", "")
	if err != nil {
		t.Fatalf("NewGemini: %v", err)
	}

	email := mailparse.Email{
		From:    "Manning Publications <promo@manning.com>",
		Subject: "Save half sitewide + a free Rust webinar",
		Text: `Our summer sale is on: save half sitewide on all Manning books, get any liveProject or liveVideo for $10, or a year of Manning Online Pro for $199.99. Sale ends July 31. https://www.manning.com/sale
All liveProjects and liveVideos $10! https://www.manning.com/liveprojects
Featured today at 50% off: Build a Reasoning Model (From Scratch), $23.99 for eBook. https://www.manning.com/books/build-a-reasoning-model-from-scratch
Free live webinar: The Rust Journey — Exploring Safe Systems Programming. https://www.manning.com/webinars/rust-journey`,
	}

	deals, err := g.Extract(ctx, email)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	t.Logf("got %d deals: %+v", len(deals), deals)
	if len(deals) != 2 {
		t.Errorf("expected exactly 2 deals (umbrella sale + free webinar), got %d", len(deals))
	}
	for _, d := range deals {
		if strings.Contains(strings.ToLower(d.Title), "reasoning") {
			t.Errorf("sale-priced book got its own entry: %+v", d)
		}
	}
	webinars := 0
	for _, d := range deals {
		if strings.Contains(strings.ToLower(d.Title), "rust") {
			webinars++
		}
	}
	if webinars != 1 {
		t.Errorf("expected exactly one entry for the free webinar, got %d", webinars)
	}
}

// TestCondenseLive hits the real Gemini API; run with GEMINI_LIVE_TEST=1 and a
// key. A screenshot-shaped queue — one sale under four wordings, one $10 promo
// under three, sale-priced books, and one free webinar — must condense to a
// handful of Manning bullets citing only queued ids, with no bullet for the
// sale-priced books.
func TestCondenseLive(t *testing.T) {
	if os.Getenv("GEMINI_LIVE_TEST") != "1" {
		t.Skip("set GEMINI_LIVE_TEST=1 and GEMINI_API_KEY to run")
	}
	g, err := NewGemini(ctx, os.Getenv("GEMINI_API_KEY"), "gemini-3.5-flash-lite", "")
	if err != nil {
		t.Fatalf("NewGemini: %v", err)
	}

	src := "Manning Publications"
	queued := []store.Deal{
		{ID: 1, Source: src, Title: "Manning Pro Annual Subscription", Price: "$199.99 (was $249.99)", EndsAt: "2026-07-31"},
		{ID: 2, Source: src, Title: "All books (including MEAP)", Price: "50% off", EndsAt: "2026-07-31"},
		{ID: 3, Source: src, Title: "All liveVideos and liveProjects", Price: "$10 each", EndsAt: "2026-07-31"},
		{ID: 4, Source: src, Title: "Books and MEAP Sale", Price: "50% off", EndsAt: "2026-07-31"},
		{ID: 5, Source: src, Title: "liveVideos and liveProjects", Price: "$10 each", EndsAt: "2026-07-31"},
		{ID: 6, Source: src, Title: "Architecting for Autonomy", Price: "$27.99 (eBook, 50% off)", EndsAt: "2026-07-31"},
		{ID: 7, Source: src, Title: "The Rust Journey – Exploring Safe Systems Programming", Price: "Free", EndsAt: "2026-07-31", Description: "Free live webinar exploring Rust and safe systems programming."},
		{ID: 8, Source: src, Title: "Manning Storewide Book Sale", Price: "50% off", EndsAt: "2026-07-31"},
		{ID: 9, Source: src, Title: "Manning liveVideos and liveProjects", Price: "$10 each", EndsAt: "2026-07-31"},
		{ID: 10, Source: src, Title: "Summer Sale Sitewide Book Discount", Price: "50% off", EndsAt: "2026-07-31"},
		{ID: 11, Source: src, Title: "Build a Reasoning Model (From Scratch)", Price: "$23.99 eBook (50% off)", EndsAt: "2026-07-31"},
		{ID: 12, Source: src, Title: "AI Agents and Applications", Price: "$23.99 eBook (50% off)", EndsAt: "2026-07-31"},
		{ID: 13, Source: src, Title: "Build an AI Agent (From Scratch)", Price: "50% off", EndsAt: "2026-07-31"},
	}

	groups, err := g.Condense(ctx, queued, nil)
	if err != nil {
		t.Fatalf("Condense: %v", err)
	}
	t.Logf("got %+v", groups)
	known := map[int64]bool{}
	for _, d := range queued {
		known[d.ID] = true
	}
	bullets := 0
	sawWebinar := false
	for _, g := range groups {
		for _, b := range g.Bullets {
			bullets++
			if !known[b.DealID] {
				t.Errorf("bullet cites unknown id %d: %+v", b.DealID, b)
			}
			if strings.Contains(strings.ToLower(b.Text), "rust") {
				sawWebinar = true
			}
		}
	}
	if bullets < 2 || bullets > 5 {
		t.Errorf("got %d bullets, want the 13 deals condensed to roughly 3-4", bullets)
	}
	if !sawWebinar {
		t.Errorf("expected a bullet for the free Rust webinar")
	}
}
