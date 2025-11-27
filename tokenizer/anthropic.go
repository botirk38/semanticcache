package tokenizer

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// AnthropicTokenizer counts tokens for Anthropic messages
type AnthropicTokenizer struct {
	client *anthropic.Client
}

// NewAnthropicTokenizer creates a new AnthropicTokenizer with the provided client
func NewAnthropicTokenizer(client *anthropic.Client) *AnthropicTokenizer {
	return &AnthropicTokenizer{
		client: client,
	}
}

// CountTokens counts tokens in Anthropic messages using the native Anthropic SDK
// This makes an API call to Anthropic's token counting endpoint
func (t *AnthropicTokenizer) CountTokens(ctx context.Context, messages []anthropic.MessageParam) (int, error) {
	if len(messages) == 0 {
		return 0, nil
	}

	// Client is required for Anthropic token counting
	if t.client == nil {
		return 0, fmt.Errorf("anthropic client is required for token counting")
	}

	// Use native Anthropic API for accurate token counting
	params := anthropic.MessageCountTokensParams{
		Model:    "claude-3-5-sonnet-20241022", // Default model for counting
		Messages: messages,
	}

	result, err := t.client.Messages.CountTokens(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("anthropic token counting failed: %w", err)
	}

	return int(result.InputTokens), nil
}
