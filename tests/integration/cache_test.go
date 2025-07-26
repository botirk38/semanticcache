package integration_test

import (
	"testing"

	"github.com/botirk38/semanticcache"
	"github.com/botirk38/semanticcache/backends"
	"github.com/botirk38/semanticcache/types"
)

// Mock provider for integration tests
type mockProvider struct {
	embeddings map[string][]float32
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		embeddings: map[string][]float32{
			"cat":    {0.8, 0.2, 0.1},
			"dog":    {0.7, 0.3, 0.2},
			"animal": {0.6, 0.3, 0.3},
			"car":    {0.1, 0.8, 0.4},
			"truck":  {0.2, 0.7, 0.5},
			"hello":  {0.3, 0.1, 0.9},
			"hi":     {0.4, 0.2, 0.8},
		},
	}
}

func (m *mockProvider) EmbedText(text string) ([]float32, error) {
	if embedding, exists := m.embeddings[text]; exists {
		return embedding, nil
	}
	// Return default embedding for unknown text
	return []float32{0.5, 0.5, 0.5}, nil
}

func (m *mockProvider) Close() {}

func TestSemanticCacheIntegration(t *testing.T) {
	backends_types := []struct {
		name        string
		backendType types.BackendType
	}{
		{"LRU", types.BackendLRU},
		{"FIFO", types.BackendFIFO},
		{"LFU", types.BackendLFU},
	}

	for _, bt := range backends_types {
		t.Run(bt.name, func(t *testing.T) {
			config := types.BackendConfig{Capacity: 5}
			factory := &backends.BackendFactory[string, string]{}
			backend, err := factory.NewBackend(bt.backendType, config)
			if err != nil {
				t.Fatalf("Failed to create %s backend: %v", bt.name, err)
			}

			provider := newMockProvider()
			cache, err := semanticcache.NewSemanticCache(backend, provider, nil)
			if err != nil {
				t.Fatalf("Failed to create semantic cache: %v", err)
			}
			defer cache.Close()

			testSemanticOperations(t, cache)
			testSemanticSearch(t, cache)
			testCacheEviction(t, cache)
		})
	}
}

func testSemanticOperations(t *testing.T, cache *semanticcache.SemanticCache[string, string]) {
	// Test basic set and get
	err := cache.Set("animal1", "cat", "A feline animal")
	if err != nil {
		t.Fatalf("Failed to set cache entry: %v", err)
	}

	value, found := cache.Get("animal1")
	if !found {
		t.Error("Expected to find animal1")
	}
	if value != "A feline animal" {
		t.Errorf("Expected 'A feline animal', got %s", value)
	}

	// Test contains
	if !cache.Contains("animal1") {
		t.Error("Expected animal1 to exist")
	}

	// Test length
	if cache.Len() != 1 {
		t.Errorf("Expected length 1, got %d", cache.Len())
	}

	// Test delete
	cache.Delete("animal1")
	if cache.Contains("animal1") {
		t.Error("Expected animal1 to be deleted")
	}
}

