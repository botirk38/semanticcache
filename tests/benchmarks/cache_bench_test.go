package benchmarks_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/botirk38/semanticcache"
	"github.com/botirk38/semanticcache/backends"
	"github.com/botirk38/semanticcache/types"
)

// Benchmark provider that returns predictable embeddings
type benchProvider struct{}

func (b *benchProvider) EmbedText(text string) ([]float32, error) {
	// Generate deterministic embedding based on text hash
	hash := hashString(text)
	embedding := make([]float32, 1536) // OpenAI embedding size

	for i := range embedding {
		embedding[i] = float32((hash+i)%1000) / 1000.0
	}

	return embedding, nil
}

func (b *benchProvider) Close() {}

// Simple hash function for deterministic embeddings
func hashString(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

func setupCache(b *testing.B, backendType types.BackendType, capacity int) *semanticcache.SemanticCache[string, string] {
	config := types.BackendConfig{Capacity: capacity}
	factory := &backends.BackendFactory[string, string]{}
	backend, err := factory.NewBackend(backendType, config)
	if err != nil {
		b.Fatalf("Failed to create backend: %v", err)
	}

	provider := &benchProvider{}
	cache, err := semanticcache.NewSemanticCache(backend, provider, nil)
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	return cache
}

func BenchmarkCacheSet(b *testing.B) {
	backends := []types.BackendType{
		types.BackendLRU,
		types.BackendFIFO,
		types.BackendLFU,
	}

	for _, backendType := range backends {
		b.Run(string(backendType), func(b *testing.B) {
			cache := setupCache(b, backendType, 10000)
			defer cache.Close()

			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				key := "key" + strconv.Itoa(i)
				text := "text" + strconv.Itoa(i%100) // Cycle through 100 different texts
				value := "value" + strconv.Itoa(i)

				err := cache.Set(key, text, value)
				if err != nil {
					b.Fatalf("Set failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkCacheGet(b *testing.B) {
	backends := []types.BackendType{
		types.BackendLRU,
		types.BackendFIFO,
		types.BackendLFU,
	}

	for _, backendType := range backends {
		b.Run(string(backendType), func(b *testing.B) {
			cache := setupCache(b, backendType, 1000)
			defer cache.Close()

			// Pre-populate cache
			for i := range 1000 {
				key := "key" + strconv.Itoa(i)
				text := "text" + strconv.Itoa(i%100)
				value := "value" + strconv.Itoa(i)
				_ = cache.Set(key, text, value)
			}

			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				key := "key" + strconv.Itoa(i%1000)
				_, _ = cache.Get(key)
			}
		})
	}
}

func BenchmarkCacheLookup(b *testing.B) {
	backends := []types.BackendType{
		types.BackendLRU,
		types.BackendFIFO,
		types.BackendLFU,
	}

	for _, backendType := range backends {
		b.Run(string(backendType), func(b *testing.B) {
			cache := setupCache(b, backendType, 1000)
			defer cache.Close()

			// Pre-populate cache
			for i := range 1000 {
				key := "key" + strconv.Itoa(i)
				text := "text" + strconv.Itoa(i%100)
				value := "value" + strconv.Itoa(i)
				_ = cache.Set(key, text, value)
			}

			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				text := "text" + strconv.Itoa(i%100)
				_, _, _ = cache.Lookup(text, 0.8)
			}
		})
	}
}

func BenchmarkCacheTopMatches(b *testing.B) {
	cache := setupCache(b, types.BackendLRU, 1000)
	defer cache.Close()

	// Pre-populate cache
	for i := range 1000 {
		key := "key" + strconv.Itoa(i)
		text := "text" + strconv.Itoa(i%100)
		value := "value" + strconv.Itoa(i)
		_ = cache.Set(key, text, value)
	}

	for i := 0; b.Loop(); i++ {
		text := "text" + strconv.Itoa(i%50)
		_, _ = cache.TopMatches(text, 10)
	}
}

func BenchmarkBackendOperations(b *testing.B) {
	backend_types := []types.BackendType{
		types.BackendLRU,
		types.BackendFIFO,
		types.BackendLFU,
	}

	for _, backendType := range backend_types {
		b.Run(string(backendType)+"_Set", func(b *testing.B) {
			config := types.BackendConfig{Capacity: 10000}
			factory := &backends.BackendFactory[string, string]{}
			backend, err := factory.NewBackend(backendType, config)
			if err != nil {
				b.Fatalf("Failed to create backend: %v", err)
			}
			defer func() { _ = backend.Close() }()

			entry := types.Entry[string]{
				Embedding: []float32{0.1, 0.2, 0.3},
				Value:     "test_value",
			}

			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				key := "key" + strconv.Itoa(i)
				_ = backend.Set(context.Background(), key, entry)
			}
		})

		b.Run(string(backendType)+"_Get", func(b *testing.B) {
			config := types.BackendConfig{Capacity: 1000}
			factory := &backends.BackendFactory[string, string]{}
			backend, err := factory.NewBackend(backendType, config)
			if err != nil {
				b.Fatalf("Failed to create backend: %v", err)
			}
			defer func() { _ = backend.Close() }()

			// Pre-populate
			entry := types.Entry[string]{
				Embedding: []float32{0.1, 0.2, 0.3},
				Value:     "test_value",
			}
			for i := range 1000 {
				key := "key" + strconv.Itoa(i)
				_ = backend.Set(context.Background(), key, entry)
			}

			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				key := "key" + strconv.Itoa(i%1000)
				_, _, _ = backend.Get(context.Background(), key)
			}
		})
	}
}

func BenchmarkEmbeddingProvider(b *testing.B) {
	provider := &benchProvider{}
	defer provider.Close()

	b.Run("ShortText", func(b *testing.B) {
		for b.Loop() {
			_, _ = provider.EmbedText("hello world")
		}
	})

	b.Run("LongText", func(b *testing.B) {
		longText := "This is a much longer text that simulates a typical document or paragraph that might be embedded in a semantic cache system. It contains multiple sentences and various words to test performance."

		for b.Loop() {
			_, _ = provider.EmbedText(longText)
		}
	})

	b.Run("VaryingText", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			text := "text number " + strconv.Itoa(i)
			_, _ = provider.EmbedText(text)
		}
	})
}

