package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yamirghofran/summarizer/internal/urlutil"
)

// IsURL checks if the input string is a URL
func IsURL(input string) bool {
	return strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://")
}

// DownloadURL downloads content from a URL
// For YouTube URLs, it uses yt-dlp to download audio
// For other URLs, it downloads the file directly via HTTP
func DownloadURL(urlStr string) (string, error) {
	// Check if it's a YouTube URL
	if urlutil.IsYouTubeURL(urlStr) {
		// Use existing YouTube downloader
		return DownloadWAV(urlStr)
	}

	// For non-YouTube URLs, download directly via HTTP
	return downloadHTTP(urlStr)
}

// downloadHTTP downloads a file from a direct URL using HTTP
func downloadHTTP(urlStr string) (string, error) {
	// Parse URL to extract filename
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Create HTTP request
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to download from URL: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	// Determine filename
	filename := getFilenameFromURL(parsedURL, resp)

	// Create temp directory if it doesn't exist
	tempDir := ".summarizer-temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create output file
	outputPath := filepath.Join(tempDir, filename)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Copy the response body to file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("failed to save downloaded file: %w", err)
	}

	return outputPath, nil
}

// getFilenameFromURL extracts a filename from URL path or Content-Disposition header
func getFilenameFromURL(parsedURL *url.URL, resp *http.Response) string {
	// Try to get filename from Content-Disposition header
	if contentDisposition := resp.Header.Get("Content-Disposition"); contentDisposition != "" {
		// Parse Content-Disposition header for filename
		// Format: attachment; filename="file.mp4" or attachment; filename=file.mp4
		if parts := strings.Split(contentDisposition, "filename="); len(parts) > 1 {
			filename := strings.TrimSpace(parts[1])
			// Remove quotes if present
			filename = strings.Trim(filename, `"`)
			if filename != "" {
				return filename
			}
		}
	}

	// Fall back to URL path
	path := parsedURL.Path
	if path != "" && path != "/" {
		// Get the last segment of the path
		segments := strings.Split(path, "/")
		if len(segments) > 0 {
			filename := segments[len(segments)-1]
			if filename != "" {
				return filename
			}
		}
	}

	// Ultimate fallback
	return "downloaded-file"
}
