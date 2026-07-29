package ingest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/db"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/extract"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/ingest"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/mailparse"
	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

const token = "secret-token"

type spyExtractor struct {
	calls int32
	mu    sync.Mutex
	deals []extract.Deal
	err   error
}

func (s *spyExtractor) Extract(_ context.Context, _ mailparse.Email) ([]extract.Deal, error) {
	atomic.AddInt32(&s.calls, 1)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deals, s.err
}

func (s *spyExtractor) set(deals []extract.Deal, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deals, s.err = deals, err
}

type fakePoster struct {
	mu  sync.Mutex
	n   int
	err error
}

func (f *fakePoster) Post(_ context.Context, _ []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.n++
	return nil
}

func (f *fakePoster) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.n
}

// fakeResolver maps input URLs to resolved ones; unmapped inputs pass through
// unchanged, and err (if set) is returned alongside — matching the URLResolver
// contract of always handing back a usable URL. ResolvePage additionally hands
// back the canned page (nil for an empty URL, like the real resolver).
type fakeResolver struct {
	mu           sync.Mutex
	byURL        map[string]string
	page         []byte
	err          error
	resolveCalls int
	pageCalls    int
}

func (f *fakeResolver) Resolve(_ context.Context, raw string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resolveCalls++
	return f.mapURL(raw), f.err
}

func (f *fakeResolver) ResolvePage(_ context.Context, raw string) (string, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pageCalls++
	if raw == "" {
		return raw, nil, nil
	}
	return f.mapURL(raw), f.page, f.err
}

// mapURL requires f.mu held.
func (f *fakeResolver) mapURL(raw string) string {
	if r, ok := f.byURL[raw]; ok {
		return r
	}
	return raw
}

type env struct {
	e      *echo.Echo
	h      *ingest.Handlers
	st     *store.Store
	ex     *spyExtractor
	poster *fakePoster
}

func setup(t *testing.T) *env {
	t.Helper()
	return setupApproved(t, nil)
}

