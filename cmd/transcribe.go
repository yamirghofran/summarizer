package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yamirghofran/summarizer/internal/downloader"
	"github.com/yamirghofran/summarizer/internal/processor"
	"github.com/yamirghofran/summarizer/internal/transcriber"
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

	// Track files for cleanup
	var downloadedFile string // Downloaded from URL
	var convertedFile string  // Video converted to audio
	var compressedFile string // Compressed for transcription

	// Cleanup function - always removes compressed, optionally removes others
	defer func() {
		// Always remove compressed audio
		if compressedFile != "" {
			os.Remove(compressedFile)
		}

		// Remove intermediate files unless --keep-audio
		if !transcribeKeepAudio {
			if downloadedFile != "" {
				os.Remove(downloadedFile)
			}
			if convertedFile != "" {
				os.Remove(convertedFile)
			}
			// Try to remove temp directory if empty
			os.Remove(".summarizer-temp")
		}
	}()

	// 1. Check dependencies
	fmt.Println("🔍 Checking dependencies...")
	checkTranscribeDependencies()

	// 2. Get input file
	var inputFile string
	if downloader.IsURL(input) {
		fmt.Println("📥 Downloading from URL...")
		start := time.Now()

		file, err := downloader.DownloadURL(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		downloadedFile = file
		inputFile = file

		fmt.Printf("   ✓ Downloaded in %s\n\n", time.Since(start).Round(time.Second))
	} else {
		// Local file
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", input)
			os.Exit(1)
		}
		inputFile = input
	}

	// 3. Convert video to audio if needed
	if processor.IsVideoFile(inputFile) {
		fmt.Println("🎬 Converting video to audio...")
		start := time.Now()

		wavFile, err := processor.ConvertToAudio(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		convertedFile = wavFile
		inputFile = wavFile

		fmt.Printf("   ✓ Converted in %s\n\n", time.Since(start).Round(time.Second))
	} else if !processor.IsAudioFile(inputFile) {
		fmt.Fprintf(os.Stderr, "Error: unsupported file format. Supported formats:\n")
		fmt.Fprintf(os.Stderr, "  Audio: mp3, wav, m4a, flac, ogg, aac, wma\n")
		fmt.Fprintf(os.Stderr, "  Video: mp4, mkv, avi, mov, wmv, flv, webm, m4v\n")
		os.Exit(1)
	}

	// 4. Compress and speed up
	fmt.Println("🎵 Compressing audio (16kHz mono, 1.7x speed)...")
	start := time.Now()

	compressedWAV, err := processor.CompressAndSpeedUp(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	compressedFile = compressedWAV
	fmt.Printf("   ✓ Compressed in %s\n\n", time.Since(start).Round(time.Second))

	// 5. Transcribe
	fmt.Println("📝 Transcribing...")
	start = time.Now()

	transcription, err := transcriber.Transcribe(compressedFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   ✓ Transcribed in %s\n\n", time.Since(start).Round(time.Second))

	// 6. Output results
	output := formatTranscriptionOutput(input, transcription)

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

	// Report kept files
	if transcribeKeepAudio && (downloadedFile != "" || convertedFile != "") {
		fmt.Println("\n📎 Kept audio files:")
		if downloadedFile != "" && convertedFile == "" {
			fmt.Printf("   %s\n", downloadedFile)
		} else if convertedFile != "" {
			fmt.Printf("   %s\n", convertedFile)
		}
	}
}

func checkTranscribeDependencies() {
	// Check ffmpeg (for video conversion and compression)
	if err := processor.CheckBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check parakeet-mlx (for transcription)
	if err := transcriber.CheckBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check yt-dlp (for YouTube downloads)
	if err := downloader.CheckBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()
}

func formatTranscriptionOutput(source, transcription string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Source: %s\n", source))
	sb.WriteString("\n")
	sb.WriteString(transcription)

	return sb.String()
}
