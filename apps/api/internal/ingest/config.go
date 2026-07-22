// Package ingest wires the email-deals pipeline into the API server: config
// loading, the bearer-token middleware, and the HTTP handlers that turn a
// forwarded email into stored, extracted, and digested deals. All domain logic
// lives in the framework-free internal packages (mailparse, store, extract,
// digest); this package is the only email-pipeline code that touches Echo.
package ingest

import (
	"fmt"
	"time"
)

// Config holds the email-ingest pipeline's runtime configuration, loaded from
// the environment. Durations (DigestInterval, RepostAfter) use Go's
// time.ParseDuration syntax — hours/minutes/seconds only, no "d" unit; express
// 45 days as "1080h". PORT and the WorkOS vars are read elsewhere (main.go).
type Config struct {
	DatabaseURL     string
	IngestToken     string
	GeminiAPIKey    string
	GeminiModel     string
	SlackWebhookURL string
	DigestInterval  time.Duration
	RepostAfter     time.Duration
}

// Load builds a Config from getenv (normally os.Getenv). GeminiModel and the two
// durations get defaults; the rest default to empty. It errors only on a
// malformed value (an unparseable or non-positive duration); a missing value is
// not an error, so the server still boots static-only when the pipeline is not
// configured. The Slack webhook comes from SLACK_WEBHOOK_FOR_DEALS_CHANNEL.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		GeminiModel:    "gemini-3.5-flash-lite",
		DigestInterval: 24 * time.Hour,
		RepostAfter:    45 * 24 * time.Hour,
	}
	cfg.DatabaseURL = getenv("DATABASE_URL")
	cfg.IngestToken = getenv("INGEST_TOKEN")
	cfg.GeminiAPIKey = getenv("GEMINI_API_KEY")
	if v := getenv("GEMINI_MODEL"); v != "" {
		cfg.GeminiModel = v
	}
	cfg.SlackWebhookURL = getenv("SLACK_WEBHOOK_FOR_DEALS_CHANNEL")
	if v := getenv("DIGEST_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid DIGEST_INTERVAL %q: %w", v, err)
		}
		cfg.DigestInterval = d
	}
	if v := getenv("REPOST_AFTER"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid REPOST_AFTER %q: %w", v, err)
		}
		cfg.RepostAfter = d
	}
	// Both durations feed time-window math that misbehaves on non-positive
	// values; reject them at load time.
	if cfg.DigestInterval <= 0 {
		return Config{}, fmt.Errorf("DIGEST_INTERVAL must be positive, got %v", cfg.DigestInterval)
	}
	if cfg.RepostAfter <= 0 {
		return Config{}, fmt.Errorf("REPOST_AFTER must be positive, got %v", cfg.RepostAfter)
	}
	return cfg, nil
}

// Enabled reports whether the pipeline has the minimum configuration to run: a
// database to store into and a shared token to authenticate the Worker. When it
// is false the server runs static-only and the /api/ingest routes are not
// registered.
func (c Config) Enabled() bool {
	return c.DatabaseURL != "" && c.IngestToken != ""
}

// String redacts secret fields so a Config is safe to log with %v/%s/%+v.
func (c Config) String() string {
	redact := func(s string) string {
		if s == "" {
			return ""
		}
		return "***"
	}
	return fmt.Sprintf(
		"ingest.Config{DatabaseURL:%s IngestToken:%s GeminiAPIKey:%s GeminiModel:%s SlackWebhookURL:%s DigestInterval:%v RepostAfter:%v}",
		redact(c.DatabaseURL), redact(c.IngestToken), redact(c.GeminiAPIKey),
		c.GeminiModel, redact(c.SlackWebhookURL), c.DigestInterval, c.RepostAfter,
	)
}
