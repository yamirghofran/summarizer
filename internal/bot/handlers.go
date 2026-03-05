package bot

import (
	"context"
	"fmt"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/yamirghofran/summarizer/internal/content"
	"github.com/yamirghofran/summarizer/internal/summarizer"
	"github.com/yamirghofran/summarizer/internal/urlutil"
)

const (
	maxMessageLength = 4096 // Telegram's max message length
)

// handleStart handles the /start command
func (b *Bot) handleStart(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	if !b.IsAllowed(userID) {
		b.sendUnauthorizedMessage(ctx, tb, update.Message.Chat.ID)
		return
	}

	welcomeMsg := `👋 Welcome to the YouTube & Web Summarizer Bot\!

Send me a link and I'll summarize it for you:

• 📺 *YouTube videos* \- I'll download, transcribe, and summarize
• 🎬 *Video/audio URLs* \- I'll download, transcribe, and summarize \(\.mp4, \.mp3, etc\.\)
• 🌐 *Web pages* \- I'll extract and summarize the content

Just paste a URL and I'll handle the rest\!

Use /help for more information\.`

	tb.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      welcomeMsg,
		ParseMode: models.ParseModeMarkdown,
	})
}

// handleHelp handles the /help command
func (b *Bot) handleHelp(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	if !b.IsAllowed(userID) {
		b.sendUnauthorizedMessage(ctx, tb, update.Message.Chat.ID)
		return
	}

	helpMsg := `🤖 *How to use this bot*

*Summarize content:*
Just send me a URL\! I support:

• YouTube videos \(youtube\.com, youtu\.be\)
• Video/audio file URLs \(\.mp4, \.mp3, etc\.\)
• Any web page or article

*Commands:*
/start \- Start the bot
/help \- Show this help message
/status \- Check bot status

*Examples:*
https://youtube\.com/watch?v=\.\.\.
https://example\.com/video\.mp4
https://example\.com/article

*Tips:*
• YouTube videos and media URLs take longer \(download \+ transcribe\)
• Web pages are processed quickly
• You'll see progress updates while processing`

	tb.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      helpMsg,
		ParseMode: models.ParseModeMarkdown,
	})
}

// handleStatus handles the /status command
func (b *Bot) handleStatus(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	if !b.IsAllowed(userID) {
		b.sendUnauthorizedMessage(ctx, tb, update.Message.Chat.ID)
		return
	}

	statusMsg := `✅ Bot is running normally

*Services:*
• YouTube processing: Available
• Video/audio URL processing: Available
• Web page processing: Available
• AI Summarization: Available`

	tb.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      statusMsg,
		ParseMode: models.ParseModeMarkdown,
	})
}

// HandleMessage handles incoming text messages
func (b *Bot) HandleMessage(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	// Check if user is allowed
	if !b.IsAllowed(userID) {
		b.sendUnauthorizedMessage(ctx, tb, chatID)
		return
	}

	// Extract URL from message
	url, found := urlutil.ExtractURL(text)
	if !found {
		// No URL found, send help
		b.sendHelpMessage(ctx, tb, chatID)
		return
	}

	// Check for multiple URLs
	if urlutil.HasMultipleURLs(text) {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   "⚠️ I found multiple URLs. I'll process the first one: " + url,
		})
	}

	// Process the URL
	b.processURL(ctx, tb, chatID, url)
}

// processURL handles the full URL processing flow
func (b *Bot) processURL(ctx context.Context, tb *tgbot.Bot, chatID int64, url string) {
	// Send initial status message
	statusMsg, _ := tb.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   "⏳ Processing your request...",
	})

	statusChatID := statusMsg.Chat.ID
	statusMsgID := statusMsg.ID

	// Helper to update status
	updateStatus := func(text string) {
		tb.EditMessageText(ctx, &tgbot.EditMessageTextParams{
			ChatID:    statusChatID,
			MessageID: statusMsgID,
			Text:      text,
		})
	}

	// Detect content type
	isYouTube := urlutil.IsYouTubeURL(url)
	isMediaURL := content.IsMediaURL(url)

	if isYouTube || isMediaURL {
		b.processMediaURL(ctx, tb, chatID, url, updateStatus)
	} else {
		b.processWebURL(ctx, tb, chatID, url, updateStatus)
	}

	// Delete status message when done
	tb.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
		ChatID:    statusChatID,
		MessageID: statusMsgID,
	})
}

