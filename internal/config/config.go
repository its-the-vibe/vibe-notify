package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration values for vibe-notify.
type Config struct {
	SlackWebhookURL string `yaml:"slack_webhook_url"`
	SlackChannel    string `yaml:"slack_channel"`
	MessageTTL      int    `yaml:"message_ttl"`
	GitHubToken     string `yaml:"github_token"`
}

// Load reads a YAML config file from the given path and returns a Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if cfg.SlackWebhookURL == "" {
		return nil, fmt.Errorf("slack_webhook_url is required in config")
	}

	return &cfg, nil
}
