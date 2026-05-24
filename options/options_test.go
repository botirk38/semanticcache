package options

import (
	"context"
	"testing"

	"github.com/botirk38/semanticcache/similarity"
	"github.com/botirk38/semanticcache/types"
)

// ---------- mock provider ----------

type mockProvider struct{ shouldErr bool }

func (m *mockProvider) EmbedText(_ context.Context, _ string) ([]float64, error) {
	if m.shouldErr {
		return nil, &testError{"mock error"}
	}
	return []float64{0.1, 0.2, 0.3}, nil
}

func (m *mockProvider) Close() error { return nil }

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

// ---------- mock backend ----------

type mockBackend[K comparable, V any] struct{}

func (m *mockBackend[K, V]) Set(_ context.Context, _ K, _ []float64, _ V) error { return nil }
func (m *mockBackend[K, V]) Get(_ context.Context, _ K) (V, bool, error) {
	var zero V
	return zero, false, nil
}
func (m *mockBackend[K, V]) Delete(_ context.Context, _ K) error           { return nil }
func (m *mockBackend[K, V]) Contains(_ context.Context, _ K) (bool, error) { return false, nil }
func (m *mockBackend[K, V]) Flush(_ context.Context) error                 { return nil }
func (m *mockBackend[K, V]) Len(_ context.Context) (int, error)            { return 0, nil }
func (m *mockBackend[K, V]) Close() error                                  { return nil }

func (m *mockBackend[K, V]) Keys(_ context.Context) ([]K, error) { return nil, nil }
func (m *mockBackend[K, V]) GetEmbedding(_ context.Context, _ K) ([]float64, bool, error) {
	return nil, false, nil
}

// ---------- tests ----------

func TestConfigCreation(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if cfg.Comparator == nil {
			t.Error("expected default comparator")
		}
		if cfg.Backend != nil {
			t.Error("expected nil backend")
		}
		if cfg.Provider != nil {
			t.Error("expected nil provider")
		}
	})

	t.Run("Validation", func(t *testing.T) {
		cfg := NewConfig[string, string]()

		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error without backend/provider")
		}

		_ = cfg.Apply(WithLRUBackend[string, string](10))
		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error without provider")
		}

		_ = cfg.Apply(WithCustomProvider[string, string](&mockProvider{}))
		if err := cfg.Validate(); err != nil {
			t.Errorf("expected validation to pass: %v", err)
		}
	})
}

func TestBackendOptions(t *testing.T) {
	t.Run("LRUBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if err := cfg.Apply(WithLRUBackend[string, string](100)); err != nil {
			t.Fatalf("LRU failed: %v", err)
		}
		if cfg.Backend == nil {
			t.Error("expected backend set")
		}
	})

	t.Run("FIFOBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if err := cfg.Apply(WithFIFOBackend[string, string](100)); err != nil {
			t.Fatalf("FIFO failed: %v", err)
		}
		if cfg.Backend == nil {
			t.Error("expected backend set")
		}
	})

	t.Run("LFUBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if err := cfg.Apply(WithLFUBackend[string, string](100)); err != nil {
			t.Fatalf("LFU failed: %v", err)
		}
		if cfg.Backend == nil {
			t.Error("expected backend set")
		}
	})

	t.Run("CustomBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if err := cfg.Apply(WithCustomBackend[string, string](&mockBackend[string, string]{})); err != nil {
			t.Fatalf("custom backend failed: %v", err)
		}
		if cfg.Backend == nil {
			t.Error("expected backend set")
		}
	})

	t.Run("NilBackend", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if err := cfg.Apply(WithCustomBackend[string, string](nil)); err == nil {
			t.Error("expected error for nil backend")
		}
	})
}

func TestProviderOptions(t *testing.T) {
	t.Run("CustomProvider", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		prov := &mockProvider{}
		if err := cfg.Apply(WithCustomProvider[string, string](prov)); err != nil {
			t.Fatalf("custom provider failed: %v", err)
		}
		if cfg.Provider != prov {
			t.Error("expected provider set")
		}
	})

	t.Run("NilProvider", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if err := cfg.Apply(WithCustomProvider[string, string](nil)); err == nil {
			t.Error("expected error for nil provider")
		}
	})

	t.Run("OpenAIProvider_EmptyKey", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if err := cfg.Apply(WithOpenAIProvider[string, string]("")); err == nil {
			t.Error("expected error for empty API key")
		}
	})
}

func TestSimilarityOptions(t *testing.T) {
	t.Run("CustomSimilarity", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		custom := func(a, b []float64) float64 { return 0.5 }
		if err := cfg.Apply(WithSimilarityComparator[string, string](custom)); err != nil {
			t.Fatalf("custom similarity failed: %v", err)
		}
		if cfg.Comparator([]float64{1, 0}, []float64{0, 1}) != 0.5 {
			t.Error("expected 0.5")
		}
	})

	t.Run("NilSimilarity", func(t *testing.T) {
		cfg := NewConfig[string, string]()
		if err := cfg.Apply(WithSimilarityComparator[string, string](nil)); err == nil {
			t.Error("expected error for nil similarity")
		}
	})

	t.Run("BuiltinSimilarities", func(t *testing.T) {
		funcs := map[string]similarity.SimilarityFunc{
			"Cosine":      similarity.CosineSimilarity,
			"Euclidean":   similarity.EuclideanSimilarity,
			"DotProduct":  similarity.DotProductSimilarity,
			"Manhattan":   similarity.ManhattanSimilarity,
			"PearsonCorr": similarity.PearsonCorrelationSimilarity,
		}
		for name, fn := range funcs {
			t.Run(name, func(t *testing.T) {
				cfg := NewConfig[string, string]()
				if err := cfg.Apply(WithSimilarityComparator[string, string](fn)); err != nil {
					t.Fatalf("failed to set %s: %v", name, err)
				}
			})
		}
	})
}

var _ types.Backend[string, string] = (*mockBackend[string, string])(nil)
