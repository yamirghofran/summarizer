package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yamirghofran/summarizer/internal/config"
	"github.com/yamirghofran/summarizer/internal/content"
	"github.com/yamirghofran/summarizer/internal/downloader"
	"github.com/yamirghofran/summarizer/internal/summarizer"
)

var (
	outputFile string
	model      string
	keepAudio  bool
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize <url-or-file>",
	Short: "Summarize a YouTube video, web page, or media file",
	Long: `Download and summarize content from various sources.

For YouTube videos:
  1. Downloads the video as audio (WAV format)
  2. Compresses and speeds up the audio for faster transcription
  3. Transcribes using parakeet-mlx
  4. Generates a summary using OpenAI-compatible API

For video/audio URLs:
  1. Downloads the media file
  2. Converts to audio if needed
  3. Transcribes using parakeet-mlx
  4. Generates a summary

For local video/audio files:
  1. Converts to audio if needed
  2. Transcribes using parakeet-mlx
  3. Generates a summary

For web pages:
  1. Fetches and parses the page using defuddle
  2. Extracts content and metadata
  3. Generates a summary using OpenAI-compatible API

Configuration files:
  ~/.config/summarizer/config.toml
  ~/.local/share/summarizer/credentials.toml

Initialize defaults with:
  summarizer config init

Required CLI tools:
  For media: yt-dlp, ffmpeg, parakeet-mlx
  For web pages: defuddle`,
	Args: cobra.ExactArgs(1),
	Run:  runSummarize,
}

func init() {
	summarizeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Save summary to file")
	summarizeCmd.Flags().StringVar(&model, "model", "", "LLM model to use for summarization (overrides configured model)")
	summarizeCmd.Flags().BoolVar(&keepAudio, "keep-audio", false, "Keep downloaded/converted audio files")
}

func runSummarize(cmd *cobra.Command, args []string) {
	input := args[0]
	ctx := context.Background()

	loaded, err := config.Load("", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	aiSettings, err := loaded.AISettings(model)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Detect content type and determine fetcher
	var fetcher content.Fetcher
	var contentType string

	isURL := downloader.IsURL(input)
	isYouTube := content.IsYouTubeURL(input)
	isMedia := content.IsMediaInput(input)

	if isYouTube {
		contentType = "YouTube video"
		fetcher = content.NewMediaFetcher(keepAudio)
	} else if isURL && content.IsMediaURL(input) {
		contentType = "video/audio URL"
		fetcher = content.NewMediaFetcher(keepAudio)
	} else if isURL {
		contentType = "web page"
		fetcher = content.NewWebpageFetcher()
	} else if isMedia {
		contentType = "local media file"
		fetcher = content.NewMediaFetcher(keepAudio)
	} else {
		fmt.Fprintf(os.Stderr, "Error: unsupported input type. Provide a URL or media file.\n")
		os.Exit(1)
	}

	// Check dependencies based on content type
	fmt.Println("🔍 Checking dependencies...")
	fmt.Printf("   Detected: %s\n", contentType)
	if err := fetcher.CheckDependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// Fetch content
	var cont *content.Content
	var start time.Time

	switch contentType {
	case "YouTube video":
		fmt.Println("📺 Fetching video information...")
	case "video/audio URL":
		fmt.Println("📥 Downloading media...")
	case "local media file":
		fmt.Println("📝 Transcribing media file...")
	case "web page":
		fmt.Println("🌐 Fetching web page...")
	}

	start = time.Now()
	cont, err = fetcher.Fetch(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if contentType == "web page" {
		fmt.Printf("   ✓ Fetched in %s\n\n", time.Since(start).Round(time.Second))
	} else {
		fmt.Printf("   ✓ Fetched and transcribed in %s\n\n", time.Since(start).Round(time.Second))
	}

	// Display content info
	displayContentInfo(cont)

	// Summarize
	fmt.Println("✨ Generating summary...")
	start = time.Now()

	summarizerInstance, err := summarizer.New(summarizer.Settings{
		APIKey:  aiSettings.APIKey,
		BaseURL: aiSettings.BaseURL,
		Model:   aiSettings.Model,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	summary, err := summarizerInstance.Summarize(ctx, cont)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating summary: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   ✓ Generated in %s\n\n", time.Since(start).Round(time.Second))

	// Output results
	output := formatOutput(cont, summary, summarizerInstance.GetModel())

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n✓ Summary saved to: %s\n", outputFile)
		return
	}

	// Print to console
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("SUMMARY")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(output)
}

func displayContentInfo(cont *content.Content) {
	fmt.Println("📄 Content Info:")
	fmt.Printf("   Title: %s\n", truncate(cont.Title, 60))
	if cont.Author != "" {
		fmt.Printf("   Author: %s\n", cont.Author)
	}
	if cont.Site != "" {
		fmt.Printf("   Source: %s\n", cont.Site)
	}
	if cont.Published != "" {
		fmt.Printf("   Published: %s\n", cont.Published)
	}
	if cont.WordCount > 0 {
		fmt.Printf("   Word count: ~%d\n", cont.WordCount)
	}
	fmt.Println()
}

func formatOutput(cont *content.Content, summary, model string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Title: %s\n", cont.Title))

	if cont.Author != "" {
		sb.WriteString(fmt.Sprintf("Author: %s\n", cont.Author))
	}
	if cont.Site != "" {
		sb.WriteString(fmt.Sprintf("Source: %s\n", cont.Site))
	}
	if cont.Published != "" {
		sb.WriteString(fmt.Sprintf("Published: %s\n", cont.Published))
	}

	sb.WriteString(fmt.Sprintf("Model: %s\n", model))
	sb.WriteString("\n")
	sb.WriteString(summary)

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
