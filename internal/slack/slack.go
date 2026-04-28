package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Message is the payload sent to the SlackLiner /message endpoint.
type Message struct {
	Channel  string                 `json:"channel,omitempty"`
	Text     string                 `json:"text"`
	TTL      int                    `json:"ttl,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Response is returned by the SlackLiner /message endpoint.
type Response struct {
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
}

// Client posts messages to a SlackLiner service via its HTTP API.
type Client struct {
	slacklinerURL string
	httpClient    *http.Client
}

// NewClient returns a new SlackLiner Client for the given base URL.
func NewClient(slacklinerURL string) *Client {
	return &Client{
		slacklinerURL: slacklinerURL,
		httpClient:    &http.Client{},
	}
}

// NewClientWithHTTP returns a SlackLiner Client using a custom *http.Client.
// Useful for testing.
func NewClientWithHTTP(slacklinerURL string, hc *http.Client) *Client {
	return &Client{slacklinerURL: slacklinerURL, httpClient: hc}
}

// PostMessage sends msg to the SlackLiner /message endpoint and returns the
// channel ID and message timestamp from the response.
func (c *Client) PostMessage(ctx context.Context, msg Message) (Response, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return Response{}, fmt.Errorf("marshaling slackliner message: %w", err)
	}

	url := strings.TrimRight(c.slacklinerURL, "/") + "/message"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return Response{}, fmt.Errorf("creating slackliner request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("posting to slackliner: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("slackliner returned status %d", resp.StatusCode)
	}

	var slResp Response
	if err := json.NewDecoder(resp.Body).Decode(&slResp); err != nil {
		return Response{}, fmt.Errorf("decoding slackliner response: %w", err)
	}

	return slResp, nil
}
