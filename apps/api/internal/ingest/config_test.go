package ingest

import (
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
	if cfg.GeminiModel != "gemini-2.5-flash-lite" {
		t.Errorf("GeminiModel = %q, want gemini-2.5-flash-lite", cfg.GeminiModel)
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
	}
	if cfg != want {
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

func TestConfigStringRedactsSecrets(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"INGEST_TOKEN":                    "supersecrettoken",
		"GEMINI_API_KEY":                  "supersecretkey",
		"SLACK_WEBHOOK_FOR_DEALS_CHANNEL": "https://hooks.slack.com/services/secret",
		"DATABASE_URL":                    "libsql://db.turso.io?authToken=secrettoken",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := cfg.String()
	for _, secret := range []string{"supersecrettoken", "supersecretkey", "secrettoken", "hooks.slack.com"} {
		if strings.Contains(s, secret) {
			t.Errorf("String() leaked secret %q: %s", secret, s)
		}
	}
	// Non-secret fields remain visible.
	if !strings.Contains(s, "gemini-2.5-flash-lite") {
		t.Errorf("String() dropped GeminiModel: %s", s)
	}
}
