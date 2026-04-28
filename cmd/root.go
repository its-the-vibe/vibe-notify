package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd is the base command for vibe-notify.
var rootCmd = &cobra.Command{
	Use:   "vibe-notify",
	Short: "Broadcast GitHub issues and pull requests to Slack",
	Long: `vibe-notify is a CLI tool that fetches GitHub issue and pull request
details and broadcasts them to a configured Slack channel.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "path to config file")
}
