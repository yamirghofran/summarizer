package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yamirghofran/summarizer/internal/content"
)

var (
	transcribeOutputFile string
	transcribeKeepAudio  bool
)

var transcribeCmd = &cobra.Command{
	Use:   "transcribe <file-or-url>",
	Short: "Transcribe audio or video to text",
	Long: `Transcribe audio or video files to text using parakeet-mlx.

Accepts:
  - Local audio files (mp3, wav, m4a, flac, ogg, aac, wma)
  - Local video files (mp4, mkv, avi, mov, wmv, flv, webm)
  - YouTube URLs (downloads and transcribes)
  - Direct video/audio URLs (downloads and transcribes)

Process:
  1. Downloads if URL provided
  2. Converts video to audio if needed
  3. Compresses and speeds up audio (16kHz mono, 1.7x)
  4. Transcribes using parakeet-mlx
  5. Outputs to stdout or file

Required CLI tools:
  For video conversion: ffmpeg
  For YouTube URLs: yt-dlp
  For transcription: parakeet-mlx`,
	Args: cobra.ExactArgs(1),
	Run:  runTranscribe,
}

func init() {
	transcribeCmd.Flags().StringVarP(&transcribeOutputFile, "output", "o", "", "Save transcription to file")
	transcribeCmd.Flags().BoolVar(&transcribeKeepAudio, "keep-audio", false, "Keep intermediate audio file (converted from video/downloaded)")
}

func runTranscribe(cmd *cobra.Command, args []string) {
	input := args[0]

	// 1. Check dependencies
	fmt.Println("🔍 Checking dependencies...")
	fetcher := content.NewMediaFetcher(transcribeKeepAudio)
	if err := fetcher.CheckDependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// 2. Fetch and transcribe
	var start time.Time
	if content.IsMediaInput(input) {
		if content.IsMediaURL(input) {
			fmt.Println("📥 Downloading and transcribing...")
		} else if content.IsYouTubeURL(input) {
			fmt.Println("📺 Downloading YouTube video and transcribing...")
		} else {
			fmt.Println("📝 Transcribing...")
		}
		start = time.Now()

		cont, err := fetcher.Fetch(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("   ✓ Transcribed in %s\n\n", time.Since(start).Round(time.Second))

		// 3. Output results
		output := formatTranscriptionOutput(input, cont.Text, cont.Title)

		// Print to console
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("TRANSCRIPTION")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println(output)

		// Save to file if specified
		if transcribeOutputFile != "" {
			if err := os.WriteFile(transcribeOutputFile, []byte(output), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n✓ Transcription saved to: %s\n", transcribeOutputFile)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: unsupported input. Provide a video/audio file or URL.\n")
		os.Exit(1)
	}
}

func formatTranscriptionOutput(source, transcription, title string) string {
	var sb strings.Builder

	if title != "" {
		sb.WriteString(fmt.Sprintf("Title: %s\n", title))
	}
	sb.WriteString(fmt.Sprintf("Source: %s\n", source))
	sb.WriteString("\n")
	sb.WriteString(transcription)

	return sb.String()
}
