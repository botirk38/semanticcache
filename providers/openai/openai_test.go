package openai

import (
	"testing"

	openai "github.com/openai/openai-go/v2"
)

func TestOpenAIProvider_GetMaxTokens(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected int
	}{
		{
			name:     "text-embedding-3-small",
			model:    openai.EmbeddingModelTextEmbedding3Small,
			expected: 8191,
		},
		{
			name:     "text-embedding-3-large",
			model:    openai.EmbeddingModelTextEmbedding3Large,
			expected: 8191,
		},
		{
			name:     "text-embedding-ada-002",
			model:    openai.EmbeddingModelTextEmbeddingAda002,
			expected: 8191,
		},
		{
			name:     "unknown model",
			model:    "unknown-model",
			expected: 8191, // Should return safe default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &OpenAIProvider{
				model: tt.model,
			}

			maxTokens := provider.GetMaxTokens()
			if maxTokens != tt.expected {
				t.Errorf("GetMaxTokens() = %d, want %d for model %s", maxTokens, tt.expected, tt.model)
			}
		})
	}
}

func TestOpenAIModelLimits(t *testing.T) {
	// Test that all expected models are in the limits map
	expectedModels := []string{
		openai.EmbeddingModelTextEmbedding3Small,
		openai.EmbeddingModelTextEmbedding3Large,
		openai.EmbeddingModelTextEmbeddingAda002,
	}

	for _, model := range expectedModels {
		if limit, exists := openAIModelLimits[model]; !exists {
			t.Errorf("model %s not found in openAIModelLimits", model)
		} else if limit != 8191 {
			t.Errorf("model %s has unexpected limit %d, expected 8191", model, limit)
		}
	}
}
