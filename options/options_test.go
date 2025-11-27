package options

import (
	"context"
	"testing"

	"github.com/botirk38/semanticcache/similarity"
	"github.com/botirk38/semanticcache/types"
)

// Mock provider for testing
type mockProvider struct {
	shouldErr bool
}

func (m *mockProvider) EmbedText(text string) ([]float64, error) {
	if m.shouldErr {
		return nil, &testError{"mock error"}
	}
	return []float64{0.1, 0.2, 0.3}, nil
}

func (m *mockProvider) GetMaxTokens() int {
	return 8191 // Default OpenAI limit
}

func (m *mockProvider) Close() {}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestConfigCreation(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if cfg.Comparator == nil {
			t.Error("Expected default comparator to be set")
		}
		if cfg.Backend != nil {
			t.Error("Expected backend to be nil initially")
		}
		if cfg.Provider != nil {
			t.Error("Expected provider to be nil initially")
		}
	})

	t.Run("Validation", func(t *testing.T) {
		cfg := NewConfig[string, string]()

		// Should fail without backend and provider
		err := cfg.Validate()
		if err == nil {
			t.Error("Expected validation error for missing backend and provider")
		}

		// Set backend, should still fail without provider
		err = cfg.Apply(WithLRUBackend[string, string](10))
		if err != nil {
			t.Fatalf("Failed to apply backend option: %v", err)
		}

		err = cfg.Validate()
		if err == nil {
			t.Error("Expected validation error for missing provider")
		}

		// Set provider, should now pass
		err = cfg.Apply(WithCustomProvider[string, string](&mockProvider{}))
		if err != nil {
			t.Fatalf("Failed to apply provider option: %v", err)
		}

		err = cfg.Validate()
		if err != nil {
			t.Errorf("Expected validation to pass, got: %v", err)
		}
	})
}

func TestBackendOptions(t *testing.T) {
	t.Run("LRUBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		err := cfg.Apply(WithLRUBackend[string, string](100))
		if err != nil {
			t.Fatalf("Failed to set LRU backend: %v", err)
		}
		if cfg.Backend == nil {
			t.Error("Expected backend to be set")
		}
	})

	t.Run("FIFOBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		err := cfg.Apply(WithFIFOBackend[string, string](100))
		if err != nil {
			t.Fatalf("Failed to set FIFO backend: %v", err)
		}
		if cfg.Backend == nil {
			t.Error("Expected backend to be set")
		}
	})

	t.Run("LFUBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		err := cfg.Apply(WithLFUBackend[string, string](100))
		if err != nil {
			t.Fatalf("Failed to set LFU backend: %v", err)
		}
		if cfg.Backend == nil {
			t.Error("Expected backend to be set")
		}
	})

	t.Run("CustomBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		mockBackend := &mockBackend[string, string]{}

		err := cfg.Apply(WithCustomBackend(mockBackend))
		if err != nil {
			t.Fatalf("Failed to set custom backend: %v", err)
		}
		if cfg.Backend == nil {
			t.Error("Expected custom backend to be set")
		}
	})

	t.Run("NilBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		err := cfg.Apply(WithCustomBackend[string, string](nil))
		if err == nil {
			t.Error("Expected error for nil backend")
		}
	})
}

func TestProviderOptions(t *testing.T) {
	t.Run("CustomProvider", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		mockProv := &mockProvider{}

		err := cfg.Apply(WithCustomProvider[string, string](mockProv))
		if err != nil {
			t.Fatalf("Failed to set custom provider: %v", err)
		}
		if cfg.Provider != mockProv {
			t.Error("Expected custom provider to be set")
		}
	})

	t.Run("NilProvider", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		err := cfg.Apply(WithCustomProvider[string, string](nil))
		if err == nil {
			t.Error("Expected error for nil provider")
		}
	})

	t.Run("OpenAIProvider", func(t *testing.T) {
		cfg := NewConfig[string, string]()

		// This should fail with invalid API key, but not crash
		err := cfg.Apply(WithOpenAIProvider[string, string](""))
		if err == nil {
			t.Error("Expected error for empty API key")
		}
	})
}