func BenchmarkCacheConcurrency(b *testing.B) {
	cache := setupCache(b, types.BackendLRU, 10000)
	defer cache.Close()

	// Pre-populate cache
	for i := range 1000 {
		key := "key" + strconv.Itoa(i)
		text := "text" + strconv.Itoa(i%100)
		value := "value" + strconv.Itoa(i)
		_ = cache.Set(key, text, value)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0: // Set
				key := "parallel_key" + strconv.Itoa(i)
				text := "text" + strconv.Itoa(i%50)
				value := "value" + strconv.Itoa(i)
				_ = cache.Set(key, text, value)
			case 1: // Get
				key := "key" + strconv.Itoa(i%1000)
				cache.Get(key)
			case 2: // Lookup
				text := "text" + strconv.Itoa(i%100)
				_, _, _ = cache.Lookup(text, 0.8)
			case 3: // TopMatches
				text := "text" + strconv.Itoa(i%50)
				_, _ = cache.TopMatches(text, 5)
			}
			i++
		}
	})
}

func BenchmarkMemoryUsage(b *testing.B) {
	// Test memory usage patterns
	b.Run("SmallCache", func(b *testing.B) {
		for b.Loop() {
			cache := setupCache(b, types.BackendLRU, 100)

			// Fill cache
			for j := range 150 { // Overfill to test eviction
				key := "key" + strconv.Itoa(j)
				text := "text" + strconv.Itoa(j%20)
				value := "value" + strconv.Itoa(j)
				_ = cache.Set(key, text, value)
			}

			cache.Close()
		}
	})

	b.Run("LargeCache", func(b *testing.B) {
		for b.Loop() {
			cache := setupCache(b, types.BackendLRU, 10000)

			// Fill cache
			for j := range 5000 {
				key := "key" + strconv.Itoa(j)
				text := "text" + strconv.Itoa(j%200)
				value := "large value with more content " + strconv.Itoa(j)
				_ = cache.Set(key, text, value)
			}

			cache.Close()
		}
	})
}