// processMediaURL handles media (YouTube, video/audio URLs) summarization
func (b *Bot) processMediaURL(ctx context.Context, tb *tgbot.Bot, chatID int64, url string, updateStatus func(string)) {
	// Determine content type for display
	contentType := "video/audio URL"
	if urlutil.IsYouTubeURL(url) {
		contentType = "YouTube video"
	}

	updateStatus(fmt.Sprintf("⏳ Processing your request...\n\n📺 Detected: %s", contentType))

	// Check dependencies
	updateStatus(fmt.Sprintf("⏳ Processing your request...\n\n🔍 Checking dependencies..."))
	fetcher := content.NewMediaFetcher(false)
	if err := fetcher.CheckDependencies(); err != nil {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error: %v", err),
		})
		return
	}

	// Fetch content (download, compress, transcribe)
	updateStatus(fmt.Sprintf("⏳ Processing your request...\n\n⬇️ Downloading and transcribing..."))
	cont, err := fetcher.Fetch(url)
	if err != nil {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error processing media: %v", err),
		})
		return
	}

	// Summarize
	updateStatus(fmt.Sprintf("⏳ Processing your request...\n\n✨ Generating summary...\n\n📹 %s", truncate(cont.Title, 50)))
	b.summarizeAndSend(ctx, tb, chatID, cont)
}

// processWebURL handles web page summarization
func (b *Bot) processWebURL(ctx context.Context, tb *tgbot.Bot, chatID int64, url string, updateStatus func(string)) {
	updateStatus("⏳ Processing your request...\n\n🌐 Detected: Web page")

	// Check dependencies
	updateStatus("⏳ Processing your request...\n\n🔍 Checking dependencies...")
	fetcher := content.NewWebpageFetcher()
	if err := fetcher.CheckDependencies(); err != nil {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error: %v", err),
		})
		return
	}

	// Fetch content
	updateStatus("⏳ Processing your request...\n\n🌐 Fetching web page...")
	cont, err := fetcher.Fetch(url)
	if err != nil {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error fetching page: %v", err),
		})
		return
	}

	// Summarize
	updateStatus(fmt.Sprintf("⏳ Processing your request...\n\n✨ Generating summary...\n\n📄 Page: %s", truncate(cont.Title, 50)))
	b.summarizeAndSend(ctx, tb, chatID, cont)
}

// summarizeAndSend generates a summary and sends it to the chat
func (b *Bot) summarizeAndSend(ctx context.Context, tb *tgbot.Bot, chatID int64, cont *content.Content) {
	// Create summarizer
	summarizerInstance, err := summarizer.New(summarizer.Settings{
		APIKey:  b.ai.APIKey,
		BaseURL: b.ai.BaseURL,
		Model:   b.ai.Model,
	})
	if err != nil {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error: %v", err),
		})
		return
	}

	// Generate summary
	summary, err := summarizerInstance.Summarize(ctx, cont)
	if err != nil {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Error generating summary: %v", err),
		})
		return
	}

	// Format and send summary
	b.sendSummaryMessage(ctx, tb, chatID, cont, summary, summarizerInstance.GetModel())
}

// sendSummaryMessage sends the formatted summary to the chat
func (b *Bot) sendSummaryMessage(ctx context.Context, tb *tgbot.Bot, chatID int64, cont *content.Content, summary, model string) {
	// Build header with markdown formatting (strict escape for metadata)
	var header strings.Builder
	header.WriteString(fmt.Sprintf("📄 *%s*\n\n", escapeMarkdownV2Strict(cont.Title)))

	if cont.Author != "" {
		header.WriteString(fmt.Sprintf("👤 Author: %s\n", escapeMarkdownV2Strict(cont.Author)))
	}
	if cont.Site != "" {
		header.WriteString(fmt.Sprintf("🌐 Source: %s\n", escapeMarkdownV2Strict(cont.Site)))
	}
	if cont.Published != "" {
		header.WriteString(fmt.Sprintf("📅 Published: %s\n", escapeMarkdownV2Strict(cont.Published)))
	}
	header.WriteString(fmt.Sprintf("🤖 Model: %s\n", escapeMarkdownV2Strict(model)))
	header.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n\n")

	headerText := header.String()

	// Escape summary while preserving markdown formatting (*, _, `)
	escapedSummary := escapeMarkdownV2(summary)

	// Check if summary fits in one message
	fullMessage := headerText + escapedSummary

	if len(fullMessage) <= maxMessageLength {
		// Send as single message with markdown
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      fullMessage,
			ParseMode: models.ParseModeMarkdown,
		})
	} else {
		// Send header first
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      headerText + "_Summary continues below\\._",
			ParseMode: models.ParseModeMarkdown,
		})

		// Split summary into chunks with markdown
		b.sendLongMessageMarkdown(ctx, tb, chatID, escapedSummary)
	}
}

// sendLongMessageMarkdown sends a long message with markdown formatting
func (b *Bot) sendLongMessageMarkdown(ctx context.Context, tb *tgbot.Bot, chatID int64, text string) {
	// Split by paragraphs first
	paragraphs := strings.Split(text, "\n\n")

	var currentChunk strings.Builder

	for _, para := range paragraphs {
		// Check if adding this paragraph would exceed the limit
		if currentChunk.Len()+len(para)+2 > maxMessageLength-100 {
			// Send current chunk
			if currentChunk.Len() > 0 {
				tb.SendMessage(ctx, &tgbot.SendMessageParams{
					ChatID:    chatID,
					Text:      currentChunk.String(),
					ParseMode: models.ParseModeMarkdown,
				})
				currentChunk.Reset()
			}
		}

		// If single paragraph is too long, split by sentences
		if len(para) > maxMessageLength-100 {
			b.sendChunkedMessageMarkdown(ctx, tb, chatID, para)
			continue
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
	}

	// Send remaining chunk
	if currentChunk.Len() > 0 {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      currentChunk.String(),
			ParseMode: models.ParseModeMarkdown,
		})
	}
}

