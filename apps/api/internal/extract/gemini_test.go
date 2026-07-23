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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/mailparse"
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
