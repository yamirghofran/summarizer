package downloader

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// DownloadWAV downloads a YouTube video as a WAV file using yt-dlp
// Returns the path to the downloaded WAV file
func DownloadWAV(url string) (string, error) {
	// Create temp directory for downloads
	tempDir := ".youtube-summarizer-temp"

	// Build the yt-dlp command
	// Use a simple output template to get the title
	cmd := exec.Command("yt-dlp",
		"-x",
		"--audio-format", "wav",
		"--audio-quality", "0",
		"-o", filepath.Join(tempDir, "%(title)s.%(ext)s"),
		"--no-playlist",
		url,
	)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w\nOutput: %s", err, string(output))
	}

	// Find the downloaded file
	files, err := filepath.Glob(filepath.Join(tempDir, "*.wav"))
	if err != nil {
		return "", fmt.Errorf("failed to find downloaded file: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no WAV file found after download")
	}

	// Return the first (and should be only) WAV file
	return files[0], nil
}

// CheckBinary verifies that yt-dlp is available in PATH
func CheckBinary() error {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return fmt.Errorf("yt-dlp not found in PATH. Please install it first: https://github.com/yt-dlp/yt-dlp")
	}
	return nil
}

// GetVideoTitle fetches the video title without downloading
func GetVideoTitle(url string) (string, error) {
	cmd := exec.Command("yt-dlp", "--get-title", "--no-playlist", url)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get video title: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
