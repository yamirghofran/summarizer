package content

import (
	"strings"
)

// ContentType represents the type of content source
type ContentType string

const (
	ContentTypeYouTube ContentType = "youtube"
	ContentTypeWebpage ContentType = "webpage"
	ContentTypeMedia   ContentType = "media" // Audio/video content (YouTube, video URLs, local files)
)

// Content represents fetched content ready for summarization
type Content struct {
	Text        string      // The main content text (transcription or article)
	Title       string      // Title of the video/article
	Description string      // Description or summary
	Author      string      // Author (for articles)
	Site        string      // Site name
	Published   string      // Publication date
	WordCount   int         // Word count
	Type        ContentType // Type of content source
}

// IsYouTubeURL checks if a URL is a YouTube video URL
func IsYouTubeURL(url string) bool {
	url = strings.ToLower(url)

	// Check for various YouTube URL patterns
	youtubePatterns := []string{
		"youtube.com/watch",
		"youtube.com/shorts/",
		"youtube.com/embed/",
		"youtube.com/v/",
		"youtu.be/",
		"m.youtube.com/watch",
	}

	for _, pattern := range youtubePatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}

	return false
}

// Fetcher defines the interface for fetching content from a URL
type Fetcher interface {
	Fetch(url string) (*Content, error)
	CheckDependencies() error
}
