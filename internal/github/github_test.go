package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gh "github.com/its-the-vibe/vibe-notify/internal/github"
)

func TestParseIssueURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    gh.IssueURL
		wantErr bool
	}{
		{
			name: "valid URL",
			url:  "https://github.com/its-the-vibe/vibe-notify/issues/1",
			want: gh.IssueURL{Owner: "its-the-vibe", Repo: "vibe-notify", Number: 1},
		},
		{
			name: "valid URL with query string",
			url:  "https://github.com/owner/repo/issues/42?foo=bar",
			want: gh.IssueURL{Owner: "owner", Repo: "repo", Number: 42},
		},
		{
			name:    "invalid URL - PR",
			url:     "https://github.com/owner/repo/pull/1",
			wantErr: true,
		},
		{
			name:    "invalid URL - no number",
			url:     "https://github.com/owner/repo/issues/",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := gh.ParseIssueURL(tc.url)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestGetIssue(t *testing.T) {
	issue := gh.Issue{
		Number:  7,
		Title:   "Test Issue",
		HTMLURL: "https://github.com/owner/repo/issues/7",
		State:   "open",
	}
	issue.User.Login = "testuser"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/issues/7" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
	}))
	defer srv.Close()

	// We can't easily inject base URL without restructuring, so test with a real-ish client
	// by using a custom transport that rewrites the host.
	client := gh.NewClient("")
	_ = client
	// Minimal smoke test: just ensure NewClient doesn't panic.
}

func TestGetIssue_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := gh.NewClient("")
	// We can't hit the test server without URL injection support.
	// The non-server path is covered by the unit tests above.
	_ = client
}

// TestGetIssueWithServer exercises GetIssue end-to-end against a local test server
// by temporarily replacing the base URL via a round-tripper.
type baseURLTransport struct {
	base    string
	wrapped http.RoundTripper
}

func (t *baseURLTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite scheme+host to the test server base
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = "http"
	req2.URL.Host = t.base
	return t.wrapped.RoundTrip(req2)
}

func TestGetIssueWithServer_Success(t *testing.T) {
	expected := gh.Issue{
		Number:  5,
		Title:   "Hello Issue",
		HTMLURL: "https://github.com/o/r/issues/5",
		State:   "open",
	}
	expected.User.Login = "alice"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	transport := &baseURLTransport{base: srv.Listener.Addr().String(), wrapped: http.DefaultTransport}
	client := gh.NewClientWithTransport("", transport)

	issue, err := client.GetIssue(context.Background(), "o", "r", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Title != expected.Title {
		t.Errorf("title: got %q, want %q", issue.Title, expected.Title)
	}
	if issue.User.Login != "alice" {
		t.Errorf("user: got %q, want %q", issue.User.Login, "alice")
	}
}

func TestGetIssueWithServer_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	transport := &baseURLTransport{base: srv.Listener.Addr().String(), wrapped: http.DefaultTransport}
	client := gh.NewClientWithTransport("", transport)

	_, err := client.GetIssue(context.Background(), "o", "r", 99)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}
