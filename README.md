# vibe-notify

[![CI](https://github.com/its-the-vibe/vibe-notify/actions/workflows/ci.yaml/badge.svg)](https://github.com/its-the-vibe/vibe-notify/actions/workflows/ci.yaml)

A command-line tool for importing and broadcasting GitHub issues and pull requests to Slack.

## Requirements

- Go 1.24+
- [GitHub CLI (`gh`)](https://cli.github.com/) – used to fetch issue and pull request details using your existing authentication

> **Note:** If the `gh` CLI is not installed, `vibe-notify` will fall back to direct GitHub API calls using the `github_token` from your config. Install and authenticate the `gh` CLI with:
> ```sh
> # Install: https://cli.github.com/
> gh auth login
> ```

## Installation

```sh
git clone https://github.com/its-the-vibe/vibe-notify.git
cd vibe-notify
make build
# Binary is placed in bin/vibe-notify
```

## Configuration

Copy the example config and fill in your values:

```sh
cp config.example.yaml config.yaml
```

Edit `config.yaml`:

```yaml
# SlackLiner service base URL (required)
slackliner_url: "http://localhost:8080"

# Slack channel to post to (optional – defaults to the SlackLiner/bot default)
slack_channel: "#general"

# How long (in seconds) the message is kept before auto-deletion via TimeBomb (0 = never)
message_ttl: 0

# GitHub personal access token (optional – only needed for private repos)
github_token: ""
```

> **Note:** `config.yaml` is listed in `.gitignore` and will never be committed.

## Usage

### Broadcast a GitHub Issue

```sh
vibe-notify issue <issue-url>
```

**Example:**

```sh
vibe-notify issue https://github.com/its-the-vibe/vibe-notify/issues/1
```

This fetches the issue details using the `gh` CLI (falling back to the GitHub API if `gh` is not installed) and posts a formatted message to the configured Slack channel.

By default `vibe-notify` looks for `config.yaml` in the current directory. Use the `--config` flag to specify a different path:

```sh
vibe-notify --config /path/to/my-config.yaml issue https://github.com/owner/repo/issues/42
```

## Development

```sh
# Run tests
make test

# Run linter
make lint

# Build binary
make build

# Clean build artefacts
make clean
```
