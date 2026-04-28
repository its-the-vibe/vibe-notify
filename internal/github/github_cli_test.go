package github

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestGetIssueViaCLI_Success(t *testing.T) {
	payload := `{"number":42,"title":"Fix the bug","url":"https://github.com/owner/repo/issues/42","body":"details","author":{"login":"alice"},"state":"OPEN"}`

	orig := ghExecCommand
	ghExecCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		// Pass payload as $1 via "--"; printf '%s' "$@" prints it.
		return exec.CommandContext(ctx, "sh", "-c", `printf '%s' "$@"`, "--", payload)
	}
	defer func() { ghExecCommand = orig }()

	issue, err := GetIssueViaCLI(context.Background(), "owner", "repo", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Number != 42 {
		t.Errorf("number: got %d, want 42", issue.Number)
	}
	if issue.Title != "Fix the bug" {
		t.Errorf("title: got %q, want %q", issue.Title, "Fix the bug")
	}
	if issue.HTMLURL != "https://github.com/owner/repo/issues/42" {
		t.Errorf("htmlURL: got %q", issue.HTMLURL)
	}
	if issue.User.Login != "alice" {
		t.Errorf("login: got %q, want alice", issue.User.Login)
	}
	// gh CLI returns uppercase state; GetIssueViaCLI must normalise to lowercase.
	if issue.State != "open" {
		t.Errorf("state: got %q, want %q", issue.State, "open")
	}
}

func TestGetIssueViaCLI_GhNotInstalled(t *testing.T) {
	orig := ghExecCommand
	ghExecCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		// Simulate gh not being installed by running a non-existent binary.
		return exec.CommandContext(ctx, "this-binary-does-not-exist-vibe-notify-test")
	}
	defer func() { ghExecCommand = orig }()

	_, err := GetIssueViaCLI(context.Background(), "owner", "repo", 1)
	if err == nil {
		t.Fatal("expected error when gh is not installed, got nil")
	}
	if !errors.Is(err, exec.ErrNotFound) {
		t.Errorf("expected exec.ErrNotFound in error chain, got: %v", err)
	}
}

func TestGetIssueViaCLI_CLIError(t *testing.T) {
	orig := ghExecCommand
	ghExecCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		// gh exits non-zero, simulating an authentication or API error.
		return exec.CommandContext(ctx, "sh", "-c", "echo 'authentication required' >&2; exit 1")
	}
	defer func() { ghExecCommand = orig }()

	_, err := GetIssueViaCLI(context.Background(), "owner", "repo", 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetIssueViaCLI_StateLowercase(t *testing.T) {
	for _, state := range []string{"OPEN", "CLOSED"} {
		state := state
		payload := `{"number":1,"title":"T","url":"https://github.com/o/r/issues/1","body":"","author":{"login":"u"},"state":"` + state + `"}`

		orig := ghExecCommand
		ghExecCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "sh", "-c", `printf '%s' "$@"`, "--", payload)
		}

		issue, err := GetIssueViaCLI(context.Background(), "o", "r", 1)
		ghExecCommand = orig

		if err != nil {
			t.Fatalf("state=%s: unexpected error: %v", state, err)
		}
		want := strings.ToLower(state)
		if issue.State != want {
			t.Errorf("state=%s: got %q, want %q", state, issue.State, want)
		}
	}
}