func testSemanticSearch(t *testing.T, cache *semanticcache.SemanticCache[string, string]) {
	// Add test data
	testData := map[string]string{
		"pet1":     "cat",
		"pet2":     "dog",
		"vehicle1": "car",
		"vehicle2": "truck",
		"greeting": "hello",
	}

	for key, text := range testData {
		err := cache.Set(key, text, "Value for "+text)
		if err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}

	t.Run("ExactMatch", func(t *testing.T) {
		// Test lookup with exact match (should have similarity = 1.0)
		value, found, err := cache.Lookup("cat", 0.99)
		if err != nil {
			t.Fatalf("Lookup failed: %v", err)
		}
		if !found {
			t.Error("Expected to find exact match for 'cat'")
		}
		if value != "Value for cat" {
			t.Errorf("Expected 'Value for cat', got %s", value)
		}
	})

	t.Run("SemanticMatch", func(t *testing.T) {
		// Test lookup with semantic similarity (should find "hi" when searching for similar greeting)
		_ = cache.Set("greeting2", "hi", "Value for hi")

		value, found, err := cache.Lookup("hello", 0.7)
		if err != nil {
			t.Fatalf("Semantic lookup failed: %v", err)
		}
		if found {
			// Should find either "hello" or "hi" depending on similarity
			if value != "Value for hello" && value != "Value for hi" {
				t.Errorf("Unexpected semantic match: %s", value)
			}
		}
	})

	t.Run("TopMatches", func(t *testing.T) {
		// Test top matches
		matches, err := cache.TopMatches("animal", 3)
		if err != nil {
			t.Fatalf("TopMatches failed: %v", err)
		}

		if len(matches) == 0 {
			t.Error("Expected at least one match")
		}

		// Matches should be sorted by similarity (descending)
		for i := 1; i < len(matches); i++ {
			if matches[i-1].Score < matches[i].Score {
				t.Error("Matches should be sorted by descending similarity")
			}
		}

		// Limit should be respected
		if len(matches) > 3 {
			t.Errorf("Expected at most 3 matches, got %d", len(matches))
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		// Test lookup with high threshold (should not match)
		_, found, err := cache.Lookup("xyz", 0.99)
		if err != nil {
			t.Fatalf("No match lookup failed: %v", err)
		}
		if found {
			t.Error("Expected no match for high threshold")
		}
	})
}

func testCacheEviction(t *testing.T, cache *semanticcache.SemanticCache[string, string]) {
	// Test cache eviction by filling beyond capacity
	initialLen := cache.Len()

	// Add entries beyond capacity (capacity is 5)
	for i := range 7 {
		key := "evict" + string(rune(i+'0'))
		text := "text" + string(rune(i+'0'))
		err := cache.Set(key, text, "Value "+string(rune(i+'0')))
		if err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}

	// Length should not exceed capacity + initial entries
	finalLen := cache.Len()
	maxExpected := 5 + initialLen
	if finalLen > maxExpected {
		t.Errorf("Cache length %d exceeds expected maximum %d", finalLen, maxExpected)
	}

	// Test flush
	err := cache.Flush()
	if err != nil {
		t.Fatalf("Failed to flush cache: %v", err)
	}

	if cache.Len() != 0 {
		t.Errorf("Expected empty cache after flush, got length %d", cache.Len())
	}
}

func TestSemanticCacheWithOpenAI(t *testing.T) {
	// This test requires OpenAI API key
	t.Skip("OpenAI integration test requires API key - enable manually")

	/*
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			t.Skip("OPENAI_API_KEY not set")
		}

		config := types.BackendConfig{Capacity: 10}
		factory := &backends.BackendFactory[string, string]{}
		backend, err := factory.NewBackend(types.BackendLRU, config)
		if err != nil {
			t.Fatalf("Failed to create backend: %v", err)
		}

		providerConfig := openai.OpenAIConfig{APIKey: apiKey}
		provider, err := providers.NewOpenAIProvider(providerConfig)
		if err != nil {
			t.Fatalf("Failed to create OpenAI provider: %v", err)
		}

		cache, err := semanticcache.NewSemanticCache(backend, provider, nil)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		defer cache.Close()

		// Test with real OpenAI embeddings
		err = cache.Set("doc1", "The cat sat on the mat", "Story about a cat")
		if err != nil {
			t.Fatalf("Failed to set with OpenAI: %v", err)
		}

		// Should find semantically similar content
		value, found, err := cache.Lookup("A feline animal on a rug", 0.8)
		if err != nil {
			t.Fatalf("OpenAI lookup failed: %v", err)
		}

		if found {
			t.Logf("Found semantic match: %s", value)
		}
	*/
}

func TestConcurrentCacheAccess(t *testing.T) {
	config := types.BackendConfig{Capacity: 100}
	factory := &backends.BackendFactory[string, string]{}
	backend, err := factory.NewBackend(types.BackendLRU, config)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	provider := newMockProvider()
	cache, err := semanticcache.NewSemanticCache(backend, provider, nil)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// Test concurrent access
	done := make(chan bool, 30)

	// Start 10 writers
	for i := range 10 {
		go func(id int) {
			key := "concurrent" + string(rune(id+'0'))
			text := "cat" // Use known embedding
			value := "Value " + string(rune(id+'0'))
			_ = cache.Set(key, text, value)
			done <- true
		}(i)
	}

	// Start 10 readers
	for i := range 10 {
		go func(id int) {
			key := "concurrent" + string(rune(id+'0'))
			cache.Get(key)
			cache.Contains(key)
			done <- true
		}(i)
	}

	// Start 10 searchers
	for range 10 {
		go func() {
			_, _, _ = cache.Lookup("dog", 0.7)
			_, _ = cache.TopMatches("animal", 5)
			done <- true
		}()
	}

	// Wait for all operations to complete
	for range 30 {
		<-done
	}

	// Verify cache is still in consistent state
	length := cache.Len()
	if length < 0 || length > 100 {
		t.Errorf("Cache in inconsistent state, length: %d", length)
	}
}
