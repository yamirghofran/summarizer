package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "summarizer",
	Short: "Summarize YouTube videos and web pages using AI",
	Long: `A CLI tool that summarizes content from YouTube videos and web pages.

For YouTube videos:
  - Downloads audio using yt-dlp
  - Compresses and speeds up audio with ffmpeg
  - Transcribes using parakeet-mlx
  - Generates summary using OpenAI-compatible API

For web pages:
  - Parses content using defuddle
  - Extracts metadata (author, date, etc.)
  - Generates summary using OpenAI-compatible API

Usage:
  summarizer summarize <url> [flags]

Examples:
  # Summarize a YouTube video
  summarizer summarize https://www.youtube.com/watch?v=dQw4w9WgXcQ

  # Summarize a web page
  summarizer summarize https://example.com/article

  # Save summary to file
  summarizer summarize https://youtube.com/watch?v=xyz -o summary.txt

  # Use a different model
  summarizer summarize https://example.com/blog --model gpt-4o`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
	rootCmd.AddCommand(transcribeCmd)
}
