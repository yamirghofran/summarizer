# Migrating from Polling to Webhooks

This guide explains how to switch the Telegram bot from polling mode to webhook mode.

## Why Use Webhooks?

**Polling (default):**
- Simpler to set up
- No public URL required
- Bot actively requests updates from Telegram
- Good for development and testing

**Webhooks:**
- More efficient - Telegram pushes updates to your server
- Better for production deployments
- Required for some hosting platforms
- Lower latency for message delivery

## Prerequisites

1. A public HTTPS URL (Telegram requires HTTPS)
2. A server accessible from the internet
3. SSL certificate (Let's Encrypt is free)

## Migration Steps

### 1. Update Your Hosting Setup

You need a server with a public URL. Options include:

- **VPS** (DigitalOcean, Linode, etc.) with nginx/caddy
- **Cloud Run** (Google Cloud)
- **Lambda + API Gateway** (AWS)
- **Railway**, **Render**, **Fly.io**
- **Ngrok** for development/testing

### 2. Modify the Bot Code

The bot is designed to easily switch between modes. Here's how:

#### Current (Polling Mode)

```go
// cmd/bot.go - current implementation
func runBot(cmd *cobra.Command, args []string) {
    cfg := &bot.Config{
        Token:        token,
        AllowedUsers: allowedUsers,
    }
    
    b, _ := bot.New(cfg)
    b.Start(ctx)  // Polling mode
}
```

#### Webhook Mode

Create a new command or add a flag:

```go
// Add to cmd/bot.go
var webhookURL string
var webhookPort string

func init() {
    botCmd.Flags().StringVar(&webhookURL, "webhook-url", "", "Webhook URL (enables webhook mode)")
    botCmd.Flags().StringVar(&webhookPort, "webhook-port", "8080", "Port for webhook server")
}

func runBot(cmd *cobra.Command, args []string) {
    cfg := &bot.Config{
        Token:        token,
        AllowedUsers: allowedUsers,
    }
    
    b, _ := bot.New(cfg)
    
    if webhookURL != "" {
        // Webhook mode
        b.StartWebhook(ctx, ":"+webhookPort, webhookURL)
    } else {
        // Polling mode (default)
        b.Start(ctx)
    }
}
```

### 3. Update internal/bot/bot.go

The webhook method is already stubbed in `internal/bot/bot.go`. Here's the full implementation:

```go
// StartWebhook begins webhook mode
func (b *Bot) StartWebhook(ctx context.Context, addr string, webhookURL string) error {
    // Register handlers
    b.registerHandlers()
    
    // Set webhook with Telegram
    _, err := b.bot.SetWebhook(ctx, &tgbot.SetWebhookParams{
        URL: webhookURL,
    })
    if err != nil {
        return fmt.Errorf("failed to set webhook: %w", err)
    }
    
    fmt.Printf("🤖 Bot started in webhook mode\n")
    fmt.Printf("📡 Webhook URL: %s\n", webhookURL)
    fmt.Printf("🌐 Listening on %s\n", addr)
    
    // Start webhook receiver in background
    go b.bot.StartWebhook(ctx)
    
    // Start HTTP server
    http.Handle("/", b.bot.WebhookHandler())
    return http.ListenAndServe(addr, nil)
}
```

### 4. Run with Webhooks

```bash
# With a public URL like https://mybot.example.com
./summarizer bot --webhook-url https://mybot.example.com --webhook-port 8080
```

### 5. Remove Webhook (if needed)

To switch back to polling, you need to delete the webhook:

```bash
# Using curl
curl "https://api.telegram.org/bot<YOUR_TOKEN>/deleteWebhook"

# Or add a delete-webhook command to the CLI
```

## Development with Ngrok

For testing webhooks locally:

```bash
# 1. Start ngrok
ngrok http 8080

# 2. Copy the HTTPS URL (e.g., https://abc123.ngrok.io)

# 3. Start the bot with webhook
./summarizer bot --webhook-url https://abc123.ngrok.io
```

## Docker Deployment Example

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o summarizer .

FROM alpine:latest
RUN apk --no-cache add ca-certificates ffmpeg
COPY --from=builder /app/summarizer /usr/local/bin/
COPY .env /.env

EXPOSE 8080
CMD ["summarizer", "bot", "--webhook-url", "https://your-domain.com", "--webhook-port", "8080"]
```

## Security Considerations

1. **Verify requests**: Telegram sends a secret token in the `X-Telegram-Bot-Api-Secret-Token` header. Set it with:

   ```go
   opts = append(opts, tgbot.WithWebhookSecretToken(os.Getenv("TELEGRAM_WEBHOOK_SECRET")))
   ```

2. **HTTPS only**: Telegram only sends webhooks to HTTPS URLs

3. **Rate limiting**: Consider adding rate limiting to your webhook endpoint

## Troubleshooting

### Webhook not receiving updates

1. Check if webhook is set:
   ```bash
   curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo"
   ```

2. Check your server logs for incoming requests

3. Verify your SSL certificate is valid

### Switching back to polling

1. Delete the webhook:
   ```bash
   curl "https://api.telegram.org/bot<TOKEN>/deleteWebhook"
   ```

2. Restart the bot without the webhook flag

## Summary

| Aspect | Polling | Webhook |
|--------|---------|---------|
| Setup | Simple | Requires public URL |
| Efficiency | Lower | Higher |
| Latency | Higher | Lower |
| Best for | Development | Production |
| Dependencies | None | HTTPS server |
