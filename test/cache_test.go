package semanticcache_test

import (
	"testing"

	"github.com/botirk38/semanticcache/semanticcache"
)

// Mock provider always returns the given "embedding" for a text.
type mockProvider struct {
	embedding []float32
}

func (m *mockProvider) EmbedText(text string) ([]float32, error) {
	return m.embedding, nil
}
func (m *mockProvider) Close() {}

func TestLookup(t *testing.T) {
	embedding := []float32{0.1, 0.2}
	provider := &mockProvider{embedding: embedding}
	cache, err := semanticcache.NewSemanticCache[string, string](10, provider, nil)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set uses inputText, but mock always returns embedding.
	if err := cache.Set("foo", "bar-text", "bar-value"); err != nil {
		t.Fatalf("Failed to set cache entry: %v", err)
	}

	val, ok, err := cache.Lookup("bar-text", 0.9)
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if !ok || val != "bar-value" {
		t.Errorf("Expected bar-value, got %v", val)
	}
}
