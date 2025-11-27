package tokenizer

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// GeminiTokenizer counts tokens for Gemini contents
type GeminiTokenizer struct {
	client *genai.Client
	model  string
}

// NewGeminiTokenizer creates a new GeminiTokenizer with the provided client and model
func NewGeminiTokenizer(client *genai.Client, model string) *GeminiTokenizer {
	return &GeminiTokenizer{
		client: client,
		model:  model,
	}
}

// CountTokens counts tokens in Gemini contents using the native Gemini SDK
// This makes an API call to Gemini's token counting endpoint
func (t *GeminiTokenizer) CountTokens(ctx context.Context, contents []*genai.Content) (int, error) {
	if len(contents) == 0 {
		return 0, nil
	}

	// Client and model are required for Gemini token counting
	if t.client == nil {
		return 0, fmt.Errorf("gemini client is required for token counting")
	}

	if t.model == "" {
		return 0, fmt.Errorf("gemini model is required for token counting")
	}

	// Use native Gemini API for accurate token counting
	result, err := t.client.Models.CountTokens(ctx, t.model, contents, nil)
	if err != nil {
		return 0, fmt.Errorf("gemini token counting failed: %w", err)
	}

	return int(result.TotalTokens), nil
}
