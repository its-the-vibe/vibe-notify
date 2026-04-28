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

# How long the message is kept before auto-deletion via TimeBomb (0 = never).
# Accepts a Go duration string (e.g. "48h", "1h30m", "30m") or a plain integer in seconds.
message_ttl: "48h"

# GitHub personal access token (optional – only needed for private repos)
# Can also be set via the GITHUB_TOKEN environment variable or a .env file.
github_token: ""
```

### GitHub Token via `.env`

You can provide the GitHub token through a `.env` file instead of (or in addition to) the config file. Copy the example and set your token:

```sh
cp .env.example .env
```

Edit `.env`:

```sh
GITHUB_TOKEN=your_github_token_here
```

The token is resolved in the following order (first non-empty value wins):
1. `github_token` field in `config.yaml`
2. `GITHUB_TOKEN` environment variable (including values loaded from `.env`)

> **Note:** Both `config.yaml` and `.env` are listed in `.gitignore` and will never be committed.

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
