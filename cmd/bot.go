package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yamirghofran/youtube-summarizer/internal/bot"
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

Required environment variables:
  TELEGRAM_BOT_TOKEN - Your Telegram bot token from @BotFather

Optional environment variables:
  ALLOWED_USER_IDS - Comma-separated list of allowed Telegram user IDs
                    (if not set, all users are allowed)

To get your Telegram user ID, message @userinfobot on Telegram.`,
	Run: runBot,
}

func init() {
	rootCmd.AddCommand(botCmd)

	botCmd.Flags().StringVar(&botToken, "token", "", "Telegram bot token (overrides TELEGRAM_BOT_TOKEN env var)")
	botCmd.Flags().BoolVar(&botDebug, "debug", false, "Enable debug logging")
}

func runBot(cmd *cobra.Command, args []string) {
	// Get bot token
	token := botToken
	if token == "" {
		token = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: TELEGRAM_BOT_TOKEN is required")
		fmt.Fprintln(os.Stderr, "Get a token from @BotFather on Telegram")
		os.Exit(1)
	}

	// Parse allowed users
	allowedUsersStr := os.Getenv("ALLOWED_USER_IDS")
	allowedUsers, err := bot.ParseAllowedUsers(allowedUsersStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing ALLOWED_USER_IDS: %v\n", err)
		os.Exit(1)
	}

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
	}

	// Run bot
	if err := bot.RunBot(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
