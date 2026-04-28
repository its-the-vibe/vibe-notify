package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/its-the-vibe/vibe-notify/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTemp(t, `
slackliner_url: "http://slackliner.example.com"
slack_channel: "#general"
message_ttl: 3600
github_token: "ghp_test"
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.SlackLinerURL != "http://slackliner.example.com" {
		t.Errorf("unexpected SlackLiner URL: %s", cfg.SlackLinerURL)
	}
	if cfg.SlackChannel != "#general" {
		t.Errorf("unexpected channel: %s", cfg.SlackChannel)
	}
	if cfg.MessageTTL != 3600 {
		t.Errorf("unexpected TTL: %d", cfg.MessageTTL)
	}
	if cfg.GitHubToken != "ghp_test" {
		t.Errorf("unexpected token: %s", cfg.GitHubToken)
	}
}

func TestLoad_MessageTTL_DurationString(t *testing.T) {
	tests := []struct {
		name    string
		ttl     string
		wantSec int
	}{
		{"hours", `"48h"`, 48 * 3600},
		{"minutes", `"30m"`, 30 * 60},
		{"combined", `"1h30m"`, 90 * 60},
		{"integer seconds", "7200", 7200},
		{"zero", "0", 0},
		{"empty", "", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content := `slackliner_url: "http://example.com"` + "\n"
			if tc.ttl != "" {
				content += "message_ttl: " + tc.ttl + "\n"
			}
			path := writeTemp(t, content)
			cfg, err := config.Load(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.MessageTTL != tc.wantSec {
				t.Errorf("message_ttl %q: got %d seconds, want %d", tc.ttl, cfg.MessageTTL, tc.wantSec)
			}
		})
	}
}

func TestLoad_MessageTTL_InvalidValue(t *testing.T) {
	path := writeTemp(t, `
slackliner_url: "http://example.com"
message_ttl: "notaduration"
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid message_ttl, got nil")
	}
}

func TestLoad_GitHubToken_FromEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env_token")

	path := writeTemp(t, `slackliner_url: "http://example.com"`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GitHubToken != "env_token" {
		t.Errorf("expected token from env, got %q", cfg.GitHubToken)
	}
}

func TestLoad_GitHubToken_ConfigTakesPrecedence(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env_token")

	path := writeTemp(t, `
slackliner_url: "http://example.com"
github_token: "config_token"
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GitHubToken != "config_token" {
		t.Errorf("expected config token to take precedence, got %q", cfg.GitHubToken)
	}
}

func TestLoad_MissingSlackLinerURL(t *testing.T) {
	path := writeTemp(t, `
slack_channel: "#general"
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing slackliner_url, got nil")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTemp(t, `:::invalid yaml:::`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

