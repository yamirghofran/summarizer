package bot

import (
	"context"
	"fmt"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Middleware is a function that wraps a handler
type Middleware func(next tgbot.HandlerFunc) tgbot.HandlerFunc

// WithAllowlist creates a middleware that checks if the user is allowed
func (b *Bot) WithAllowlist(next tgbot.HandlerFunc) tgbot.HandlerFunc {
	return func(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
		// Skip if no message
		if update.Message == nil {
			return
		}

		// Check if user is allowed
		userID := update.Message.From.ID
		if !b.IsAllowed(userID) {
			b.sendUnauthorizedMessage(ctx, tb, update.Message.Chat.ID)
			return
		}

		// Call next handler
		next(ctx, tb, update)
	}
}

// WithLogging creates a middleware that logs incoming messages
func (b *Bot) WithLogging(next tgbot.HandlerFunc) tgbot.HandlerFunc {
	return func(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
		if update.Message != nil {
			user := update.Message.From
			text := update.Message.Text
			if len(text) > 50 {
				text = text[:50] + "..."
			}
			if b.debug {
				fmt.Printf("[MSG] User %d (%s %s): %s\n",
					user.ID, user.FirstName, user.LastName, text)
			}
		}
		next(ctx, tb, update)
	}
}

// ChainMiddleware chains multiple middlewares together
func ChainMiddleware(final tgbot.HandlerFunc, middlewares ...Middleware) tgbot.HandlerFunc {
	// Apply middlewares in reverse order
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}
	return final
}
