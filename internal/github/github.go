package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

// IssueURL is the parsed representation of a GitHub issue URL.
type IssueURL struct {
	Owner  string
	Repo   string
	Number int
}

// issueURLPattern matches https://github.com/<owner>/<repo>/issues/<number>
var issueURLPattern = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/issues/(\d+)`)

// ParseIssueURL parses a GitHub issue URL and returns its components.
func ParseIssueURL(rawURL string) (IssueURL, error) {
	m := issueURLPattern.FindStringSubmatch(rawURL)
	if m == nil {
		return IssueURL{}, fmt.Errorf("invalid GitHub issue URL: %q", rawURL)
	}
	num, _ := strconv.Atoi(m[3])
	return IssueURL{Owner: m[1], Repo: m[2], Number: num}, nil
}

// Issue represents the relevant fields of a GitHub issue API response.
type Issue struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
	State string `json:"state"`
}

// Client is a minimal GitHub REST API client.
type Client struct {
	httpClient *http.Client
	token      string
}

// NewClient returns a new GitHub Client. token may be empty for public repos.
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{},
		token:      token,
	}
}

// NewClientWithTransport returns a Client that uses the provided RoundTripper.
// Useful for testing with a custom transport that redirects requests to a test server.
func NewClientWithTransport(token string, transport http.RoundTripper) *Client {
	return &Client{
		httpClient: &http.Client{Transport: transport},
		token:      token,
	}
}

// GetIssue fetches a GitHub issue by owner, repo, and issue number.
func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for %s", resp.StatusCode, url)
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("decoding issue response: %w", err)
	}

	return &issue, nil
}
