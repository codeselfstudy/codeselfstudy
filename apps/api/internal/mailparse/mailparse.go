// Package mailparse turns raw RFC 822 MIME (as delivered by the Cloudflare Email
// Worker) into a normalized Email ready for LLM extraction. It is framework-free
// and uses only the standard library plus golang.org/x/text for charset decoding.
package mailparse

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/htmlindex"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/htmltext"
)

const (
	// maxBodyBytes caps the stored/returned normalized text.
	maxBodyBytes = 64 * 1024
	// stubThreshold: a text/plain part shorter than this, when an HTML part also
	// exists, is treated as a "view in browser" stub and the HTML is preferred.
	stubThreshold = 200
	truncSentinel = "\n[truncated]"
)

// Email is the normalized result of parsing a raw message.
type Email struct {
	MessageID string
	From      string
	To        string
	Subject   string
	SentAt    *time.Time
	Text      string
}

// Parse parses raw MIME bytes. Headers are decoded (RFC 2047 encoded-words are
// resolved); the body is reduced to a single normalized text part, preferring
// text/plain but falling back to text/html (converted to text). A missing
// Message-ID falls back to "sha256:<hex of raw>". Attachments are ignored.
func Parse(raw []byte) (Email, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return Email{}, fmt.Errorf("read message: %w", err)
	}

	dec := &mime.WordDecoder{
		// Decode encoded-words in any charset x/text knows, matching the body path
		// (the default decoder only handles utf-8/iso-8859-1/us-ascii).
		CharsetReader: func(charset string, input io.Reader) (io.Reader, error) {
			enc, err := htmlindex.Get(charset)
			if err != nil {
				return nil, err
			}
			return enc.NewDecoder().Reader(input), nil
		},
	}
	decodeHeader := func(key string) string {
		v, err := dec.DecodeHeader(msg.Header.Get(key))
		if err != nil {
			return msg.Header.Get(key)
		}
		return v
	}

	e := Email{
		MessageID: strings.TrimSpace(msg.Header.Get("Message-Id")),
		From:      decodeHeader("From"),
		To:        decodeHeader("To"),
		Subject:   decodeHeader("Subject"),
	}
	if e.MessageID == "" {
		sum := sha256.Sum256(raw)
		e.MessageID = "sha256:" + hex.EncodeToString(sum[:])
	}
	if d, err := msg.Header.Date(); err == nil {
		du := d.UTC()
		e.SentAt = &du
	}

	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}
	var p parts
	if err := p.walk(msg.Body, contentType,
		msg.Header.Get("Content-Transfer-Encoding"),
		msg.Header.Get("Content-Disposition")); err != nil {
		return Email{}, fmt.Errorf("walk body: %w", err)
	}
	// Guarantee valid UTF-8 even when a body mislabels its charset (e.g. declares
	// utf-8 but carries Latin-1) — a common sender misconfiguration.
	e.Text = truncate(strings.ToValidUTF8(p.choose(), ""))
	return e, nil
}

// parts accumulates the first text/plain and text/html leaves encountered.
type parts struct {
	plain string
	html  string
}

func (p *parts) walk(r io.Reader, contentType, cte, disposition string) error {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType, params = "text/plain", map[string]string{}
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return nil
		}
		mr := multipart.NewReader(r, boundary)
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			perr := p.walk(part,
				part.Header.Get("Content-Type"),
				part.Header.Get("Content-Transfer-Encoding"),
				part.Header.Get("Content-Disposition"))
			part.Close()
			if perr != nil {
				return perr
			}
		}
	}

	// Leaf part. Skip attachments and any non-text content (images, PDFs) without
	// reading their bodies into memory.
	if d, _, _ := mime.ParseMediaType(disposition); strings.EqualFold(d, "attachment") {
		return nil
	}
	if !strings.HasPrefix(mediaType, "text/") {
		return nil
	}

	body, err := decodeBody(r, cte)
	if err != nil {
		return err
	}
	text := transcode(body, params["charset"])
	switch {
	case strings.HasPrefix(mediaType, "text/plain") && p.plain == "":
		p.plain = text
	case strings.HasPrefix(mediaType, "text/html") && p.html == "":
		p.html = text
	}
	return nil
}

// choose returns the best text: the plain part, unless it is a short stub and an
// HTML part exists, in which case the HTML is converted to text.
func (p *parts) choose() string {
	plain := strings.TrimSpace(p.plain)
	if plain != "" && !(len(plain) < stubThreshold && p.html != "") {
		return plain
	}
	if p.html != "" {
		if txt, err := htmltext.Convert(p.html); err == nil && strings.TrimSpace(txt) != "" {
			return txt
		}
	}
	return plain
}

func decodeBody(r io.Reader, cte string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(cte)) {
	case "base64":
		b, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, r))
		return string(b), err
	case "quoted-printable":
		b, err := io.ReadAll(quotedprintable.NewReader(r))
		return string(b), err
	default:
		b, err := io.ReadAll(r)
		return string(b), err
	}
}

// transcode converts a body to UTF-8 based on its declared charset. Unknown or
// already-UTF-8 charsets pass through unchanged.
func transcode(body, charset string) string {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "", "utf-8", "utf8", "us-ascii", "ascii":
		return body
	}
	enc, err := htmlindex.Get(charset)
	if err != nil {
		return body
	}
	out, err := enc.NewDecoder().String(body)
	if err != nil {
		return body
	}
	return out
}

// truncate caps s at maxBodyBytes on a UTF-8 rune boundary, appending a sentinel.
func truncate(s string) string {
	if len(s) <= maxBodyBytes {
		return s
	}
	cut := maxBodyBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + truncSentinel
}
