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

	got := buildIssueMessage(issue)
	want := "🐛 *GitHub Issue #42*\n\n*Title:* Fix the bug\n*Author:* @alice\n*State:* open\n*URL:* https://github.com/owner/repo/issues/42"
	if got != want {
		t.Errorf("message mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}
