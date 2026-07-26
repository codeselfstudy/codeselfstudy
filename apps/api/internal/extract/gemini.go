package extract

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/codeselfstudy/codeselfstudy/apps/api/internal/mailparse"
)

// maxPromptChars caps the email body sent to Gemini. Deals live at the top of
// newsletters; the rest is unsubscribe/footer sludge that only costs tokens.
const maxPromptChars = 16000

const systemInstruction = `You extract software deals from newsletter emails for a software developer.
A deal is a discounted or time-limited offer whose subject is software or software development: developer tools and software/app licenses, programming and technical books or ebooks, and online courses on software or programming.
Include an offer only when it is about software or software development. Exclude everything else, even when discounted: video games and game bundles, comics, art or asset bundles, tabletop/RPG material, and books or courses on non-technical topics.
For a bundle that mixes software titles with unrelated ones, include it only if it is primarily about software.
Return one entry per bundle or offer — never one entry per item inside a bundle.
Set url to the link printed next to that offer if one is present.
The prompt's Date line is the day the email was sent. When a deadline omits the year, resolve it against that date so the deadline is on or after it: "ends November 27" in an email dated 2026-07-26 means 2026-11-27.
Copy prices, promo codes, and deadlines only from the email's own text; never invent, infer, or embellish them. Leave a field null when the email does not state it.
If the email has no software deals (receipts, shipping notices, account or security alerts, plain content newsletters, or only non-software offers), return an empty array.
Respond only with the JSON array described by the schema.`

// defaultRetryDelays is the backoff schedule between attempts (jitter is added).
// With the default MaxAttempts=3 only the first two are used; the third is
// headroom for a higher MaxAttempts (backoff clamps to the last entry).
var defaultRetryDelays = []time.Duration{2 * time.Second, 8 * time.Second, 20 * time.Second}

// Gemini is an Extractor backed by the Gemini API.
type Gemini struct {
	client    *genai.Client
	model     string
	genConfig *genai.GenerateContentConfig

	// MaxAttempts is the total number of attempts (including the first) on
	// retriable errors. RetryDelays is the base backoff before each retry.
	MaxAttempts int
	RetryDelays []time.Duration
}

// NewGemini builds a Gemini extractor. baseURL overrides the API endpoint (used
// by tests to point at an httptest server); pass "" for the real API.
func NewGemini(ctx context.Context, apiKey, model, baseURL string) (*Gemini, error) {
	cfg := &genai.ClientConfig{APIKey: apiKey, Backend: genai.BackendGeminiAPI}
	if baseURL != "" {
		cfg.HTTPOptions = genai.HTTPOptions{BaseURL: baseURL}
	}
	client, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new gemini client: %w", err)
	}
	return &Gemini{
		client:      client,
		model:       model,
		genConfig:   dealGenConfig(),
		MaxAttempts: 3,
		RetryDelays: defaultRetryDelays,
	}, nil
}

// dealGenConfig forces a JSON response conforming to the deal schema.
func dealGenConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction, genai.RoleUser),
		ResponseMIMEType:  "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"source":      {Type: genai.TypeString, Description: "vendor or newsletter name"},
					"title":       {Type: genai.TypeString},
					"url":         {Type: genai.TypeString, Nullable: genai.Ptr(true)},
					"price":       {Type: genai.TypeString, Nullable: genai.Ptr(true), Description: `free-form, e.g. "$25 (96% off)", "pay what you want"`},
					"ends_at":     {Type: genai.TypeString, Nullable: genai.Ptr(true), Description: "ISO 8601 date if stated; a missing year resolves against the Date line, never before it"},
					"description": {Type: genai.TypeString, Nullable: genai.Ptr(true), Description: "one short sentence"},
				},
				Required: []string{"source", "title"},
			},
		},
	}
}

// Extract implements Extractor. It retries retriable errors (HTTP 429 / 5xx and
// network errors) with backoff; a malformed response is not retried.
func (g *Gemini) Extract(ctx context.Context, e mailparse.Email) ([]Deal, error) {
	prompt := buildUserPrompt(e)

	var lastErr error
	for attempt := 0; attempt < g.MaxAttempts; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, g.backoff(attempt)); err != nil {
				return nil, err
			}
		}
		resp, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), g.genConfig)
		if err != nil {
			if isRetriable(err) {
				lastErr = err
				continue
			}
			return nil, fmt.Errorf("gemini generate: %w", err)
		}
		return parseDeals(resp.Text())
	}
	return nil, fmt.Errorf("gemini extract failed after %d attempts: %w", g.MaxAttempts, lastErr)
}

func buildUserPrompt(e mailparse.Email) string {
	body := e.Text
	if r := []rune(body); len(r) > maxPromptChars {
		body = string(r[:maxPromptChars])
	}
	// The Date line anchors yearless deadlines (see systemInstruction). Kept in
	// the sender's own zone: that is the calendar the deadline was written
	// against.
	date := ""
	if e.SentAt != nil {
		date = fmt.Sprintf("Date: %s\n", e.SentAt.Format("2006-01-02"))
	}
	return fmt.Sprintf("%sFrom: %s\nSubject: %s\n\n%s", date, e.From, e.Subject, body)
}

func parseDeals(text string) ([]Deal, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return []Deal{}, nil
	}
	var deals []Deal
	if err := json.Unmarshal([]byte(text), &deals); err != nil {
		return nil, fmt.Errorf("parse deals json: %w", err)
	}
	if deals == nil {
		deals = []Deal{}
	}
	return deals, nil
}

// backoff returns the base delay for the given attempt (1-indexed retry) plus
// jitter of up to a quarter of the base.
func (g *Gemini) backoff(attempt int) time.Duration {
	delays := g.RetryDelays
	if len(delays) == 0 {
		delays = defaultRetryDelays
	}
	idx := attempt - 1
	if idx >= len(delays) {
		idx = len(delays) - 1
	}
	base := delays[idx]
	if base <= 0 {
		return 0
	}
	return base + time.Duration(rand.Int64N(int64(base)/4+1))
}

func isRetriable(err error) bool {
	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == 429 || apiErr.Code >= 500
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
