package ingest

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func envFrom(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(envFrom(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GeminiModel != "gemini-3.5-flash-lite" {
		t.Errorf("GeminiModel = %q, want gemini-3.5-flash-lite", cfg.GeminiModel)
	}
	if cfg.DigestInterval != 24*time.Hour {
		t.Errorf("DigestInterval = %v, want 24h", cfg.DigestInterval)
	}
	if cfg.RepostAfter != 45*24*time.Hour {
		t.Errorf("RepostAfter = %v, want 1080h", cfg.RepostAfter)
	}
	// No defaults for DatabaseURL or the secrets; the pipeline is disabled.
	if cfg.DatabaseURL != "" || cfg.IngestToken != "" || cfg.GeminiAPIKey != "" || cfg.SlackWebhookURL != "" {
		t.Errorf("expected empty DatabaseURL/secrets by default, got %+v", cfg)
	}
	if len(cfg.ApprovedSenders) != 0 {
		t.Errorf("ApprovedSenders = %v, want empty", cfg.ApprovedSenders)
	}
	if cfg.Enabled() {
		t.Error("Enabled() = true with no DATABASE_URL/INGEST_TOKEN, want false")
	}
}

func TestLoadOverrides(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"DATABASE_URL":                    "libsql://x.turso.io?authToken=t",
		"INGEST_TOKEN":                    "secret",
		"GEMINI_API_KEY":                  "key",
		"GEMINI_MODEL":                    "gemini-2.5-pro",
		"SLACK_WEBHOOK_FOR_DEALS_CHANNEL": "https://hooks.slack.com/services/x",
		"DIGEST_INTERVAL":                 "12h",
		"REPOST_AFTER":                    "720h",
		"APPROVED_FORWARDING_EMAILS":      "alice@example.com",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Config{
		DatabaseURL:     "libsql://x.turso.io?authToken=t",
		IngestToken:     "secret",
		GeminiAPIKey:    "key",
		GeminiModel:     "gemini-2.5-pro",
		SlackWebhookURL: "https://hooks.slack.com/services/x",
		DigestInterval:  12 * time.Hour,
		RepostAfter:     720 * time.Hour,
		ApprovedSenders: map[string]bool{"alice@example.com": true},
	}
	// DeepEqual, not ==: the ApprovedSenders map makes Config non-comparable.
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
	if !cfg.Enabled() {
		t.Error("Enabled() = false with DATABASE_URL+INGEST_TOKEN set, want true")
	}
}

func TestEnabledRequiresBoth(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"neither", nil, false},
		{"only db", map[string]string{"DATABASE_URL": "file:x"}, false},
		{"only token", map[string]string{"INGEST_TOKEN": "t"}, false},
		{"both", map[string]string{"DATABASE_URL": "file:x", "INGEST_TOKEN": "t"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(envFrom(tc.env))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.Enabled() != tc.want {
				t.Errorf("Enabled() = %v, want %v", cfg.Enabled(), tc.want)
			}
		})
	}
}

func TestLoadBadDuration(t *testing.T) {
	for _, key := range []string{"DIGEST_INTERVAL", "REPOST_AFTER"} {
		if _, err := Load(envFrom(map[string]string{key: "notaduration"})); err == nil {
			t.Errorf("expected error for bad %s", key)
		}
	}
}

func TestLoadNonPositiveDuration(t *testing.T) {
	for _, key := range []string{"DIGEST_INTERVAL", "REPOST_AFTER"} {
		for _, v := range []string{"0", "-5h"} {
			if _, err := Load(envFrom(map[string]string{key: v})); err == nil {
				t.Errorf("expected error for %s=%q", key, v)
			}
		}
	}
}

func TestLoadApprovedSenders(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want map[string]bool
	}{
		{"unset", "", map[string]bool{}},
		{"single", "alice@example.com", map[string]bool{"alice@example.com": true}},
		{
			"list is trimmed and lowercased",
			"A@Example.com, bob@x.io",
			map[string]bool{"a@example.com": true, "bob@x.io": true},
		},
		{
			"display names are stripped",
			"Alice <alice@x.io>,\tBob <BOB@X.io>",
			map[string]bool{"alice@x.io": true, "bob@x.io": true},
		},
		{"blank entries are dropped", ",,  ,", map[string]bool{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(envFrom(map[string]string{"APPROVED_FORWARDING_EMAILS": tc.env}))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if !reflect.DeepEqual(cfg.ApprovedSenders, tc.want) {
				t.Errorf("ApprovedSenders = %v, want %v", cfg.ApprovedSenders, tc.want)
			}
		})
	}
}

func TestIsApprovedSender(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{"APPROVED_FORWARDING_EMAILS": "alice@example.com"}))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	cases := []struct {
		name    string
		headers []string
		want    bool
	}{
		{"bare address", []string{"alice@example.com"}, true},
		{"display name and mixed case", []string{"Alice <ALICE@Example.com>"}, true},
		{"matches the second header", []string{"deals@humblebundle.com", "alice@example.com"}, true},
		{"no match", []string{"deals@humblebundle.com", "mallory@example.com"}, false},
		{"empty headers", []string{"", ""}, false},
		{"no headers", nil, false},
		// A substring of an approved address must not match.
		{"substring", []string{"alice@example.com.evil.test"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cfg.IsApprovedSender(tc.headers...); got != tc.want {
				t.Errorf("IsApprovedSender(%q) = %v, want %v", tc.headers, got, tc.want)
			}
		})
	}

	// An unconfigured allowlist approves nobody, including an empty header.
	empty, err := Load(envFrom(nil))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if empty.IsApprovedSender("alice@example.com", "") {
		t.Error("IsApprovedSender() = true with no allowlist, want false")
	}
}

func TestConfigStringRedactsSecrets(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"INGEST_TOKEN":                    "supersecrettoken",
		"GEMINI_API_KEY":                  "supersecretkey",
		"SLACK_WEBHOOK_FOR_DEALS_CHANNEL": "https://hooks.slack.com/services/secret",
		"DATABASE_URL":                    "libsql://db.turso.io?authToken=secrettoken",
		"APPROVED_FORWARDING_EMAILS":      "secretperson@example.com",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := cfg.String()
	for _, secret := range []string{"supersecrettoken", "supersecretkey", "secrettoken", "hooks.slack.com", "secretperson@example.com"} {
		if strings.Contains(s, secret) {
			t.Errorf("String() leaked secret %q: %s", secret, s)
		}
	}
	// Non-secret fields remain visible.
	if !strings.Contains(s, "gemini-3.5-flash-lite") {
		t.Errorf("String() dropped GeminiModel: %s", s)
	}
}