// sendLongMessage sends a long message as plain text (no markdown parsing)
func (b *Bot) sendLongMessage(ctx context.Context, tb *tgbot.Bot, chatID int64, text string) {
	// Split by paragraphs first
	paragraphs := strings.Split(text, "\n\n")

	var currentChunk strings.Builder

	for _, para := range paragraphs {
		// Check if adding this paragraph would exceed the limit
		if currentChunk.Len()+len(para)+2 > maxMessageLength-100 {
			// Send current chunk
			if currentChunk.Len() > 0 {
				tb.SendMessage(ctx, &tgbot.SendMessageParams{
					ChatID: chatID,
					Text:   currentChunk.String(),
				})
				currentChunk.Reset()
			}
		}

		// If single paragraph is too long, split by sentences
		if len(para) > maxMessageLength-100 {
			b.sendChunkedMessage(ctx, tb, chatID, para)
			continue
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
	}

	// Send remaining chunk
	if currentChunk.Len() > 0 {
		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   currentChunk.String(),
		})
	}
}

// sendChunkedMessageMarkdown sends a very long text with markdown
func (b *Bot) sendChunkedMessageMarkdown(ctx context.Context, tb *tgbot.Bot, chatID int64, text string) {
	chunkSize := maxMessageLength - 100

	for len(text) > 0 {
		chunk := text
		if len(chunk) > chunkSize {
			chunk = text[:chunkSize]
			// Try to find a good break point
			if idx := strings.LastIndexAny(chunk, ".!?\n"); idx > chunkSize/2 {
				chunk = text[:idx+1]
			}
		}

		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      chunk,
			ParseMode: models.ParseModeMarkdown,
		})

		text = text[len(chunk):]
	}
}

// sendChunkedMessage sends a very long text by splitting into fixed-size chunks
func (b *Bot) sendChunkedMessage(ctx context.Context, tb *tgbot.Bot, chatID int64, text string) {
	chunkSize := maxMessageLength - 100

	for len(text) > 0 {
		chunk := text
		if len(chunk) > chunkSize {
			chunk = text[:chunkSize]
			// Try to find a good break point
			if idx := strings.LastIndexAny(chunk, ".!?\n"); idx > chunkSize/2 {
				chunk = text[:idx+1]
			}
		}

		tb.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   chunk,
		})

		text = text[len(chunk):]
	}
}

// sendHelpMessage sends a help message when no URL is found
func (b *Bot) sendHelpMessage(ctx context.Context, tb *tgbot.Bot, chatID int64) {
	msg := `👋 Send me a link to summarize!

• 📺 YouTube videos
• 🎬 Video/audio file URLs (.mp4, .mp3, etc.)
• 🌐 Web pages

Just paste a URL and I'll handle the rest.

Use /help for more information.`

	tb.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   msg,
	})
}

// sendUnauthorizedMessage sends a message to unauthorized users
func (b *Bot) sendUnauthorizedMessage(ctx context.Context, tb *tgbot.Bot, chatID int64) {
	tb.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   "⛔ Sorry, you're not authorized to use this bot.",
	})
}

// escapeMarkdownV2 escapes special characters for Telegram MarkdownV2 format
// Preserves markdown formatting characters (*, _, `) while escaping others
// In MarkdownV2, these characters must be escaped: _ * [ ] ( ) ~ ` > # + - = | { } . !
// But we keep * _ ` unescaped to allow markdown formatting
func escapeMarkdownV2(text string) string {
	var result strings.Builder
	result.Grow(len(text) + len(text)/10) // Pre-allocate extra space for escapes

	for _, char := range text {
		switch char {
		// Keep markdown formatting characters unescaped
		case '*', '_', '`':
			result.WriteRune(char)
		// Escape all other special characters
		case '[', ']', '(', ')', '~', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!':
			result.WriteRune('\\')
			result.WriteRune(char)
		default:
			result.WriteRune(char)
		}
	}

	return result.String()
}

// escapeMarkdownV2Strict escapes ALL special characters (for non-LLM content like titles)
func escapeMarkdownV2Strict(text string) string {
	var result strings.Builder
	result.Grow(len(text) + len(text)/10) // Pre-allocate extra space for escapes

	for _, char := range text {
		switch char {
		case '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!':
			result.WriteRune('\\')
			result.WriteRune(char)
		default:
			result.WriteRune(char)
		}
	}

	return result.String()
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
