package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yamirghofran/summarizer/internal/downloader"
	"github.com/yamirghofran/summarizer/internal/processor"
	"github.com/yamirghofran/summarizer/internal/transcriber"
	"github.com/yamirghofran/summarizer/internal/urlutil"
)

// MediaFetcher fetches content from media sources (YouTube URLs, video/audio URLs, local files)
type MediaFetcher struct {
	keepAudio bool
}

// NewMediaFetcher creates a new media fetcher
func NewMediaFetcher(keepAudio bool) *MediaFetcher {
	return &MediaFetcher{keepAudio: keepAudio}
}

// Fetch downloads/processes media and transcribes it to text
// Accepts:
//   - YouTube URLs
//   - Direct video/audio URLs
//   - Local video files
//   - Local audio files
func (f *MediaFetcher) Fetch(input string) (*Content, error) {
	var downloadedFile string // Downloaded from URL
	var convertedFile string  // Video converted to audio
	var compressedFile string // Compressed for transcription

	// Cleanup function - always removes compressed, optionally removes others
	defer func() {
		// Always remove compressed audio
		if compressedFile != "" {
			os.Remove(compressedFile)
		}

		// Remove intermediate files unless keepAudio
		if !f.keepAudio {
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

	// 1. Get input file (download if URL, or use local path)
	var inputFile string
	var title string

	if downloader.IsURL(input) {
		// Download from URL
		file, err := downloader.DownloadURL(input)
		if err != nil {
			return nil, fmt.Errorf("failed to download: %w", err)
		}
		downloadedFile = file
		inputFile = file

		// Try to get title from YouTube if it's a YouTube URL
		if urlutil.IsYouTubeURL(input) {
			title, _ = downloader.GetVideoTitle(input)
		}
	} else {
		// Local file
		if _, err := os.Stat(input); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", input)
		}
		inputFile = input
	}

	// 2. Determine title if not already set
	if title == "" {
		title = extractTitleFromPath(inputFile)
	}

	// 3. Convert video to audio if needed
	if processor.IsVideoFile(inputFile) {
		wavFile, err := processor.ConvertToAudio(inputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to convert video to audio: %w", err)
		}
		convertedFile = wavFile
		inputFile = wavFile
	} else if !processor.IsAudioFile(inputFile) {
		return nil, fmt.Errorf("unsupported file format. Supported formats:\n  Audio: mp3, wav, m4a, flac, ogg, aac, wma\n  Video: mp4, mkv, avi, mov, wmv, flv, webm, m4v")
	}

	// 4. Compress and speed up audio
	compressedWAV, err := processor.CompressAndSpeedUp(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to compress audio: %w", err)
	}
	compressedFile = compressedWAV

	// 5. Transcribe
	transcription, err := transcriber.Transcribe(compressedWAV)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe: %w", err)
	}

	// Determine site name
	site := "Local File"
	if downloader.IsURL(input) {
		if urlutil.IsYouTubeURL(input) {
			site = "YouTube"
		} else {
			site = "Video/Audio URL"
		}
	}

	// Return content
	return &Content{
		Text:        transcription,
		Title:       title,
		Description: "",
		Author:      "",
		Site:        site,
		Published:   "",
		WordCount:   countWords(transcription),
		Type:        ContentTypeMedia,
	}, nil
}

// CheckDependencies checks if all required tools are available
func (f *MediaFetcher) CheckDependencies() error {
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

// extractTitleFromPath extracts a title from a file path
func extractTitleFromPath(path string) string {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// countWords counts the number of words in a string
func countWords(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Fields(s))
}

// IsMediaInput checks if an input is likely a media file or URL
// Returns true for:
//   - Local video/audio files
//   - YouTube URLs
//   - URLs with video/audio extensions
func IsMediaInput(input string) bool {
	// Check for YouTube URL
	if urlutil.IsYouTubeURL(input) {
		return true
	}

	// Check for local media file
	if !downloader.IsURL(input) {
		return processor.IsVideoFile(input) || processor.IsAudioFile(input)
	}

	// Check for URL with media extension
	return IsMediaURL(input)
}

// IsMediaURL checks if a URL points to a video or audio file based on extension
func IsMediaURL(urlStr string) bool {
	// Parse the URL path and check extension
	lower := strings.ToLower(urlStr)

	videoExts := []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v"}
	audioExts := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac", ".wma"}

	for _, ext := range videoExts {
		if strings.Contains(lower, ext+"?") || strings.HasSuffix(lower, ext) {
			return true
		}
	}
	for _, ext := range audioExts {
		if strings.Contains(lower, ext+"?") || strings.HasSuffix(lower, ext) {
			return true
		}
	}

	return false
}
