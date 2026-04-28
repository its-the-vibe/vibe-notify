package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// rawConfig is used to unmarshal the YAML file before post-processing.
type rawConfig struct {
	SlackLinerURL string `yaml:"slackliner_url"`
	SlackChannel  string `yaml:"slack_channel"`
	MessageTTL    string `yaml:"message_ttl"`
	GitHubToken   string `yaml:"github_token"`
}

// Config holds all configuration values for vibe-notify.
type Config struct {
	SlackLinerURL string
	SlackChannel  string
	// MessageTTL is the number of seconds a message is kept before auto-deletion (0 = never).
	MessageTTL  int
	GitHubToken string
}

// Load reads a YAML config file from the given path and returns a Config.
// It also loads a .env file from the current working directory if one exists
// (errors are silently ignored so the application works without a .env file),
// and falls back to the GITHUB_TOKEN environment variable when github_token is
// not set in the config file.
func Load(path string) (*Config, error) {
	// Load .env file if present; errors are intentionally ignored so the
	// application works without a .env file.
	_ = godotenv.Load()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if raw.SlackLinerURL == "" {
		return nil, fmt.Errorf("slackliner_url is required in config")
	}

	ttl, err := parseTTL(raw.MessageTTL)
	if err != nil {
		return nil, fmt.Errorf("parsing message_ttl: %w", err)
	}

	token := raw.GitHubToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	return &Config{
		SlackLinerURL: raw.SlackLinerURL,
		SlackChannel:  raw.SlackChannel,
		MessageTTL:    ttl,
		GitHubToken:   token,
	}, nil
}

// parseTTL converts a message_ttl value to an integer number of seconds.
// It accepts either a plain integer string (e.g. "3600") or a Go duration
// string (e.g. "48h", "1h30m"). An empty string is treated as 0 (no TTL).
func parseTTL(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	// Try parsing as a Go duration first (e.g. "48h", "1h30m").
	if d, err := time.ParseDuration(s); err == nil {
		return int(d.Seconds()), nil
	}

	// Fall back to a plain integer number of seconds for backward compatibility.
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, fmt.Errorf("value %q is not a valid duration or integer", s)
	}
	return n, nil
}
