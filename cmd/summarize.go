package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yamirghofran/summarizer/internal/content"
	"github.com/yamirghofran/summarizer/internal/summarizer"
)

var (
	outputFile string
	model      string
	keepAudio  bool
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize <url>",
	Short: "Summarize a YouTube video or web page",
	Long: `Download and summarize content from YouTube videos or web pages.

For YouTube videos:
  1. Downloads the video as audio (WAV format)
  2. Compresses and speeds up the audio for faster transcription
  3. Transcribes using parakeet-mlx
  4. Generates a summary using OpenAI-compatible API

For web pages:
  1. Fetches and parses the page using defuddle
  2. Extracts content and metadata
  3. Generates a summary using OpenAI-compatible API

Required environment variables:
  OPENAI_API_KEY - Your OpenAI API key

Optional environment variables:
  OPENAI_BASE_URL - Custom API endpoint (for OpenAI-compatible APIs)
  OPENAI_MODEL    - Default model to use (can be overridden with --model flag)

Required CLI tools:
  For YouTube: yt-dlp, ffmpeg, parakeet-mlx
  For web pages: defuddle`,
	Args: cobra.ExactArgs(1),
	Run:  runSummarize,
}

func init() {
	summarizeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Save summary to file")
	summarizeCmd.Flags().StringVar(&model, "model", "", "LLM model to use for summarization (default: from .env or gpt-4o-mini)")
	summarizeCmd.Flags().BoolVar(&keepAudio, "keep-audio", false, "Keep downloaded audio files (YouTube only)")
}

func runSummarize(cmd *cobra.Command, args []string) {
	url := args[0]
	ctx := context.Background()

	// Override model if specified via flag (takes precedence over env)
	if model != "" {
		os.Setenv("OPENAI_MODEL", model)
	}

	// Detect content type
	isYouTube := content.IsYouTubeURL(url)

	// Check dependencies based on content type
	fmt.Println("🔍 Checking dependencies...")
	if isYouTube {
		fmt.Println("   Detected: YouTube video")
		fetcher := content.NewYouTubeFetcher(keepAudio)
		if err := fetcher.CheckDependencies(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("   Detected: Web page")
		fetcher := content.NewWebpageFetcher()
		if err := fetcher.CheckDependencies(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println()

	// Fetch content
	var cont *content.Content
	var err error
	var start time.Time

	if isYouTube {
		// YouTube flow
		fmt.Println("📺 Fetching video information...")
		fetcher := content.NewYouTubeFetcher(keepAudio)

		// Get title first for progress display
		start = time.Now()
		cont, err = fetcher.Fetch(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("   ✓ Fetched and transcribed in %s\n\n", time.Since(start).Round(time.Second))
	} else {
		// Web page flow
		fmt.Println("🌐 Fetching web page...")
		fetcher := content.NewWebpageFetcher()

		start = time.Now()
		cont, err = fetcher.Fetch(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("   ✓ Fetched in %s\n\n", time.Since(start).Round(time.Second))
	}

	// Display content info
	displayContentInfo(cont)

	// Summarize
	fmt.Println("✨ Generating summary...")
	start = time.Now()

	summarizerInstance, err := summarizer.New()
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

	// Print to console
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("SUMMARY")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(output)

	// Save to file if specified
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n✓ Summary saved to: %s\n", outputFile)
	}
}

func displayContentInfo(cont *content.Content) {
	fmt.Println("📄 Content Info:")
	fmt.Printf("   Title: %s\n", truncate(cont.Title, 60))
	if cont.Author != "" {
		fmt.Printf("   Author: %s\n", cont.Author)
	}
	if cont.Site != "" {
		fmt.Printf("   Site: %s\n", cont.Site)
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
