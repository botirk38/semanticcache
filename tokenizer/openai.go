package tokenizer

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go/v2"
	"github.com/tiktoken-go/tokenizer"
)

// OpenAITokenizer counts tokens for OpenAI messages using tiktoken
type OpenAITokenizer struct{}

// NewOpenAITokenizer creates a new OpenAITokenizer
func NewOpenAITokenizer() *OpenAITokenizer {
	return &OpenAITokenizer{}
}

// CountTokens counts tokens in OpenAI chat completion messages using tiktoken
// This is a local, fast operation that doesn't require an API call
func (t *OpenAITokenizer) CountTokens(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (int, error) {
	if len(messages) == 0 {
		return 0, nil
	}

	// Serialize data to JSON
	jsonBytes, err := json.Marshal(messages)
	if err != nil {
		return 0, err
	}

	// Use tiktoken Cl100kBase encoding for accurate token counting
	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		return 0, err
	}

	// Count tokens
	ids, _, _ := enc.Encode(string(jsonBytes))
	return len(ids), nil
}
