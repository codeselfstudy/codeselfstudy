package digest

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/store"
)

var update = flag.Bool("update", false, "update golden files")

// testNow is a fixed posting time so deadline suppression is deterministic; it
// sits before every EndsAt used in the fixtures below.
var testNow = time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)

func TestBuildBlocksGolden(t *testing.T) {
	deals := []store.Deal{
		{
			Title: "Humble Programming Bundle", URL: "https://humblebundle.com/books/prog",
			Price: "$25 (96% off)", EndsAt: "2026-07-27", Source: "Humble Bundle",
			Description: "15 ebooks, pay what you want.",
		},
		{Title: "GameDev Assets", Source: "Indie Deals"}, // only title + source
		{
			Title: "Books & <Widgets>", URL: "https://x.example/deal",
			Description: "Save now", // exercises mrkdwn escaping
		},
	}

	got, err := BuildBlocks(deals, testNow)
	if err != nil {
		t.Fatalf("BuildBlocks: %v", err)
	}
	if !json.Valid(got) {
		t.Fatal("output is not valid JSON")
	}

	golden := filepath.Join("testdata", "golden_blocks.json")
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with -update to create): %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("blocks mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestBuildBlocksNoTimestampFooter(t *testing.T) {
	// Slack stamps every message itself, so the payload must not carry its own
	// timestamp footer (removed with the Pacific-time machinery it needed).
	// Without overflow, no context block of any kind belongs in the payload —
	// asserting on block types rather than footer text catches a reworded
	// footer too.
	got, err := BuildBlocks([]store.Deal{{Title: "X"}}, testNow)
	if err != nil {
		t.Fatalf("BuildBlocks: %v", err)
	}
	var p struct {
		Blocks []struct {
			Type string `json:"type"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal(got, &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, b := range p.Blocks {
		if b.Type == "context" {
			t.Errorf("payload still contains a context footer:\n%s", got)
		}
	}
}

func TestBuildBlocksOverflow(t *testing.T) {
	deals := make([]store.Deal, 30)
	for i := range deals {
		deals[i] = store.Deal{Title: fmt.Sprintf("Deal %d", i)}
	}
	got, err := BuildBlocks(deals, testNow)
	if err != nil {
		t.Fatalf("BuildBlocks: %v", err)
	}

	var p struct {
		Blocks []map[string]any `json:"blocks"`
	}
	if err := json.Unmarshal(got, &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	sections := 0
	for _, b := range p.Blocks {
		if b["type"] == "section" {
			sections++
		}
	}
	if sections != MaxDealsPerDigest {
		t.Errorf("section blocks = %d, want %d", sections, MaxDealsPerDigest)
	}
	if !strings.Contains(string(got), "+ 5 more") {
		t.Errorf("missing overflow note for 5 extra deals")
	}
	if !strings.Contains(string(got), "30 new deals") {
		t.Errorf("header should count all 30 new deals")
	}
}

func TestHeaderTextSingular(t *testing.T) {
	if got := headerText(1); got != "1 new deal" {
		t.Errorf("headerText(1) = %q", got)
	}
	if got := headerText(2); got != "2 new deals" {
		t.Errorf("headerText(2) = %q", got)
	}
}

func TestEscapeMrkdwn(t *testing.T) {
	if got := escapeMrkdwn("A & B <c> >d"); got != "A &amp; B &lt;c&gt; &gt;d" {
		t.Errorf("escapeMrkdwn = %q", got)
	}
}

func TestDealTextOmitsEmptyFields(t *testing.T) {
	got := dealText(store.Deal{Title: "Bare"}, testNow)
	if got != "*Bare*" {
		t.Errorf("dealText(bare) = %q, want *Bare*", got)
	}
}

func TestDealTextDeadlineGuard(t *testing.T) {
	// A stored deadline already in the past (usually a hallucinated year) must
	// not render; a current, future, or free-form one must.
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name, endsAt string
		want         bool // "ends …" rendered?
	}{
		{"past hidden", "2025-11-27", false},
		{"today shown", "2026-07-26", true},
		{"future shown", "2026-11-27", true},
		{"free-form shown", "while supplies last", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dealText(store.Deal{Title: "T", EndsAt: tc.endsAt}, now)
			if rendered := strings.Contains(got, "ends "+tc.endsAt); rendered != tc.want {
				t.Errorf("dealText with ends_at %q rendered=%v, want %v (text %q)", tc.endsAt, rendered, tc.want, got)
			}
		})
	}
}

func TestDealTextEscapesURLDelimiters(t *testing.T) {
	// Any |, <, > that survives param-stripping (e.g. in the path) must not break
	// Slack's <url|text> syntax.
	got := dealText(store.Deal{Title: "T", URL: "https://x.example/d|e<f>"}, testNow)
	if strings.Contains(got, "|e") || strings.Contains(got, "<f>") {
		t.Errorf("URL delimiters not escaped: %q", got)
	}
	want := "*<https://x.example/d%7Ce%3Cf%3E|T>*"
	if got != want {
		t.Errorf("dealText = %q, want %q", got, want)
	}
}

func TestDealTextStripsQueryParams(t *testing.T) {
	// The tracking query attached by newsletters must be gone from the Slack link.
	got := dealText(store.Deal{Title: "T", URL: "https://x.example/deal?utm_source=news&id=9"}, testNow)
	want := "*<https://x.example/deal|T>*"
	if got != want {
		t.Errorf("dealText = %q, want %q", got, want)
	}
}

func TestStripQueryParams(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{
			"humble bundle newsletter",
			"https://www.humblebundle.com/books/red-team-blue-team-hackers-complete-playbook-packt-books?mcID=102:6a5f9e600af1ea383606e189:ot:6a6064f6d273259125eb7a94:1&linkID={$linkID}&utm_source=Humble+Bundle+Newsletter&utm_content=cta_button&utm_medium=email&utm_campaign=teenagemutantninjaturtlesreturntonewyorkidw_bookbundle",
			"https://www.humblebundle.com/books/red-team-blue-team-hackers-complete-playbook-packt-books",
		},
		{"no query unchanged", "https://x.example/deal", "https://x.example/deal"},
		{"trailing question mark", "https://x.example/deal?", "https://x.example/deal"},
		{"fragment kept when no query", "https://x.example/deal#section", "https://x.example/deal#section"},
		{"query dropped, fragment kept", "https://x.example/deal?utm=1#anchor", "https://x.example/deal#anchor"},
		{"question mark inside fragment is not a query", "https://x.example/deal#section?foo", "https://x.example/deal#section?foo"},
		{"hash route with params kept", "https://x.example/#/deals?id=9", "https://x.example/#/deals?id=9"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripQueryParams(tc.in); got != tc.want {
				t.Errorf("stripQueryParams(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
