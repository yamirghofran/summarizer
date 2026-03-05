package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yamirghofran/summarizer/internal/bot"
	"github.com/yamirghofran/summarizer/internal/config"
)

var (
	botToken string
	botDebug bool
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Start the Telegram bot",
	Long: `Start the Telegram bot to summarize YouTube videos and web pages.

The bot will:
• Listen for messages containing URLs
• Automatically detect YouTube vs web page URLs
• Process and summarize the content
• Send the summary back to the user

Configuration files:
  ~/.config/summarizer/config.toml
  ~/.local/share/summarizer/credentials.toml

Initialize defaults with:
  summarizer config init

To get your Telegram user ID, message @userinfobot on Telegram.`,
	Run: runBot,
}

func init() {
	rootCmd.AddCommand(botCmd)

	botCmd.Flags().StringVar(&botToken, "token", "", "Telegram bot token (overrides credentials file)")
	botCmd.Flags().BoolVar(&botDebug, "debug", false, "Enable debug logging")
}

func runBot(cmd *cobra.Command, args []string) {
	loaded, err := config.Load("", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	aiSettings, err := loaded.AISettings("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	token, err := loaded.TelegramBotToken(botToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Get a token from @BotFather on Telegram")
		os.Exit(1)
	}

	allowedUsers := loaded.Config.Telegram.AllowedUserIDs
	if len(allowedUsers) > 0 {
		fmt.Printf("🔒 Allowlist enabled: %d user(s) allowed\n", len(allowedUsers))
	} else {
		fmt.Println("🌐 Open access: All users allowed")
	}

	// Create bot config
	cfg := &bot.Config{
		Token:        token,
		AllowedUsers: allowedUsers,
		Debug:        botDebug,
		AI: bot.AIConfig{
			APIKey:  aiSettings.APIKey,
			BaseURL: aiSettings.BaseURL,
			Model:   aiSettings.Model,
		},
	}

	// Run bot
	if err := bot.RunBot(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
