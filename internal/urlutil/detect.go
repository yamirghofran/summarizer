package urlutil

import (
	"net/url"
	"regexp"
	"strings"
)

// urlRegex matches most common URL patterns
var urlRegex = regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `[\]]+`)

// ExtractURL extracts the first URL from a message text
// Returns the URL and true if found, empty string and false otherwise
func ExtractURL(text string) (string, bool) {
	matches := urlRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return "", false
	}

	// Return the first URL found
	return matches[0], true
}

// IsURL checks if text is a valid URL
func IsURL(text string) bool {
	u, err := url.Parse(text)
	if err != nil {
		return false
	}

	// Check that it has a scheme and host
	return u.Scheme != "" && u.Host != ""
}

// NormalizeURL cleans up a URL by removing tracking parameters and fragments
func NormalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove common tracking parameters
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "ref", "source", "_ga",
	}

	q := u.Query()
	for _, param := range trackingParams {
		q.Del(param)
	}
	u.RawQuery = q.Encode()

	// Remove fragment
	u.Fragment = ""

	return u.String()
}

// GetDomain extracts the domain from a URL
func GetDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Host
}

// HasMultipleURLs checks if the text contains multiple URLs
func HasMultipleURLs(text string) bool {
	matches := urlRegex.FindAllString(text, -1)
	return len(matches) > 1
}

// ExtractAllURLs extracts all URLs from a message text
func ExtractAllURLs(text string) []string {
	matches := urlRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	// Deduplicate and normalize
	seen := make(map[string]bool)
	var result []string
	for _, match := range matches {
		normalized := NormalizeURL(match)
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}

// IsYouTubeURL checks if a URL is a YouTube video URL
func IsYouTubeURL(rawURL string) bool {
	u, err := url.Parse(strings.ToLower(rawURL))
	if err != nil {
		return false
	}

	// Check host
	hosts := []string{"youtube.com", "www.youtube.com", "m.youtube.com", "youtu.be", "www.youtu.be"}
	hostMatch := false
	for _, host := range hosts {
		if u.Host == host || strings.HasSuffix(u.Host, "."+host) {
			hostMatch = true
			break
		}
	}

	if !hostMatch {
		return false
	}

	// Check path patterns
	if strings.Contains(u.Host, "youtu.be") {
		return u.Path != "" && u.Path != "/"
	}

	// Check for watch, shorts, embed, etc.
	pathPatterns := []string{"/watch", "/shorts/", "/embed/", "/v/"}
	for _, pattern := range pathPatterns {
		if strings.HasPrefix(u.Path, pattern) {
			return true
		}
	}

	return false
}
