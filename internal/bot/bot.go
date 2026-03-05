package bot

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	tgbot "github.com/go-telegram/bot"
)

// Bot represents the Telegram bot
type Bot struct {
	bot          *tgbot.Bot
	allowedUsers map[int64]bool
	debug        bool
	ai           AIConfig
}

// AIConfig holds provider settings used by summarization handlers.
type AIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Config holds bot configuration
type Config struct {
	Token        string
	AllowedUsers []int64
	Debug        bool
	AI           AIConfig
}

// New creates a new Telegram bot instance
func New(cfg *Config) (*Bot, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}
	if cfg.AI.APIKey == "" {
		return nil, fmt.Errorf("AI API key is required")
	}

	// Build allowed users map
	allowedUsers := make(map[int64]bool)
	for _, id := range cfg.AllowedUsers {
		allowedUsers[id] = true
	}

	// Create bot options (no default handler yet - we'll register it after)
	opts := []tgbot.Option{}

	if cfg.Debug {
		opts = append(opts, tgbot.WithDebug())
	}

	// Initialize telegram bot
	tgBotInstance, err := tgbot.New(cfg.Token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot := &Bot{
		bot:          tgBotInstance,
		allowedUsers: allowedUsers,
		debug:        cfg.Debug,
		ai:           cfg.AI,
	}

	// Now register all handlers (including the default message handler)
	bot.registerHandlers()

	return bot, nil
}

// Start begins polling for updates
func (b *Bot) Start(ctx context.Context) error {
	fmt.Println("🤖 Bot started! Press Ctrl+C to stop.")
	b.bot.Start(ctx)
	return nil
}

// StartWebhook begins webhook mode (for future use)
// This allows easy migration from polling to webhooks
func (b *Bot) StartWebhook(ctx context.Context, addr string, webhookURL string) error {
	// Set webhook
	_, err := b.bot.SetWebhook(ctx, &tgbot.SetWebhookParams{
		URL: webhookURL,
	})
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	fmt.Printf("🤖 Bot started in webhook mode on %s\n", addr)
	fmt.Printf("📡 Webhook URL: %s\n", webhookURL)

	// Start webhook receiver in background
	go b.bot.StartWebhook(ctx)

	// Start HTTP server
	return http.ListenAndServe(addr, b.bot.WebhookHandler())
}

// registerHandlers registers all message handlers
func (b *Bot) registerHandlers() {
	// Register command handlers
	b.bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, b.handleStart)
	b.bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/help", tgbot.MatchTypeExact, b.handleHelp)
	b.bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/status", tgbot.MatchTypeExact, b.handleStatus)

	// Register default handler for all other messages (URL processing)
	b.bot.RegisterHandler(tgbot.HandlerTypeMessageText, "", tgbot.MatchTypePrefix, b.HandleMessage)
}

// IsAllowed checks if a user is allowed to use the bot
func (b *Bot) IsAllowed(userID int64) bool {
	// If no allowlist is configured, allow everyone
	if len(b.allowedUsers) == 0 {
		return true
	}
	return b.allowedUsers[userID]
}

// GetBot returns the underlying telegram bot instance
func (b *Bot) GetBot() *tgbot.Bot {
	return b.bot
}

// RunBot is a convenience function to run the bot with graceful shutdown
func RunBot(cfg *Config) error {
	// Create context with cancel for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create bot
	b, err := New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	// Start bot
	return b.Start(ctx)
}
