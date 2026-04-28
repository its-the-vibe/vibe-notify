package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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

// ghExecCommand is the function used to create exec.Cmd for the gh CLI.
// It can be overridden in tests.
var ghExecCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, arg...)
}

// GetIssueViaCLI fetches a GitHub issue using the gh CLI tool.
// It requires the gh CLI to be installed and authenticated.
func GetIssueViaCLI(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	cmd := ghExecCommand(ctx, "gh", "issue", "view", strconv.Itoa(number),
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--json", "number,title,url,body,author,state",
	)

	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("gh CLI not found: install it from https://cli.github.com and run 'gh auth login': %w", exec.ErrNotFound)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh CLI error: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("running gh CLI: %w", err)
	}

	var raw struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		URL    string `json:"url"`
		Body   string `json:"body"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		State string `json:"state"`
	}

	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parsing gh CLI output: %w", err)
	}

	issue := &Issue{
		Number:  raw.Number,
		Title:   raw.Title,
		HTMLURL: raw.URL,
		Body:    raw.Body,
		State:   strings.ToLower(raw.State),
	}
	issue.User.Login = raw.Author.Login

	return issue, nil
}

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
