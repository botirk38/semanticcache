package providers_test

import (
	"errors"
	"math"
	"os"
	"testing"

	"github.com/botirk38/semanticcache/providers"
	"github.com/botirk38/semanticcache/providers/openai"
	"github.com/botirk38/semanticcache/types"
)

// Mock provider for testing
type mockEmbeddingProvider struct {
	embedding []float32
	shouldErr bool
}

func (m *mockEmbeddingProvider) EmbedText(text string) ([]float32, error) {
	if m.shouldErr {
		return nil, errors.New("mock error")
	}
	return m.embedding, nil
}

func (m *mockEmbeddingProvider) Close() {}

func TestOpenAIProvider(t *testing.T) {
	// Skip if no API key available
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping OpenAI provider tests")
	}

	config := openai.OpenAIConfig{
		APIKey: apiKey,
		Model:  "", // Use default model
	}

	provider, err := providers.NewOpenAIProvider(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}
	defer provider.Close()

	t.Run("EmbedText", func(t *testing.T) {
		testText := "Hello, world!"
		embedding, err := provider.EmbedText(testText)
		if err != nil {
			t.Fatalf("Failed to embed text: %v", err)
		}

		if len(embedding) == 0 {
			t.Error("Expected non-empty embedding")
		}

		// OpenAI text-embedding-3-small should return 1536 dimensions
		expectedDim := 1536
		if len(embedding) != expectedDim {
			t.Errorf("Expected %d dimensions, got %d", expectedDim, len(embedding))
		}

		// Embeddings should be normalized (roughly unit length)
		var magnitude float32
		for _, val := range embedding {
			magnitude += val * val
		}
		magnitude = float32(math.Sqrt(float64(magnitude)))

		if magnitude < 0.9 || magnitude > 1.1 {
			t.Errorf("Expected embedding magnitude ~1.0, got %f", magnitude)
		}
	})

	t.Run("ConsistentEmbeddings", func(t *testing.T) {
		testText := "Consistent test text"

		embedding1, err1 := provider.EmbedText(testText)
		embedding2, err2 := provider.EmbedText(testText)

		if err1 != nil || err2 != nil {
			t.Fatalf("Failed to embed text: %v, %v", err1, err2)
		}

		if len(embedding1) != len(embedding2) {
			t.Error("Embeddings should have same length")
		}

		// Embeddings should be identical for same input
		for i := range embedding1 {
			if embedding1[i] != embedding2[i] {
				t.Error("Embeddings should be identical for same input")
				break
			}
		}
	})

	t.Run("DifferentTexts", func(t *testing.T) {
		embedding1, err1 := provider.EmbedText("cat")
		embedding2, err2 := provider.EmbedText("dog")

		if err1 != nil || err2 != nil {
			t.Fatalf("Failed to embed text: %v, %v", err1, err2)
		}

		// Calculate cosine similarity
		similarity := cosineSimilarity(embedding1, embedding2)

		// "cat" and "dog" should be somewhat similar (both animals) but not identical
		if similarity > 0.95 {
			t.Error("Expected different embeddings for different words")
		}
		if similarity < 0.3 {
			t.Error("Expected some similarity between related words")
		}
	})
}

func TestOpenAIProviderConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      openai.OpenAIConfig
		expectError bool
	}{
		{
			name:        "empty config",
			config:      openai.OpenAIConfig{},
			expectError: true, // No API key
		},
		{
			name: "invalid API key",
			config: openai.OpenAIConfig{
				APIKey: "invalid-key",
			},
			expectError: false, // Error will occur during API call, not creation
		},
		{
			name: "custom model",
			config: openai.OpenAIConfig{
				APIKey: "test-key",
				Model:  "text-embedding-3-large",
			},
			expectError: false,
		},
		{
			name: "custom base URL",
			config: openai.OpenAIConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.custom.com",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := providers.NewOpenAIProvider(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					if provider != nil {
						provider.Close()
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if provider != nil {
					provider.Close()
				}
			}
		})
	}
}

func TestProviderInterface(t *testing.T) {
	// Test that our mock implements the interface correctly
	var provider types.EmbeddingProvider = &mockEmbeddingProvider{
		embedding: []float32{0.1, 0.2, 0.3},
		shouldErr: false,
	}

	embedding, err := provider.EmbedText("test")
	if err != nil {
		t.Fatalf("Mock provider failed: %v", err)
	}

	if len(embedding) != 3 {
		t.Errorf("Expected 3 dimensions, got %d", len(embedding))
	}

	provider.Close() // Should not panic
}

func TestProviderErrorHandling(t *testing.T) {
	provider := &mockEmbeddingProvider{
		shouldErr: true,
	}

	_, err := provider.EmbedText("test")
	if err == nil {
		t.Error("Expected error from mock provider")
	}
}

// Helper function to calculate cosine similarity
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float32) float32 {
	if x == 0 {
		return 0
	}
	z := x / 2
	for range 8 {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
