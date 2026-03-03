package content

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// DefuddleResponse represents the JSON response from defuddle CLI
type DefuddleResponse struct {
	Content       string `json:"content"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Author        string `json:"author"`
	Site          string `json:"site"`
	Domain        string `json:"domain"`
	Favicon       string `json:"favicon"`
	Image         string `json:"image"`
	Published     string `json:"published"`
	WordCount     int    `json:"wordCount"`
	ParseTime     int    `json:"parseTime"`
	SchemaOrgData []any  `json:"schemaOrgData"`
}

// WebpageFetcher fetches content from web pages using defuddle
type WebpageFetcher struct{}

// NewWebpageFetcher creates a new webpage fetcher
func NewWebpageFetcher() *WebpageFetcher {
	return &WebpageFetcher{}
}

// Fetch fetches and parses a web page using defuddle
func (f *WebpageFetcher) Fetch(url string) (*Content, error) {
	// Run defuddle with markdown and JSON output
	cmd := exec.Command("defuddle", "parse", url, "--markdown", "--json")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("defuddle failed: %w\nStderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("defuddle failed: %w", err)
	}

	// Find the start of JSON (defuddle may output info messages before the JSON)
	jsonStart := strings.Index(string(output), "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("no JSON found in defuddle output")
	}
	jsonOutput := output[jsonStart:]

	// Parse JSON response
	var resp DefuddleResponse
	if err := json.Unmarshal(jsonOutput, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse defuddle response: %w", err)
	}

	// Use site name if available, otherwise use domain
	site := resp.Site
	if site == "" {
		site = resp.Domain
	}

	// Return content
	return &Content{
		Text:        resp.Content,
		Title:       resp.Title,
		Description: resp.Description,
		Author:      resp.Author,
		Site:        site,
		Published:   resp.Published,
		WordCount:   resp.WordCount,
		Type:        ContentTypeWebpage,
	}, nil
}

// CheckDependencies checks if defuddle is available
func (f *WebpageFetcher) CheckDependencies() error {
	_, err := exec.LookPath("defuddle")
	if err != nil {
		return fmt.Errorf("defuddle not found in PATH. Install it to summarize web pages")
	}
	return nil
}

// CheckDefuddleBinary is a standalone function to check for defuddle
func CheckDefuddleBinary() error {
	_, err := exec.LookPath("defuddle")
	if err != nil {
		return fmt.Errorf("defuddle not found in PATH. Install it to summarize web pages")
	}
	return nil
}

// ParseLocalFile parses a local HTML file using defuddle
func (f *WebpageFetcher) ParseLocalFile(filePath string) (*Content, error) {
	cmd := exec.Command("defuddle", "parse", filePath, "--markdown", "--json")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("defuddle failed: %w\nStderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("defuddle failed: %w", err)
	}

	// Find the start of JSON (defuddle may output info messages before the JSON)
	jsonStart := strings.Index(string(output), "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("no JSON found in defuddle output")
	}
	jsonOutput := output[jsonStart:]

	var resp DefuddleResponse
	if err := json.Unmarshal(jsonOutput, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse defuddle response: %w", err)
	}

	site := resp.Site
	if site == "" {
		site = resp.Domain
	}

	return &Content{
		Text:        resp.Content,
		Title:       resp.Title,
		Description: resp.Description,
		Author:      resp.Author,
		Site:        site,
		Published:   resp.Published,
		WordCount:   resp.WordCount,
		Type:        ContentTypeWebpage,
	}, nil
}

// ExtractProperty extracts a specific property from a web page
func ExtractProperty(url, property string) (string, error) {
	cmd := exec.Command("defuddle", "parse", url, "--property", property)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("defuddle failed: %w\nStderr: %s", err, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("defuddle failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
