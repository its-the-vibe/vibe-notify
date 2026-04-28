package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Message is the payload sent to a Slack incoming webhook.
type Message struct {
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text"`
}

// Client sends messages to Slack via an incoming webhook.
type Client struct {
	webhookURL string
	httpClient *http.Client
}

// NewClient returns a new Slack Client for the given webhook URL.
func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{},
	}
}

// NewClientWithHTTP returns a Slack Client using a custom *http.Client.
// Useful for testing.
func NewClientWithHTTP(webhookURL string, hc *http.Client) *Client {
	return &Client{webhookURL: webhookURL, httpClient: hc}
}

// PostMessage sends msg to the configured Slack webhook.
func (c *Client) PostMessage(ctx context.Context, msg Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting to slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}
