package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/its-the-vibe/vibe-notify/internal/config"
	gh "github.com/its-the-vibe/vibe-notify/internal/github"
	"github.com/its-the-vibe/vibe-notify/internal/slack"
)

const issueBroadcastEventType = "issue_broadcast"

var issueCmd = &cobra.Command{
	Use:   "issue <issue-url>",
	Short: "Broadcast a GitHub issue to Slack",
	Long: `Fetch details about a GitHub issue and post a notification to the
configured Slack channel via SlackLiner.

Example:
  vibe-notify issue https://github.com/owner/repo/issues/42`,
	Args: cobra.ExactArgs(1),
	RunE: runIssue,
}

func init() {
	rootCmd.AddCommand(issueCmd)
}

func runIssue(cmd *cobra.Command, args []string) error {
	issueURL := args[0]

	parsed, err := gh.ParseIssueURL(issueURL)
	if err != nil {
		return err
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx := context.Background()

	// Prefer the gh CLI for authentication; fall back to the HTTP API if gh is not installed.
	issue, err := gh.GetIssueViaCLI(ctx, parsed.Owner, parsed.Repo, parsed.Number)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintln(os.Stderr, "⚠️  gh CLI not found – falling back to GitHub API (authentication via config token)")
			ghClient := gh.NewClient(cfg.GitHubToken)
			issue, err = ghClient.GetIssue(ctx, parsed.Owner, parsed.Repo, parsed.Number)
			if err != nil {
				return fmt.Errorf("fetching GitHub issue: %w", err)
			}
		} else {
			return fmt.Errorf("fetching GitHub issue via gh CLI: %w", err)
		}
	}

	repository := parsed.Owner + "/" + parsed.Repo
	text := buildIssueMessage(issue, repository)
	metadata := buildIssueMetadata(issue, repository)

	slackClient := slack.NewClient(cfg.SlackLinerURL)
	msg := slack.Message{
		Channel:  cfg.SlackChannel,
		Text:     text,
		TTL:      cfg.MessageTTL,
		Metadata: metadata,
	}
	resp, err := slackClient.PostMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("posting to SlackLiner: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✅ Posted issue #%d to Slack (channel=%s ts=%s): %s\n",
		issue.Number, resp.Channel, resp.Ts, issue.HTMLURL)
	return nil
}

// buildIssueMessage formats the Slack message text for a GitHub issue.
func buildIssueMessage(issue *gh.Issue, repository string) string {
	return fmt.Sprintf(
		"🐛 *GitHub Issue #%d*\n\n*Repository:* %s\n*Title:* %s\n*Author:* @%s\n*State:* %s\n*URL:* %s",
		issue.Number,
		repository,
		issue.Title,
		issue.User.Login,
		issue.State,
		issue.HTMLURL,
	)
}

// buildIssueMetadata constructs the Slack message metadata for an issue broadcast event.
func buildIssueMetadata(issue *gh.Issue, repository string) map[string]interface{} {
	return map[string]interface{}{
		"event_type": issueBroadcastEventType,
		"event_payload": map[string]interface{}{
			"title":        issue.Title,
			"issue_number": issue.Number,
			"issue_url":    issue.HTMLURL,
			"repository":   repository,
			"author":       issue.User.Login,
			"state":        issue.State,
		},
	}
}
