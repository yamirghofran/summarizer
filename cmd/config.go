package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage summarizer configuration files",
	Long: `Manage configuration and credentials files used by summarize and bot commands.

Configuration is loaded from XDG paths by default:
  ~/.config/summarizer/config.toml
  ~/.local/share/summarizer/credentials.toml`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
