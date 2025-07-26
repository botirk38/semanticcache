package providers

import (
	"github.com/botirk38/semanticcache/providers/openai"
	"github.com/botirk38/semanticcache/types"
)

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config openai.OpenAIConfig) (types.EmbeddingProvider, error) {
	return openai.NewOpenAIProvider(config)
}
