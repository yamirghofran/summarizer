package content

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yamirghofran/youtube-summarizer/internal/downloader"
	"github.com/yamirghofran/youtube-summarizer/internal/processor"
	"github.com/yamirghofran/youtube-summarizer/internal/transcriber"
)

// YouTubeFetcher fetches content from YouTube videos
type YouTubeFetcher struct {
	keepAudio bool
}

// NewYouTubeFetcher creates a new YouTube fetcher
func NewYouTubeFetcher(keepAudio bool) *YouTubeFetcher {
	return &YouTubeFetcher{keepAudio: keepAudio}
}

// Fetch downloads a YouTube video, processes it, and transcribes it
func (f *YouTubeFetcher) Fetch(url string) (*Content, error) {
	var originalWAV, compressedWAV string

	// Cleanup function
	defer func() {
		if !f.keepAudio {
			if originalWAV != "" {
				os.Remove(originalWAV)
			}
			if compressedWAV != "" {
				os.Remove(compressedWAV)
			}
			// Try to remove temp directory if empty
			os.Remove(".youtube-summarizer-temp")
		}
	}()

	// Get video title
	title, err := downloader.GetVideoTitle(url)
	if err != nil {
		title = "Unknown"
	}

	// Download audio
	originalWAV, err = downloader.DownloadWAV(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}

	// Compress and speed up audio
	compressedWAV, err = processor.CompressAndSpeedUp(originalWAV)
	if err != nil {
		return nil, fmt.Errorf("failed to compress audio: %w", err)
	}

	// Transcribe
	transcription, err := transcriber.Transcribe(compressedWAV)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe: %w", err)
	}

	// Return content
	return &Content{
		Text:        transcription,
		Title:       title,
		Description: "",
		Author:      "",
		Site:        "YouTube",
		Published:   "",
		WordCount:   countWords(transcription),
		Type:        ContentTypeYouTube,
	}, nil
}

// CheckDependencies checks if all required tools are available
func (f *YouTubeFetcher) CheckDependencies() error {
	if err := downloader.CheckBinary(); err != nil {
		return err
	}
	if err := processor.CheckBinary(); err != nil {
		return err
	}
	if err := transcriber.CheckBinary(); err != nil {
		return err
	}
	return nil
}

// GetAudioPaths returns the paths to audio files if keepAudio is true
func (f *YouTubeFetcher) GetAudioPaths() (original, compressed string) {
	// This would need to be tracked differently if we want to expose paths
	// For now, return empty strings as paths are managed internally
	return "", ""
}

// countWords counts the number of words in a string
func countWords(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Fields(s))
}

// EstimateTranscriptionTime provides a rough estimate of transcription time
func EstimateTranscriptionTime(audioDuration time.Duration) time.Duration {
	// Rough estimate: parakeet-mlx processes at roughly 0.3x realtime after compression
	return time.Duration(float64(audioDuration) * 0.3)
}
