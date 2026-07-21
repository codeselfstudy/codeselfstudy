package mailparse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return raw
}

func TestParseFixtures(t *testing.T) {
	tests := []struct {
		file            string
		messageID       string // exact; "" means expect a "sha256:" fallback
		subject         string
		fromContains    string
		wantContains    []string
		wantNotContains []string
	}{
		{
			file:         "multipart_alternative.eml",
			messageID:    "<week1@humblebundle.com>",
			subject:      "This Week's Deals", // RFC 2047 encoded-word decoded
			fromContains: "deals@humblebundle.com",
			wantContains: []string{"Humble Programming Bundle", "Pay what you want", "https://humblebundle.com/books/programming"},
		},
		{
			file:            "html_only.eml",
			messageID:       "<html-only@indiedeals.io>",
			subject:         "HTML Only Newsletter",
			fromContains:    "news@indiedeals.io",
			wantContains:    []string{"GameDev Asset Sale", "50% off", "Browse assets (https://indiedeals.io/gamedev)"},
			wantNotContains: []string{"color:red", "trackOpen", "ignore me"},
		},
		{
			file:         "quoted_printable.eml",
			messageID:    "<qp@bookstore.example>",
			subject:      "QP Encoded",
			fromContains: "deals@bookstore.example",
			wantContains: []string{"Big Sale — 50% off", "https://bookstore.example/sale", "the full list"},
		},
		{
			file:         "base64_html.eml",
			messageID:    "<base64@courses.example>",
			subject:      "Base64 HTML",
			fromContains: "deals@courses.example",
			wantContains: []string{"Course Bootcamp Sale", "Half price bootcamp", "Enroll now (https://courses.example/deal)"},
		},
		{
			file:         "latin1.eml",
			messageID:    "<latin1@cafe.example>",
			subject:      "Latin1 Charset",
			fromContains: "deals@cafe.example",
			wantContains: []string{"Café résumé workshop", "30% off", "https://cafe.example/deal"},
		},
		{
			file:         "no_message_id.eml",
			messageID:    "", // sha256 fallback
			subject:      "No Message ID Header",
			fromContains: "hello@nomessageid.example",
			wantContains: []string{"sha256 fallback"},
		},
		{
			file:            "stub_plain.eml",
			messageID:       "<stub@courses.example>",
			subject:         "Stub Plain Part",
			fromContains:    "deals@courses.example",
			wantContains:    []string{"Python Mega Course Sale", "Lifetime access", "Enroll now (https://courses.example/python)"},
			wantNotContains: []string{"View this email in your browser"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			e, err := Parse(loadFixture(t, tt.file))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if tt.messageID == "" {
				if !strings.HasPrefix(e.MessageID, "sha256:") {
					t.Errorf("MessageID = %q, want sha256: fallback", e.MessageID)
				}
			} else if e.MessageID != tt.messageID {
				t.Errorf("MessageID = %q, want %q", e.MessageID, tt.messageID)
			}
			if e.Subject != tt.subject {
				t.Errorf("Subject = %q, want %q", e.Subject, tt.subject)
			}
			if !strings.Contains(e.From, tt.fromContains) {
				t.Errorf("From = %q, want to contain %q", e.From, tt.fromContains)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(e.Text, want) {
					t.Errorf("Text missing %q\n---\n%s\n---", want, e.Text)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(e.Text, notWant) {
					t.Errorf("Text should not contain %q\n---\n%s\n---", notWant, e.Text)
				}
			}
			if !utf8.ValidString(e.Text) {
				t.Errorf("Text is not valid UTF-8")
			}
		})
	}
}

func TestParseSentAt(t *testing.T) {
	e, err := Parse(loadFixture(t, "multipart_alternative.eml"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if e.SentAt == nil {
		t.Fatal("SentAt is nil, want parsed Date")
	}
	want := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)
	if !e.SentAt.Equal(want) {
		t.Errorf("SentAt = %v, want %v", e.SentAt, want)
	}
}

func TestParseMislabeledUTF8(t *testing.T) {
	// Declares utf-8 but the body byte 0xE9 ("é" in Latin-1) is invalid UTF-8.
	raw := "From: x@example.com\nSubject: Mislabeled\nMessage-Id: <mis@x>\n" +
		"Date: Mon, 20 Jul 2026 10:00:00 +0000\n" +
		"Content-Type: text/plain; charset=utf-8\n\n" +
		"Caf\xe9 deal 30% off"

	e, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !utf8.ValidString(e.Text) {
		t.Errorf("Text is not valid UTF-8: %q", e.Text)
	}
	if !strings.Contains(e.Text, "deal 30% off") {
		t.Errorf("Text = %q, want to contain the ASCII remainder", e.Text)
	}
}

func TestParseNonLatin1EncodedWordSubject(t *testing.T) {
	// windows-1252 encoded-word; 0xE9 = "é". Requires the CharsetReader wiring.
	raw := "From: x@example.com\nSubject: =?windows-1252?Q?Caf=E9_Sale?=\n" +
		"Message-Id: <w1252@x>\nDate: Mon, 20 Jul 2026 10:00:00 +0000\n" +
		"Content-Type: text/plain; charset=utf-8\n\nbody"

	e, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if e.Subject != "Café Sale" {
		t.Errorf("Subject = %q, want %q", e.Subject, "Café Sale")
	}
}

func TestParseTruncatesLongBody(t *testing.T) {
	body := strings.Repeat("a", 70000)
	raw := "From: x@example.com\nSubject: Big\nMessage-Id: <big@x>\n" +
		"Date: Mon, 20 Jul 2026 10:00:00 +0000\n" +
		"Content-Type: text/plain; charset=utf-8\n\n" + body

	e, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(e.Text) >= len(body) {
		t.Errorf("Text not truncated: len=%d", len(e.Text))
	}
	if !strings.HasSuffix(e.Text, truncSentinel) {
		t.Errorf("Text missing truncation sentinel")
	}
	if len(e.Text) > maxBodyBytes+len(truncSentinel) {
		t.Errorf("Text length %d exceeds cap", len(e.Text))
	}
}

func TestParseTruncatesOnRuneBoundary(t *testing.T) {
	// Multibyte runes so the byte cap is likely to land mid-rune.
	body := strings.Repeat("é", 40000) // 80000 bytes
	raw := "From: x@example.com\nSubject: Runes\nMessage-Id: <r@x>\n" +
		"Date: Mon, 20 Jul 2026 10:00:00 +0000\n" +
		"Content-Type: text/plain; charset=utf-8\n\n" + body

	e, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !utf8.ValidString(e.Text) {
		t.Error("truncated text is not valid UTF-8 (cut mid-rune)")
	}
	if !strings.HasSuffix(e.Text, truncSentinel) {
		t.Error("Text missing truncation sentinel")
	}
}

// nestedMultipart builds an email whose body is `levels` nested multipart/mixed
// wrappers around a single text/plain leaf ("deepest"); the leaf sits at walk
// depth == levels.
func nestedMultipart(levels int) []byte {
	part := "Content-Type: text/plain\r\n\r\ndeepest\r\n"
	for i := levels - 1; i >= 0; i-- {
		b := fmt.Sprintf("b%d", i)
		part = "Content-Type: multipart/mixed; boundary=\"" + b + "\"\r\n\r\n" +
			"--" + b + "\r\n" + part + "\r\n--" + b + "--\r\n"
	}
	return []byte("From: x@example.com\r\nSubject: Nested\r\nMessage-Id: <nest@x>\r\n" +
		"Date: Mon, 20 Jul 2026 10:00:00 +0000\r\n" + part)
}

func TestParseMIMEDepthCap(t *testing.T) {
	// Within the cap: a deep-but-bounded message still reaches its leaf.
	within, err := Parse(nestedMultipart(maxMIMEDepth / 2))
	if err != nil {
		t.Fatalf("within-cap Parse: %v", err)
	}
	if !strings.Contains(within.Text, "deepest") {
		t.Errorf("within-cap: leaf not reached, Text = %q", within.Text)
	}

	// Beyond the cap: recursion stops before the leaf. It must not panic or
	// error, and the too-deep leaf is not collected.
	beyond, err := Parse(nestedMultipart(maxMIMEDepth * 2))
	if err != nil {
		t.Fatalf("beyond-cap Parse: %v", err)
	}
	if strings.Contains(beyond.Text, "deepest") {
		t.Errorf("beyond-cap: leaf past the depth cap should not be reached, Text = %q", beyond.Text)
	}
}
