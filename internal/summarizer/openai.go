package summarizer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/yamirghofran/summarizer/internal/content"
)

const (
	defaultModel = "gpt-4o-mini"
)

// Summarizer handles summarizing content using OpenAI-compatible API
type Summarizer struct {
	client *openai.Client
	model  string
}

// New creates a new Summarizer instance
func New() (*Summarizer, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Get configuration from environment
	baseURL := os.Getenv("OPENAI_BASE_URL")
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = defaultModel
	}

	// Create OpenAI client configuration
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	return &Summarizer{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}, nil
}

// Summarize generates a summary of the content
func (s *Summarizer) Summarize(ctx context.Context, cont *content.Content) (string, error) {
	if cont.Text == "" {
		return "", fmt.Errorf("content text is empty")
	}

	// Build system prompt based on content type
	systemPrompt := buildSystemPrompt(cont)

	// Build user prompt with context
	userPrompt := buildUserPrompt(cont)

	// Create the chat completion request
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature:         0.7,
		MaxCompletionTokens: 10000,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	return resp.Choices[0].Message.Content, nil
}

// buildSystemPrompt creates the system prompt based on content type
func buildSystemPrompt(cont *content.Content) string {
	formattingInstructions := `
Format your response using Telegram Markdown:
- Use *asterisks* for bold text (e.g., *Important Point*)
- Use _underscores_ for italic text (e.g., _emphasis_)
- Use inline code for technical terms (e.g., ` + "`code`" + `)
- Use bullet points with dashes (e.g., - Item)
- Do NOT use ## or ** for formatting
- Do NOT use headers (#) - use bold for section titles instead

Structure your summary clearly with bold section titles.`

	switch cont.Type {
	case content.ContentTypeYouTube:
		return `You are a helpful assistant that creates concise summaries of video content.

Please summarize the video transcription in a clear and organized way. Include:
1. A brief overview of the main topic
2. Key points and insights discussed
3. Any important conclusions or takeaways

Keep the summary concise but informative.` + formattingInstructions

	case content.ContentTypeWebpage:
		return `You are a helpful assistant that creates concise summaries of articles and web content.

Please summarize the content in a clear and organized way. Include:
1. A brief overview of the main topic
2. Key points and insights
3. Any important conclusions or actionable takeaways

Keep the summary concise but informative.` + formattingInstructions

	default:
		return `You are a helpful assistant that creates concise summaries.

Please summarize the content in a clear and organized way. Include:
1. A brief overview of the main topic
2. Key points and insights
3. Any important conclusions

Keep the summary concise but informative.` + formattingInstructions
	}
}

// buildUserPrompt creates the user prompt with content and context
func buildUserPrompt(cont *content.Content) string {
	var sb strings.Builder

	// Add context about the content
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

	sb.WriteString("\n")

	// Add the appropriate prompt based on content type
	switch cont.Type {
	case content.ContentTypeYouTube:
		sb.WriteString("Please summarize this video transcription:\n\n")
	case content.ContentTypeWebpage:
		sb.WriteString("Please summarize this article:\n\n")
	default:
		sb.WriteString("Please summarize this content:\n\n")
	}

	sb.WriteString(cont.Text)

	return sb.String()
}

// GetModel returns the model being used
func (s *Summarizer) GetModel() string {
	return s.model
}
