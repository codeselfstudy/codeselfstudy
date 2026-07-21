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

func TestBuildBlocksGolden(t *testing.T) {
	now := time.Date(2026, 7, 20, 14, 2, 0, 0, time.UTC)
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

	got, err := BuildBlocks(deals, now)
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

func TestBuildBlocksOverflow(t *testing.T) {
	deals := make([]store.Deal, 30)
	for i := range deals {
		deals[i] = store.Deal{Title: fmt.Sprintf("Deal %d", i)}
	}
	got, err := BuildBlocks(deals, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
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
	got := dealText(store.Deal{Title: "Bare"})
	if got != "*Bare*" {
		t.Errorf("dealText(bare) = %q, want *Bare*", got)
	}
}

func TestDealTextEscapesURLDelimiters(t *testing.T) {
	// A tracking URL with |, <, > must not break Slack's <url|text> syntax.
	got := dealText(store.Deal{Title: "T", URL: "https://x.example/d?a=1|2&b=<3>"})
	if strings.Contains(got, "1|2") || strings.Contains(got, "<3>") {
		t.Errorf("URL delimiters not escaped: %q", got)
	}
	want := "*<https://x.example/d?a=1%7C2&b=%3C3%3E|T>*"
	if got != want {
		t.Errorf("dealText = %q, want %q", got, want)
	}
}
