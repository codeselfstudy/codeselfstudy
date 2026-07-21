package digest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WebhookPoster posts a prebuilt payload to a Slack incoming webhook. It is an
// interface so tests can substitute a fake without a network.
type WebhookPoster interface {
	Post(ctx context.Context, payload []byte) error
}

// HTTPPoster posts to a Slack incoming webhook URL over HTTP.
type HTTPPoster struct {
	URL    string
	Client *http.Client
}

// NewHTTPPoster returns an HTTPPoster with a sensible timeout.
func NewHTTPPoster(url string) *HTTPPoster {
	return &HTTPPoster{URL: url, Client: &http.Client{Timeout: 10 * time.Second}}
}

// Post sends the payload as JSON. Any non-2xx response is an error.
func (p *HTTPPoster) Post(ctx context.Context, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return fmt.Errorf("post to slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("slack webhook returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
