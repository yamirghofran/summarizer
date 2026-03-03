# YouTube & Web Summarizer

A Go CLI tool that summarizes YouTube videos and web pages using AI.

## Features

- **YouTube Videos**: Downloads audio, transcribes with parakeet-mlx, and summarizes
- **Web Pages**: Extracts content with defuddle and summarizes
- **Auto-detection**: Automatically detects YouTube vs web URLs
- **OpenAI Compatible**: Works with OpenAI, Ollama, LM Studio, Together AI, and other compatible APIs
- **Metadata Rich**: Includes author, site, and publication date in summaries

## Prerequisites

### Required
- **Go 1.23+**
- **OpenAI API key** (or compatible API)

### For YouTube Videos
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - Download YouTube videos
- [ffmpeg](https://ffmpeg.org) - Audio processing
- [parakeet-mlx](https://github.com/anthropics/claude-code) - Audio transcription

### For Web Pages
- [defuddle](https://github.com/silverbulletmd/defuddle) - Web content extraction

## Installation

```bash
# Clone the repository
git clone https://github.com/yamirghofran/youtube-summarizer.git
cd youtube-summarizer

# Install dependencies and build
go mod tidy
go build -o youtube-summarizer .
```

## Configuration

1. Create your environment file:
```bash
cp .env.example .env
```

2. Edit `.env` with your settings:
```env
# Required
OPENAI_API_KEY=your-api-key-here

# Optional - use a custom API endpoint
OPENAI_BASE_URL=http://localhost:11434/v1

# Optional - default model
OPENAI_MODEL=gpt-4o-mini
```

## Usage

### Basic Usage

```bash
# Summarize a YouTube video
./youtube-summarizer summarize "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Summarize a web page
./youtube-summarizer summarize "https://example.com/article"

# Short YouTube URLs work too
./youtube-summarizer summarize "https://youtu.be/dQw4w9WgXcQ"
```

### Save to File

```bash
./youtube-summarizer summarize "https://youtube.com/..." -o summary.txt
```

### Use a Different Model

```bash
./youtube-summarizer summarize "https://youtube.com/..." --model gpt-4o
```

### Keep Audio Files (YouTube only)

```bash
./youtube-summarizer summarize "https://youtube.com/..." --keep-audio
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Save summary to file |
| `--model` | | LLM model to use (overrides env) |
| `--keep-audio` | | Keep downloaded audio files (YouTube only) |

## How It Works

### YouTube Pipeline
1. **Download** - Uses `yt-dlp` to extract audio as WAV
2. **Compress** - Uses `ffmpeg` to convert to 16kHz mono at 1.7x speed
3. **Transcribe** - Uses `parakeet-mlx` for speech-to-text
4. **Summarize** - Uses OpenAI-compatible API to generate summary
5. **Cleanup** - Removes temp files (unless `--keep-audio`)

### Web Page Pipeline
1. **Fetch** - Uses `defuddle` to extract content and metadata
2. **Summarize** - Uses OpenAI-compatible API to generate summary

## Using with Local Models

Works with Ollama, LM Studio, and other local LLM servers:

```env
OPENAI_BASE_URL=http://localhost:11434/v1
OPENAI_MODEL=llama3.2
```

## Example Output

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

## Project Structure

```
youtube-summarizer/
├── main.go                 # Entry point
├── cmd/
│   ├── root.go            # Root command
│   └── summarize.go       # Summarize command
├── internal/
│   ├── content/
│   │   ├── source.go      # Content interface & URL detection
│   │   ├── youtube.go     # YouTube fetcher
│   │   └── webpage.go     # Web page fetcher
│   ├── downloader/
│   │   └── youtube.go     # yt-dlp wrapper
│   ├── processor/
│   │   └── audio.go       # ffmpeg wrapper
│   ├── transcriber/
│   │   └── parakeet.go    # parakeet-mlx wrapper
│   └── summarizer/
│       └── openai.go      # OpenAI API client
├── .env.example           # Environment template
└── .gitignore
```

## License

MIT
