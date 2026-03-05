# YouTube & Web Summarizer

A Go CLI tool and Telegram bot that summarizes YouTube videos and web pages using AI.

## Features

- **YouTube Videos**: Downloads audio, transcribes with parakeet-mlx, and summarizes
- **Audio/Video Files**: Transcribe local audio or video files directly
- **Web Pages**: Extracts content with defuddle and summarizes
- **Auto-detection**: Automatically detects YouTube vs web URLs
- **OpenAI Compatible**: Works with OpenAI, Ollama, LM Studio, Together AI, and other compatible APIs
- **Metadata Rich**: Includes author, site, and publication date in summaries
- **Telegram Bot**: Interactive bot for easy summarization via Telegram

## Prerequisites

### Required
- **Go 1.23+**
- **OpenAI API key** (or compatible API)

### For YouTube Videos
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - Download YouTube videos
- [ffmpeg](https://ffmpeg.org) - Audio processing
- [parakeet-mlx](https://github.com/anthropics/claude-code) - Audio transcription

### For Audio/Video Files (Transcribe command)
- [ffmpeg](https://ffmpeg.org) - Video to audio conversion & compression
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - Download from URLs
- [parakeet-mlx](https://github.com/anthropics/claude-code) - Audio transcription

### For Web Pages
- [defuddle](https://github.com/silverbulletmd/defuddle) - Web content extraction

### For Telegram Bot
- Telegram bot token from [@BotFather](https://t.me/BotFather)

## Installation

```bash
# Clone the repository
git clone https://github.com/yamirghofran/summarizer.git
cd summarizer

# Install dependencies and build
go mod tidy
go build -o summarizer .
```

## Configuration

1. Initialize config files:
```bash
./summarizer config init
```

2. Edit `~/.config/summarizer/config.toml`:
```toml
default_provider = "openai"

[providers.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-4o-mini"

[telegram]
allowed_user_ids = [123456789, 987654321]
```

3. Edit `~/.local/share/summarizer/credentials.toml`:
```toml
[providers.openai]
api_key = "your-api-key-here"

[telegram]
bot_token = "your-bot-token-here"
```

Migration note:
- `summarize` and `bot` now load from TOML files only.
- Env vars like `OPENAI_API_KEY` and `TELEGRAM_BOT_TOKEN` are no longer used.
- `transcribe` is unchanged.

`config init` options:
```bash
./summarizer config init --force
./summarizer config init --config-path /custom/config.toml --credentials-path /custom/credentials.toml
```

## Usage

### CLI Mode

#### Basic Usage

```bash
# Summarize a YouTube video
./summarizer summarize "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Summarize a web page
./summarizer summarize "https://example.com/article"

# Short YouTube URLs work too
./summarizer summarize "https://youtu.be/dQw4w9WgXcQ"
```

#### Save to File

```bash
./summarizer summarize "https://youtube.com/..." -o summary.txt
```

#### Use a Different Model

```bash
./summarizer summarize "https://youtube.com/..." --model gpt-4o
```

#### Keep Audio Files (YouTube only)

```bash
./summarizer summarize "https://youtube.com/..." --keep-audio
```

#### Transcribe Audio/Video

```bash
# Transcribe a local audio file
./summarizer transcribe "audio.mp3"

# Transcribe a local video file
./summarizer transcribe "video.mp4"

# Transcribe a YouTube video
./summarizer transcribe "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Transcribe from a direct URL
./summarizer transcribe "https://example.com/audio.mp3"

# Save transcription to file
./summarizer transcribe "audio.mp3" -o transcription.txt

# Keep intermediate audio files
./summarizer transcribe "video.mp4" --keep-audio
```

The transcribe command supports:
- **Local audio files**: mp3, wav, m4a, flac, ogg, aac, wma
- **Local video files**: mp4, mkv, avi, mov, wmv, flv, webm, m4v
- **YouTube URLs**: Downloads and transcribes
- **Direct URLs**: Downloads and transcribes

#### CLI Flags

##### Summarize Command

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Save summary to file |
| `--model` | | LLM model to use (overrides configured model) |
| `--keep-audio` | | Keep downloaded audio files (YouTube only) |

##### Transcribe Command

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Save transcription to file |
| `--keep-audio` | | Keep intermediate audio files (converted from video/downloaded) |

### Telegram Bot Mode

Start the Telegram bot to summarize content via chat:

```bash
# Start the bot
./summarizer bot

# Start with debug logging
./summarizer bot --debug

# Start with token override
./summarizer bot --token "your-bot-token"
```

#### Bot Commands

- `/start` - Welcome message and introduction
- `/help` - Show usage instructions
- `/status` - Check bot status

#### Bot Features

- **URL Detection**: Automatically detects YouTube and web page URLs
- **Progress Updates**: Shows real-time status (downloading, transcribing, etc.)
- **Markdown Formatting**: Beautiful formatted summaries with bold text and structure
- **Access Control**: Restrict to specific Telegram user IDs via `telegram.allowed_user_ids`
- **Long Messages**: Automatically splits long summaries into multiple messages

#### Setting Up the Bot

1. **Create a Telegram Bot**:
   - Message [@BotFather](https://t.me/BotFather) on Telegram
   - Send `/newbot` and follow the instructions
   - Copy the bot token

2. **Get Your User ID**:
   - Message [@userinfobot](https://t.me/userinfobot) on Telegram
   - Note your user ID

3. **Configure the Bot**:
   - Set `telegram.bot_token` in `~/.local/share/summarizer/credentials.toml`
   - Set `telegram.allowed_user_ids` in `~/.config/summarizer/config.toml`

4. **Start the Bot**:
   ```bash
   ./summarizer bot
   ```

5. **Use the Bot**:
   - Open your bot on Telegram
   - Send `/start` to begin
   - Send any YouTube or web page URL to get a summary

## How It Works

### YouTube Pipeline
1. **Download** - Uses `yt-dlp` to extract audio as WAV
2. **Compress** - Uses `ffmpeg` to convert to 16kHz mono at 1.7x speed
3. **Transcribe** - Uses `parakeet-mlx` for speech-to-text
4. **Summarize** - Uses OpenAI-compatible API to generate summary
5. **Cleanup** - Removes temp files (unless `--keep-audio`)

### Transcribe Pipeline
1. **Download** - Uses `yt-dlp` for YouTube URLs, HTTP for direct URLs (skipped for local files)
2. **Convert** - Uses `ffmpeg` to extract audio from video files (skipped for audio files)
3. **Compress** - Uses `ffmpeg` to convert to 16kHz mono at 1.7x speed
4. **Transcribe** - Uses `parakeet-mlx` for speech-to-text
5. **Output** - Prints to stdout or saves to file
6. **Cleanup** - Removes temp files (unless `--keep-audio`)

### Web Page Pipeline
1. **Fetch** - Uses `defuddle` to extract content and metadata
2. **Summarize** - Uses OpenAI-compatible API to generate summary

## Using with Local Models

Works with Ollama, LM Studio, and other local LLM servers:

```toml
default_provider = "ollama"

[providers.ollama]
base_url = "http://localhost:11434/v1"
model = "llama3.2"
```

## Example Output

### CLI Output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Title: Understanding Go Concurrency
Author: Jane Doe
Source: techblog.example.com
Published: 2024-01-15
Model: gpt-4o-mini

This video covers the fundamentals of concurrency in Go...

Key points:
- Goroutines are lightweight threads managed by the Go runtime
- Channels enable safe communication between goroutines
- The select statement handles multiple channel operations
```

### Telegram Bot Output

The bot sends beautifully formatted markdown messages:

```
📄 Understanding Go Concurrency

👤 Author: Jane Doe
🌐 Source: techblog.example.com
📅 Published: 2024-01-15
🤖 Model: gpt-4o-mini

━━━━━━━━━━━━━━━━━━━━

Overview
This video covers the fundamentals...

Key Points
• Goroutines are lightweight threads...
• Channels enable safe communication...
• The select statement handles multiple...

Takeaways
• Start with goroutines for simple concurrency
• Use channels for goroutine communication
```

## Project Structure

```
summarizer/
├── main.go                 # Entry point
├── cmd/
│   ├── root.go            # Root command
│   ├── config.go          # Config command group
│   ├── config_init.go     # Config initialization command
│   ├── summarize.go       # Summarize command
│   ├── transcribe.go      # Transcribe command
│   └── bot.go             # Telegram bot command
├── internal/
│   ├── bot/
│   │   ├── bot.go         # Bot initialization & polling
│   │   ├── handlers.go    # Message handlers
│   │   └── middleware.go  # Allowlist & logging
│   ├── content/
│   │   ├── source.go      # Content interface & URL detection
│   │   ├── youtube.go     # YouTube fetcher
│   │   └── webpage.go     # Web page fetcher
│   ├── downloader/
│   │   ├── youtube.go     # yt-dlp wrapper
│   │   └── url.go         # URL downloader
│   ├── processor/
│   │   ├── audio.go       # ffmpeg wrapper
│   │   └── video.go       # Video to audio conversion
│   ├── transcriber/
│   │   └── parakeet.go    # parakeet-mlx wrapper
│   ├── summarizer/
│   │   └── openai.go      # OpenAI API client
│   ├── config/
│   │   ├── config.go      # TOML config loading and resolution
│   │   └── config_test.go # Config unit tests
│   └── urlutil/
│       └── detect.go      # URL extraction & detection
├── .gitignore
├── README.md
└── WEBHOOK_MIGRATION.md   # Guide for webhook deployment
```

## Webhook Deployment

The bot supports webhook mode for production deployments. See [WEBHOOK_MIGRATION.md](WEBHOOK_MIGRATION.md) for instructions on:
- Switching from polling to webhooks
- Deploying to cloud platforms
- Setting up HTTPS and SSL

## Development

### Building

```bash
go build -o summarizer .
```

### Running Tests

```bash
go test ./...
```

## License

MIT
