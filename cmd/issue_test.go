package cmd

import (
	"testing"

	gh "github.com/its-the-vibe/vibe-notify/internal/github"
)

func TestBuildIssueMessage(t *testing.T) {
	issue := &gh.Issue{
		Number:  42,
		Title:   "Fix the bug",
		HTMLURL: "https://github.com/owner/repo/issues/42",
		State:   "open",
	}
	issue.User.Login = "alice"

	got := buildIssueMessage(issue, "owner/repo")
	want := "🐛 *GitHub Issue #42*\n\n*Repository:* owner/repo\n*Title:* Fix the bug\n*Author:* @alice\n*State:* open\n*URL:* https://github.com/owner/repo/issues/42"
	if got != want {
		t.Errorf("message mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestBuildIssueMetadata(t *testing.T) {
	issue := &gh.Issue{
		Number:  42,
		Title:   "Fix the bug",
		HTMLURL: "https://github.com/owner/repo/issues/42",
		State:   "open",
	}
	issue.User.Login = "alice"

	meta := buildIssueMetadata(issue, "owner/repo")

	if meta["event_type"] != issueBroadcastEventType {
		t.Errorf("event_type: got %v, want %v", meta["event_type"], issueBroadcastEventType)
	}

	payload, ok := meta["event_payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("event_payload is not a map")
	}
	checks := map[string]interface{}{
		"title":        "Fix the bug",
		"issue_number": 42,
		"issue_url":    "https://github.com/owner/repo/issues/42",
		"repository":   "owner/repo",
		"author":       "alice",
		"state":        "open",
	}
	for k, want := range checks {
		if got := payload[k]; got != want {
			t.Errorf("event_payload[%q]: got %v, want %v", k, got, want)
		}
	}
}