// setupApproved is setup with an APPROVED_FORWARDING_EMAILS allowlist, whose
// senders bypass the DigestInterval wait.
func setupApproved(t *testing.T, approved map[string]bool) *env {
	t.Helper()
	d, err := db.Open("file:" + filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Migrate(context.Background(), d); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	st := store.New(d)

	ex := &spyExtractor{deals: []extract.Deal{
		{Source: "Humble", Title: "Bundle A", URL: "https://h/a"},
		{Source: "Humble", Title: "Bundle B"},
	}}
	poster := &fakePoster{}
	cfg := ingest.Config{
		IngestToken:     token,
		DigestInterval:  24 * time.Hour,
		RepostAfter:     45 * 24 * time.Hour,
		ApprovedSenders: approved,
	}
	h := ingest.New(cfg, st, ex, poster)

	e := echo.New()
	e.HideBanner = true
	h.Register(e)
	return &env{e: e, h: h, st: st, ex: ex, poster: poster}
}

func rawEmail(messageID, subject, body string) []byte {
	return rawEmailFrom("Humble Bundle <deals@humblebundle.com>", messageID, subject, body)
}

func rawEmailFrom(from, messageID, subject, body string) []byte {
	return []byte("From: " + from + "\r\n" +
		"To: me@example.com\r\nSubject: " + subject + "\r\nMessage-Id: " + messageID + "\r\n" +
		"Date: Mon, 20 Jul 2026 10:00:00 +0000\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" + body)
}

func (e *env) post(t *testing.T, path, authToken string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	return e.postEnvelope(t, path, authToken, "", body)
}

// postEnvelope is post plus the X-Envelope-From header the Cloudflare Worker sets
// on every /api/ingest POST (empty means the header is omitted).
func (e *env) postEnvelope(t *testing.T, path, authToken, envelopeFrom string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	if envelopeFrom != "" {
		req.Header.Set("X-Envelope-From", envelopeFrom)
	}
	req.Header.Set("Content-Type", "message/rfc822")
	rec := httptest.NewRecorder()
	e.e.ServeHTTP(rec, req)
	return rec
}

func decodeResult(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	return m
}

func TestIngestHappyPath(t *testing.T) {
	e := setup(t)
	rec := e.post(t, "/api/ingest", token, rawEmail("<a@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	res := decodeResult(t, rec)
	if res["deals_extracted"].(float64) != 2 {
		t.Errorf("deals_extracted = %v, want 2", res["deals_extracted"])
	}
	if res["digest_posted"] != true {
		t.Errorf("digest_posted = %v, want true", res["digest_posted"])
	}
	if got := atomic.LoadInt32(&e.ex.calls); got != 1 {
		t.Errorf("extractor calls = %d, want 1", got)
	}
	if e.poster.count() != 1 {
		t.Errorf("poster count = %d, want 1", e.poster.count())
	}
	got, err := e.st.GetEmailByMessageID(context.Background(), "<a@x>")
	if err != nil {
		t.Fatalf("GetEmailByMessageID: %v", err)
	}
	if got.Status != store.StatusExtracted {
		t.Errorf("email status = %q, want extracted", got.Status)
	}
}

func TestIngestResolvesDealURLs(t *testing.T) {
	// The stored deal carries the resolver's cleaned URL, not the extracted one.
	e := setup(t)
	fr := &fakeResolver{byURL: map[string]string{
		"https://h/a": "https://deals.example/a",
	}}
	e.h.Resolver = fr

	rec := e.post(t, "/api/ingest", token, rawEmail("<r1@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}

	ctx := context.Background()
	a, err := e.st.GetDealByDedupeKey(ctx, store.DedupeKey("deals@humblebundle.com", "Bundle A"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey A: %v", err)
	}
	if a.URL != "https://deals.example/a" {
		t.Errorf("deal A URL = %q, want the resolved URL", a.URL)
	}
	if a.EndsAt != "" {
		t.Errorf("deal A ends_at = %q, want empty (no page returned)", a.EndsAt)
	}
	b, err := e.st.GetDealByDedupeKey(ctx, store.DedupeKey("deals@humblebundle.com", "Bundle B"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey B: %v", err)
	}
	if b.URL != "" {
		t.Errorf("deal B URL = %q, want empty (no URL extracted)", b.URL)
	}
	if fr.pageCalls != 2 {
		t.Errorf("resolver page calls = %d, want 2 (one per deal; both lack ends_at)", fr.pageCalls)
	}
}

func TestIngestFillsEndsAtFromPage(t *testing.T) {
	// A deal whose email stated no deadline gets one mined from the resolved
	// page's structured data.
	e := setup(t)
	e.h.Resolver = &fakeResolver{
		page: []byte(`<html><head><script type="application/ld+json">
			{"@type":"Product","offers":{"priceValidUntil":"2026-08-01"}}
		</script></head><body></body></html>`),
	}

	rec := e.post(t, "/api/ingest", token, rawEmail("<x1@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	ctx := context.Background()
	a, err := e.st.GetDealByDedupeKey(ctx, store.DedupeKey("deals@humblebundle.com", "Bundle A"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey A: %v", err)
	}
	if a.EndsAt != "2026-08-01" {
		t.Errorf("deal A ends_at = %q, want 2026-08-01 from the page", a.EndsAt)
	}
	// Bundle B has no URL, so there is no page to mine.
	b, err := e.st.GetDealByDedupeKey(ctx, store.DedupeKey("deals@humblebundle.com", "Bundle B"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey B: %v", err)
	}
	if b.EndsAt != "" {
		t.Errorf("deal B ends_at = %q, want empty", b.EndsAt)
	}
}

func TestIngestEmailEndsAtNotOverwritten(t *testing.T) {
	// With both the deadline and the price stated by the email, there is
	// nothing to mine: the page is not even fetched, and the stated deadline
	// wins.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Humble", Title: "Bundle C", URL: "https://h/c", Price: "$10", EndsAt: "2026-12-31"},
	}, nil)
	fr := &fakeResolver{
		page: []byte(`<html><head><script type="application/ld+json">
			{"priceValidUntil":"2026-08-01"}
		</script></head></html>`),
	}
	e.h.Resolver = fr

	rec := e.post(t, "/api/ingest", token, rawEmail("<x2@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle C"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "2026-12-31" {
		t.Errorf("ends_at = %q, want the email's 2026-12-31 kept", d.EndsAt)
	}
	if fr.pageCalls != 0 {
		t.Errorf("page calls = %d, want 0 (deadline and price already known)", fr.pageCalls)
	}
	if fr.resolveCalls != 1 {
		t.Errorf("resolve calls = %d, want 1", fr.resolveCalls)
	}
}

func TestIngestFillsPriceFromPage(t *testing.T) {
	// A deal with an email-stated deadline but no price still gets the page
	// fetched, and the page's offer price fills the gap — without touching the
	// stated deadline.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Humble", Title: "Bundle D", URL: "https://h/d", EndsAt: "2026-12-31"},
	}, nil)
	fr := &fakeResolver{
		page: []byte(`<html><head><script type="application/ld+json">
			{"offers":{"price":"25","priceCurrency":"USD","priceValidUntil":"2026-08-01"}}
		</script></head></html>`),
	}
	e.h.Resolver = fr

	rec := e.post(t, "/api/ingest", token, rawEmail("<x5@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle D"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.Price != "$25" {
		t.Errorf("price = %q, want $25 from the page", d.Price)
	}
	if d.EndsAt != "2026-12-31" {
		t.Errorf("ends_at = %q, want the email's 2026-12-31 kept (page date ignored)", d.EndsAt)
	}
	if fr.pageCalls != 1 {
		t.Errorf("page calls = %d, want 1 (missing price forces the fetch)", fr.pageCalls)
	}
}

func TestIngestEmailPriceNotOverwritten(t *testing.T) {
	// The page is fetched for the missing deadline, but the email-stated price
	// must survive the mining.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Humble", Title: "Bundle E", URL: "https://h/e", Price: "$5 (90% off)"},
	}, nil)
	fr := &fakeResolver{
		page: []byte(`<html><head><script type="application/ld+json">
			{"offers":{"price":"25","priceCurrency":"USD","priceValidUntil":"2026-08-01"}}
		</script></head></html>`),
	}
	e.h.Resolver = fr

	rec := e.post(t, "/api/ingest", token, rawEmail("<x6@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle E"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.Price != "$5 (90% off)" {
		t.Errorf("price = %q, want the email's price kept", d.Price)
	}
	if d.EndsAt != "2026-08-01" {
		t.Errorf("ends_at = %q, want 2026-08-01 mined from the page", d.EndsAt)
	}
}

// fakeEnricher returns a canned Enrichment (or error) and records what it was
// asked about.
type fakeEnricher struct {
	mu       sync.Mutex
	en       extract.Enrichment
	err      error
	calls    int
	gotTitle string
	gotText  string
}

func (f *fakeEnricher) EnrichFromPage(_ context.Context, title, pageText string, _ *time.Time) (extract.Enrichment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.gotTitle = title
	f.gotText = pageText
	return f.en, f.err
}

func TestIngestEnrichesFromPageText(t *testing.T) {
	// No deadline anywhere structured: the enricher reads the page text and
	// fills both missing fields; the text it sees is the converted page, not
	// raw HTML.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Humble", Title: "Bundle F", URL: "https://h/f"},
	}, nil)
	e.h.Resolver = &fakeResolver{page: []byte(`<html><body><p>Sale ends August 20. Now $19.</p></body></html>`)}
	fe := &fakeEnricher{en: extract.Enrichment{EndsAt: "2026-08-20", Price: "$19"}}
	e.h.Enricher = fe

	rec := e.post(t, "/api/ingest", token, rawEmail("<x7@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle F"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "2026-08-20" || d.Price != "$19" {
		t.Errorf("deal = ends %q price %q, want the enricher's 2026-08-20 / $19", d.EndsAt, d.Price)
	}
	if fe.calls != 1 {
		t.Errorf("enricher calls = %d, want 1", fe.calls)
	}
	if fe.gotTitle != "Bundle F" || strings.Contains(fe.gotText, "<p>") {
		t.Errorf("enricher saw title %q text %q, want the title and tag-free text", fe.gotTitle, fe.gotText)
	}
}

func TestIngestEnricherSkippedWhenStructuredDataHit(t *testing.T) {
	// JSON-LD already answered the deadline: no model call is spent.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Humble", Title: "Bundle G", URL: "https://h/g"},
	}, nil)
	e.h.Resolver = &fakeResolver{
		page: []byte(`<html><head><script type="application/ld+json">
			{"priceValidUntil":"2026-08-01"}
		</script></head></html>`),
	}
	fe := &fakeEnricher{en: extract.Enrichment{EndsAt: "2026-09-09"}}
	e.h.Enricher = fe

	rec := e.post(t, "/api/ingest", token, rawEmail("<x8@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle G"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "2026-08-01" {
		t.Errorf("ends_at = %q, want the structured data's 2026-08-01", d.EndsAt)
	}
	if fe.calls != 0 {
		t.Errorf("enricher calls = %d, want 0 (structured data already answered)", fe.calls)
	}
}

func TestIngestEnricherPastDateDropped(t *testing.T) {
	// The model gets the same skepticism as every other deadline source: a
	// date before the email's own send date (fixture Date: 2026-07-20) is
	// dropped, while its price still lands.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Humble", Title: "Bundle H", URL: "https://h/h"},
	}, nil)
	e.h.Resolver = &fakeResolver{page: []byte(`<html><body><p>an old promo page</p></body></html>`)}
	e.h.Enricher = &fakeEnricher{en: extract.Enrichment{EndsAt: "2025-01-01", Price: "$9"}}

	rec := e.post(t, "/api/ingest", token, rawEmail("<x9@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle H"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "" {
		t.Errorf("ends_at = %q, want empty (implausible enriched date dropped)", d.EndsAt)
	}
	if d.Price != "$9" {
		t.Errorf("price = %q, want $9 (price is not date-guarded)", d.Price)
	}
}

func TestIngestEnricherFailureKeepsValues(t *testing.T) {
	// An enricher outage must not fail the ingest or corrupt the deal.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Humble", Title: "Bundle I", URL: "https://h/i"},
	}, nil)
	e.h.Resolver = &fakeResolver{page: []byte(`<html><body><p>page</p></body></html>`)}
	e.h.Enricher = &fakeEnricher{err: errors.New("model down")}

	rec := e.post(t, "/api/ingest", token, rawEmail("<x10@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle I"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "" || d.Price != "" {
		t.Errorf("deal = ends %q price %q, want both empty after enricher failure", d.EndsAt, d.Price)
	}
}

func TestIngestEnricherPriceNotOverwritten(t *testing.T) {
	// The enricher runs for the missing deadline, but an email-stated price
	// must survive it.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Humble", Title: "Bundle J", URL: "https://h/j", Price: "$5 (90% off)"},
	}, nil)
	e.h.Resolver = &fakeResolver{page: []byte(`<html><body><p>Now $25!</p></body></html>`)}
	e.h.Enricher = &fakeEnricher{en: extract.Enrichment{EndsAt: "2026-08-20", Price: "$25"}}

	rec := e.post(t, "/api/ingest", token, rawEmail("<x11@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle J"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.Price != "$5 (90% off)" {
		t.Errorf("price = %q, want the email's price kept", d.Price)
	}
	if d.EndsAt != "2026-08-20" {
		t.Errorf("ends_at = %q, want the enricher's date", d.EndsAt)
	}
}

func TestIngestImplausibleEndsAtRefilledFromPage(t *testing.T) {
	// An extracted deadline already in the past when the email was sent (the
	// fixture's Date is 2026-07-20) carries a stale or guessed year, not a deadline: it
	// must be dropped, which routes the deal through the page fetch so the
	// page's structured data supplies the real date.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Manning", Title: "AI Agents in Action", URL: "https://h/m", EndsAt: "2025-11-27"},
	}, nil)
	fr := &fakeResolver{
		page: []byte(`<html><head><script type="application/ld+json">
			{"@type":"Product","offers":{"priceValidUntil":"2026-11-27"}}
		</script></head></html>`),
	}
	e.h.Resolver = fr

	rec := e.post(t, "/api/ingest", token, rawEmail("<x3@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "AI Agents in Action"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "2026-11-27" {
		t.Errorf("ends_at = %q, want the page's 2026-11-27 replacing the impossible 2025-11-27", d.EndsAt)
	}
	if fr.pageCalls != 1 {
		t.Errorf("page calls = %d, want 1 (implausible deadline forces the page fetch)", fr.pageCalls)
	}
	if fr.resolveCalls != 0 {
		t.Errorf("resolve calls = %d, want 0", fr.resolveCalls)
	}
}

func TestIngestImplausibleEndsAtDroppedWhenPageHasNoDate(t *testing.T) {
	// Same impossible deadline, but the page offers nothing to mine: the deal
	// must store an empty ends_at rather than keep the impossible date.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Manning", Title: "AI Agents in Action", URL: "https://h/m", EndsAt: "2025-11-27"},
	}, nil)
	e.h.Resolver = &fakeResolver{page: []byte(`<html><body>no structured data</body></html>`)}

	rec := e.post(t, "/api/ingest", token, rawEmail("<x4@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "AI Agents in Action"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "" {
		t.Errorf("ends_at = %q, want empty (impossible date dropped, page had none)", d.EndsAt)
	}
}

func TestIngestPastPageDeadlineAlsoDropped(t *testing.T) {
	// When the extracted deadline is rejected and the fetched page's own
	// structured data is also in the past (stale priceValidUntil from an
	// earlier promotion), neither may be stored.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Manning", Title: "AI Agents in Action", URL: "https://h/m", EndsAt: "2025-11-27"},
	}, nil)
	e.h.Resolver = &fakeResolver{
		page: []byte(`<html><head><script type="application/ld+json">
			{"@type":"Product","offers":{"priceValidUntil":"2025-06-01"}}
		</script></head></html>`),
	}

	rec := e.post(t, "/api/ingest", token, rawEmail("<x7@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "AI Agents in Action"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "" {
		t.Errorf("ends_at = %q, want empty (both extracted and page deadlines implausible)", d.EndsAt)
	}
}

func TestIngestImplausibleEndsAtDroppedWithoutResolver(t *testing.T) {
	// The deadline guard must run even with no resolver wired: an impossible
	// date never reaches the store.
	e := setup(t)
	e.ex.set([]extract.Deal{
		{Source: "Manning", Title: "AI Agents in Action", URL: "https://h/m", EndsAt: "2025-11-27"},
	}, nil)

	rec := e.post(t, "/api/ingest", token, rawEmail("<x5@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "AI Agents in Action"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "" {
		t.Errorf("ends_at = %q, want empty (guard runs without a resolver)", d.EndsAt)
	}
}

func TestIngestReingestClearsStoredHallucinatedDeadline(t *testing.T) {
	// A bad deadline stored by an earlier ingest (before the guard existed)
	// must be erased when the same deal is sighted again and the guard rejects
	// the same extracted date — plain COALESCE would keep the stored copy.
	e := setup(t)
	ctx := context.Background()

	seedEmail, _, err := e.st.InsertEmail(ctx, store.Email{
		MessageID: "<seed@x>", From: "Humble Bundle <deals@humblebundle.com>", Subject: "Deals", BodyText: "body",
	})
	if err != nil {
		t.Fatalf("InsertEmail: %v", err)
	}
	key := store.DedupeKey("deals@humblebundle.com", "AI Agents in Action")
	if err := e.st.UpsertDeal(ctx, store.Deal{
		EmailID: seedEmail.ID, DedupeKey: key, Source: "Manning",
		Title: "AI Agents in Action", URL: "https://h/m", EndsAt: "2025-11-27",
	}, time.Time{}); err != nil {
		t.Fatalf("UpsertDeal seed: %v", err)
	}

	e.ex.set([]extract.Deal{
		{Source: "Manning", Title: "AI Agents in Action", URL: "https://h/m", EndsAt: "2025-11-27"},
	}, nil)
	e.h.Resolver = &fakeResolver{page: []byte(`<html><body>no structured data</body></html>`)}

	rec := e.post(t, "/api/ingest", token, rawEmail("<x6@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	d, err := e.st.GetDealByDedupeKey(ctx, key)
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if d.EndsAt != "" {
		t.Errorf("ends_at = %q, want the stored impossible date cleared on re-sighting", d.EndsAt)
	}
	if d.SeenCount != 2 {
		t.Errorf("seen_count = %d, want 2", d.SeenCount)
	}
}

func TestIngestResolverErrorKeepsURLAndSucceeds(t *testing.T) {
	// A failing resolver must not fail the ingest or lose the extracted URL.
	e := setup(t)
	e.h.Resolver = &fakeResolver{err: errors.New("resolver down")}

	rec := e.post(t, "/api/ingest", token, rawEmail("<r2@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	a, err := e.st.GetDealByDedupeKey(context.Background(), store.DedupeKey("deals@humblebundle.com", "Bundle A"))
	if err != nil {
		t.Fatalf("GetDealByDedupeKey: %v", err)
	}
	if a.URL != "https://h/a" {
		t.Errorf("deal A URL = %q, want the extracted URL kept", a.URL)
	}
}

func TestIngestDuplicateSkipsExtraction(t *testing.T) {
	e := setup(t)
	if rec := e.post(t, "/api/ingest", token, rawEmail("<dup@x>", "Deals", "body")); rec.Code != 200 {
		t.Fatalf("first: %d", rec.Code)
	}
	rec := e.post(t, "/api/ingest", token, rawEmail("<dup@x>", "Deals", "body"))
	if rec.Code != http.StatusOK {
		t.Fatalf("second status = %d", rec.Code)
	}
	if res := decodeResult(t, rec); res["duplicate"] != true {
		t.Errorf("duplicate = %v, want true", res["duplicate"])
	}
	if got := atomic.LoadInt32(&e.ex.calls); got != 1 {
		t.Errorf("extractor called %d times, want 1 (no re-extract)", got)
	}
}

func TestIngestAuth(t *testing.T) {
	e := setup(t)
	if rec := e.post(t, "/api/ingest", "", rawEmail("<n@x>", "s", "b")); rec.Code != http.StatusUnauthorized {
		t.Errorf("no token: status = %d, want 401", rec.Code)
	}
	if rec := e.post(t, "/api/ingest", "wrong", rawEmail("<n@x>", "s", "b")); rec.Code != http.StatusUnauthorized {
		t.Errorf("bad token: status = %d, want 401", rec.Code)
	}
	if got := atomic.LoadInt32(&e.ex.calls); got != 0 {
		t.Errorf("extractor should not run for unauthorized requests, got %d calls", got)
	}
}

func TestIngestOversize(t *testing.T) {
	e := setup(t)
	e.h.MaxIngestBytes = 64
	rec := e.post(t, "/api/ingest", token, rawEmail("<big@x>", "s", strings.Repeat("x", 500)))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", rec.Code)
	}
	if got := atomic.LoadInt32(&e.ex.calls); got != 0 {
		t.Errorf("extractor should not run for oversize body, got %d calls", got)
	}
}

func TestIngestMalformedEmail(t *testing.T) {
	e := setup(t)
	// A line without a colon is an invalid header; mail.ReadMessage rejects it.
	rec := e.post(t, "/api/ingest", token, []byte("this is not a valid email"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if got := atomic.LoadInt32(&e.ex.calls); got != 0 {
		t.Errorf("extractor should not run for malformed email, got %d calls", got)
	}
}

func TestIngestDigestFailureDoesNotFailIngest(t *testing.T) {
	e := setup(t)
	e.poster.err = errors.New("slack down") // set before the synchronous request; no concurrency

	rec := e.post(t, "/api/ingest", token, rawEmail("<d@x>", "s", "b"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingest must succeed even if the digest post fails", rec.Code)
	}
	if res := decodeResult(t, rec); res["digest_posted"] != false {
		t.Errorf("digest_posted = %v, want false", res["digest_posted"])
	}
	got, _ := e.st.GetEmailByMessageID(context.Background(), "<d@x>")
	if got.Status != store.StatusExtracted {
		t.Errorf("email status = %q, want extracted", got.Status)
	}
	// Deals stay queued for a later attempt.
	if left, _ := e.st.UnpostedDeals(context.Background(), 0); len(left) != 2 {
		t.Errorf("deals should stay queued, got %d unposted", len(left))
	}
}

func TestIngestExtractorErrorThenRetry(t *testing.T) {
	e := setup(t)
	e.ex.set(nil, errors.New("gemini down"))

	rec := e.post(t, "/api/ingest", token, rawEmail("<r@x>", "s", "b"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	got, _ := e.st.GetEmailByMessageID(context.Background(), "<r@x>")
	if got.Status != store.StatusExtractFailed || got.ExtractError == "" {
		t.Errorf("email status = %q err = %q, want extract_failed", got.Status, got.ExtractError)
	}

	// Re-POST after the extractor recovers: the failed email retries extraction.
	e.ex.set([]extract.Deal{{Source: "H", Title: "Recovered"}}, nil)
	rec = e.post(t, "/api/ingest", token, rawEmail("<r@x>", "s", "b"))
	if rec.Code != http.StatusOK {
		t.Fatalf("retry status = %d", rec.Code)
	}
	got, _ = e.st.GetEmailByMessageID(context.Background(), "<r@x>")
	if got.Status != store.StatusExtracted {
		t.Errorf("after retry status = %q, want extracted", got.Status)
	}
	if calls := atomic.LoadInt32(&e.ex.calls); calls != 2 {
		t.Errorf("extractor calls = %d, want 2 (initial + retry)", calls)
	}
}

func TestIngestDigestSuppressedWithinInterval(t *testing.T) {
	e := setup(t)
	// First email posts a digest.
	if rec := e.post(t, "/api/ingest", token, rawEmail("<a@x>", "s", "b")); rec.Code != 200 {
		t.Fatalf("first: %d", rec.Code)
	}
	if e.poster.count() != 1 {
		t.Fatalf("poster count = %d, want 1", e.poster.count())
	}
	// A fresh deal arrives within the interval; ingest must not post again.
	e.ex.set([]extract.Deal{{Source: "H", Title: "Latecomer", URL: "https://h/late"}}, nil)
	rec := e.post(t, "/api/ingest", token, rawEmail("<b@x>", "s", "b"))
	if rec.Code != http.StatusOK {
		t.Fatalf("second: %d", rec.Code)
	}
	if res := decodeResult(t, rec); res["digest_posted"] != false {
		t.Errorf("digest_posted = %v, want false (suppressed)", res["digest_posted"])
	}
	if e.poster.count() != 1 {
		t.Errorf("poster count = %d, want 1 (no double post)", e.poster.count())
	}
}

// approvedSenders is the allowlist used by the bypass tests below.
func approvedSenders() map[string]bool { return map[string]bool{"alice@example.com": true} }

// primeInterval ingests one email (which posts a digest and so opens the
// once-per-interval suppression window) and queues a fresh deal behind it.
func primeInterval(t *testing.T, e *env) {
	t.Helper()
	if rec := e.post(t, "/api/ingest", token, rawEmail("<a@x>", "s", "b")); rec.Code != http.StatusOK {
		t.Fatalf("priming ingest: %d", rec.Code)
	}
	if e.poster.count() != 1 {
		t.Fatalf("poster count = %d, want 1 after priming", e.poster.count())
	}
	e.ex.set([]extract.Deal{{Source: "H", Title: "Latecomer", URL: "https://h/late"}}, nil)
}

func TestIngestApprovedSenderForcesDigest(t *testing.T) {
	e := setupApproved(t, approvedSenders())
	primeInterval(t, e)

	// Same window as TestIngestDigestSuppressedWithinInterval, but the From:
	// header is on the allowlist, so the queued deal posts now.
	rec := e.post(t, "/api/ingest", token, rawEmailFrom("Alice <ALICE@Example.com>", "<b@x>", "s", "b"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	res := decodeResult(t, rec)
	if res["digest_posted"] != true {
		t.Errorf("digest_posted = %v, want true (forced)", res["digest_posted"])
	}
	if res["forced"] != true {
		t.Errorf("forced = %v, want true", res["forced"])
	}
	if e.poster.count() != 2 {
		t.Errorf("poster count = %d, want 2", e.poster.count())
	}
}

func TestIngestApprovedEnvelopeFromForcesDigest(t *testing.T) {
	e := setupApproved(t, approvedSenders())
	primeInterval(t, e)

	// An auto-forward rule keeps the newsletter's From: and only the envelope
	// sender identifies the forwarder, so X-Envelope-From must be honored too.
	rec := e.postEnvelope(t, "/api/ingest", token, "alice@example.com", rawEmail("<b@x>", "s", "b"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	res := decodeResult(t, rec)
	if res["digest_posted"] != true {
		t.Errorf("digest_posted = %v, want true (forced)", res["digest_posted"])
	}
	if res["forced"] != true {
		t.Errorf("forced = %v, want true", res["forced"])
	}
	if e.poster.count() != 2 {
		t.Errorf("poster count = %d, want 2", e.poster.count())
	}
}

func TestIngestUnapprovedSenderStillSuppressed(t *testing.T) {
	e := setupApproved(t, approvedSenders())
	primeInterval(t, e)

	rec := e.postEnvelope(t, "/api/ingest", token, "mallory@example.com",
		rawEmailFrom("Mallory <mallory@example.com>", "<b@x>", "s", "b"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	res := decodeResult(t, rec)
	if res["digest_posted"] != false {
		t.Errorf("digest_posted = %v, want false (suppressed)", res["digest_posted"])
	}
	if res["forced"] == true {
		t.Errorf("forced = %v, want absent", res["forced"])
	}
	if e.poster.count() != 1 {
		t.Errorf("poster count = %d, want 1 (no double post)", e.poster.count())
	}
}

func TestAdminDigestForces(t *testing.T) {
	e := setup(t)
	if rec := e.post(t, "/api/ingest", token, rawEmail("<a@x>", "s", "b")); rec.Code != 200 {
		t.Fatalf("ingest: %d", rec.Code)
	}
	// New deal queued within the interval (ingest won't post it).
	e.ex.set([]extract.Deal{{Source: "H", Title: "Forced Deal", URL: "https://h/f"}}, nil)
	if rec := e.post(t, "/api/ingest", token, rawEmail("<b@x>", "s", "b")); rec.Code != 200 {
		t.Fatalf("ingest b: %d", rec.Code)
	}
	if e.poster.count() != 1 {
		t.Fatalf("poster count = %d, want 1 before force", e.poster.count())
	}

	rec := e.post(t, "/api/admin/digest", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin/digest status = %d", rec.Code)
	}
	if res := decodeResult(t, rec); res["posted"] != true {
		t.Errorf("posted = %v, want true", res["posted"])
	}
	if e.poster.count() != 2 {
		t.Errorf("poster count = %d, want 2 after force", e.poster.count())
	}
}

func TestAdminDigestAuth(t *testing.T) {
	e := setup(t)
	if rec := e.post(t, "/api/admin/digest", "", nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
