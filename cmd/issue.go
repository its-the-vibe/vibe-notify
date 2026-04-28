package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/its-the-vibe/vibe-notify/internal/config"
	gh "github.com/its-the-vibe/vibe-notify/internal/github"
	"github.com/its-the-vibe/vibe-notify/internal/slack"
)

var issueCmd = &cobra.Command{
	Use:   "issue <issue-url>",
	Short: "Broadcast a GitHub issue to Slack",
	Long: `Fetch details about a GitHub issue and post a notification to the
configured Slack channel.

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

	ghClient := gh.NewClient(cfg.GitHubToken)
	issue, err := ghClient.GetIssue(ctx, parsed.Owner, parsed.Repo, parsed.Number)
	if err != nil {
		return fmt.Errorf("fetching GitHub issue: %w", err)
	}

	text := buildIssueMessage(issue)

	slackClient := slack.NewClient(cfg.SlackWebhookURL)
	msg := slack.Message{
		Text:    text,
		Channel: cfg.SlackChannel,
	}
	if err := slackClient.PostMessage(ctx, msg); err != nil {
		return fmt.Errorf("posting to Slack: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✅ Posted issue #%d to Slack: %s\n", issue.Number, issue.HTMLURL)
	return nil
}

// buildIssueMessage formats the Slack message for a GitHub issue.
func buildIssueMessage(issue *gh.Issue) string {
	return fmt.Sprintf(
		"🐛 *GitHub Issue #%d*\n\n*Title:* %s\n*Author:* @%s\n*State:* %s\n*URL:* %s",
		issue.Number,
		issue.Title,
		issue.User.Login,
		issue.State,
		issue.HTMLURL,
	)
}
