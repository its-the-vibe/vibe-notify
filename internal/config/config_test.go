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
slack_webhook_url: "https://hooks.slack.com/services/T/B/X"
slack_channel: "#general"
message_ttl: 3600
github_token: "ghp_test"
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.SlackWebhookURL != "https://hooks.slack.com/services/T/B/X" {
		t.Errorf("unexpected webhook URL: %s", cfg.SlackWebhookURL)
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

func TestLoad_MissingWebhook(t *testing.T) {
	path := writeTemp(t, `
slack_channel: "#general"
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing slack_webhook_url, got nil")
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
