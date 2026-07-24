// Package ingest wires the email-deals pipeline into the API server: config
// loading, the bearer-token middleware, and the HTTP handlers that turn a
// forwarded email into stored, extracted, and digested deals. All domain logic
// lives in the framework-free internal packages (mailparse, store, extract,
// digest); this package is the only email-pipeline code that touches Echo.
package ingest

import (
	"fmt"
	"net/mail"
	"strings"
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
	// ApprovedSenders is the set of normalized addresses whose mail bypasses the
	// DigestInterval wait (see IsApprovedSender). Nil/empty means nobody.
	ApprovedSenders map[string]bool
}

// Load builds a Config from getenv (normally os.Getenv). GeminiModel and the two
// durations get defaults; the rest default to empty. It errors only on a
// malformed value (an unparseable or non-positive duration); a missing value is
// not an error, so the server still boots static-only when the pipeline is not
// configured. The Slack webhook comes from SLACK_WEBHOOK_FOR_DEALS_CHANNEL and
// the approved-sender allowlist from APPROVED_FORWARDING_EMAILS.
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
	cfg.ApprovedSenders = parseApprovedSenders(getenv("APPROVED_FORWARDING_EMAILS"))
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

// parseApprovedSenders turns a comma-separated APPROVED_FORWARDING_EMAILS value
// into a lookup set of normalized addresses. Blank entries are dropped, so an
// unset or all-whitespace value yields an empty set (nobody is approved) rather
// than an error — a typo must not take the pipeline down.
func parseApprovedSenders(v string) map[string]bool {
	set := map[string]bool{}
	for _, part := range strings.Split(v, ",") {
		if addr := normalizeAddress(part); addr != "" {
			set[addr] = true
		}
	}
	return set
}

// normalizeAddress reduces a From / X-Envelope-From header value (or a configured
// allowlist entry) to a comparable bare address: display name stripped, trimmed,
// lowercased. A value net/mail cannot parse is compared as-is, which still
// matches a plain "alice@example.com" written without angle brackets.
func normalizeAddress(s string) string {
	if a, err := mail.ParseAddress(s); err == nil {
		s = a.Address
	}
	return strings.ToLower(strings.TrimSpace(s))
}

// IsApprovedSender reports whether any of the given From / X-Envelope-From header
// values is on the APPROVED_FORWARDING_EMAILS allowlist. Such a sender's email
// posts its digest immediately instead of waiting out DigestInterval.
//
// Both headers are checked because the two forwarding styles differ: a
// hand-composed "Forward" puts the forwarder in From:, while an auto-forward rule
// preserves the original newsletter's From: and only changes the envelope sender.
func (c Config) IsApprovedSender(headers ...string) bool {
	if len(c.ApprovedSenders) == 0 {
		return false
	}
	for _, h := range headers {
		if addr := normalizeAddress(h); addr != "" && c.ApprovedSenders[addr] {
			return true
		}
	}
	return false
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
	// The allowlist is secret-managed too, so log only how many addresses it holds.
	return fmt.Sprintf(
		"ingest.Config{DatabaseURL:%s IngestToken:%s GeminiAPIKey:%s GeminiModel:%s SlackWebhookURL:%s DigestInterval:%v RepostAfter:%v ApprovedSenders:%d}",
		redact(c.DatabaseURL), redact(c.IngestToken), redact(c.GeminiAPIKey),
		c.GeminiModel, redact(c.SlackWebhookURL), c.DigestInterval, c.RepostAfter,
		len(c.ApprovedSenders),
	)
}