func TestSimilarityOptions(t *testing.T) {
	t.Run("CustomSimilarity", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		customSim := func(a, b []float64) float64 { return 0.5 }

		err := cfg.Apply(WithSimilarityComparator[string, string](customSim))
		if err != nil {
			t.Fatalf("Failed to set custom similarity: %v", err)
		}

		// Test that the custom function works
		result := cfg.Comparator([]float64{1, 0}, []float64{0, 1})
		if result != 0.5 {
			t.Errorf("Expected 0.5, got %f", result)
		}
	})

	t.Run("NilSimilarity", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		err := cfg.Apply(WithSimilarityComparator[string, string](nil))
		if err == nil {
			t.Error("Expected error for nil similarity function")
		}
	})

	t.Run("BuiltinSimilarities", func(t *testing.T) {
		similarities := map[string]similarity.SimilarityFunc{
			"Cosine":      similarity.CosineSimilarity,
			"Euclidean":   similarity.EuclideanSimilarity,
			"DotProduct":  similarity.DotProductSimilarity,
			"Manhattan":   similarity.ManhattanSimilarity,
			"PearsonCorr": similarity.PearsonCorrelationSimilarity,
		}

		for name, simFunc := range similarities {
			t.Run(name, func(t *testing.T) {
				cfg := NewConfig[string, string]()
				err := cfg.Apply(WithSimilarityComparator[string, string](simFunc))
				if err != nil {
					t.Fatalf("Failed to set %s similarity: %v", name, err)
				}
				if cfg.Comparator == nil {
					t.Errorf("Expected %s similarity to be set", name)
				}
			})
		}
	})
}

// Mock backend for testing
type mockBackend[K comparable, V any] struct{}

func (m *mockBackend[K, V]) Set(ctx context.Context, key K, entry types.Entry[V]) error {
	return nil
}

func (m *mockBackend[K, V]) Get(ctx context.Context, key K) (types.Entry[V], bool, error) {
	return types.Entry[V]{}, false, nil
}

func (m *mockBackend[K, V]) Delete(ctx context.Context, key K) error {
	return nil
}

func (m *mockBackend[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	return false, nil
}

func (m *mockBackend[K, V]) Keys(ctx context.Context) ([]K, error) {
	return nil, nil
}

func (m *mockBackend[K, V]) Len(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockBackend[K, V]) Flush(ctx context.Context) error {
	return nil
}

func (m *mockBackend[K, V]) GetEmbedding(ctx context.Context, key K) ([]float64, bool, error) {
	return nil, false, nil
}

func (m *mockBackend[K, V]) Close() error {
	return nil
}

func (m *mockBackend[K, V]) SetAsync(ctx context.Context, key K, entry types.Entry[V]) <-chan error {
	errCh := make(chan error, 1)
	errCh <- nil
	close(errCh)
	return errCh
}

func (m *mockBackend[K, V]) GetAsync(ctx context.Context, key K) <-chan types.AsyncGetResult[V] {
	resultCh := make(chan types.AsyncGetResult[V], 1)
	resultCh <- types.AsyncGetResult[V]{Entry: types.Entry[V]{}, Found: false, Error: nil}
	close(resultCh)
	return resultCh
}

func (m *mockBackend[K, V]) DeleteAsync(ctx context.Context, key K) <-chan error {
	errCh := make(chan error, 1)
	errCh <- nil
	close(errCh)
	return errCh
}

func (m *mockBackend[K, V]) GetBatchAsync(ctx context.Context, keys []K) <-chan types.AsyncBatchResult[K, V] {
	resultCh := make(chan types.AsyncBatchResult[K, V], 1)
	resultCh <- types.AsyncBatchResult[K, V]{Entries: make(map[K]types.Entry[V]), Error: nil}
	close(resultCh)
	return resultCh
}
