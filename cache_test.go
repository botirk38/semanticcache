package semanticcache

import (
	"context"
	"testing"

	scerrors "github.com/botirk38/semanticcache/errors"
	"github.com/botirk38/semanticcache/options"
	"github.com/botirk38/semanticcache/similarity"
	"github.com/botirk38/semanticcache/types"
)

// ---------- mock provider ----------

type mockProvider struct {
	shouldErr  bool
	embeddings map[string][]float64
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		embeddings: map[string][]float64{
			"hello":            {1.0, 0.0, 0.0},
			"world":            {0.0, 1.0, 0.0},
			"test":             {0.0, 0.0, 1.0},
			"similar to hello": {0.9, 0.1, 0.0},
		},
	}
}

func (m *mockProvider) EmbedText(_ context.Context, text string) ([]float64, error) {
	if m.shouldErr {
		return nil, &testError{"mock embedding error"}
	}
	if emb, ok := m.embeddings[text]; ok {
		return emb, nil
	}
	return []float64{0.5, 0.5, 0.5}, nil
}

func (m *mockProvider) Close() error { return nil }

// ---------- mock backend ----------

type mockBackend[K comparable, V any] struct {
	data      map[K]types.Entry[V]
	shouldErr bool
}

func newMockBackend[K comparable, V any]() *mockBackend[K, V] {
	return &mockBackend[K, V]{data: make(map[K]types.Entry[V])}
}

func (m *mockBackend[K, V]) Set(_ context.Context, key K, embedding []float64, value V) error {
	if m.shouldErr {
		return &testError{"mock backend error"}
	}
	m.data[key] = types.Entry[V]{Embedding: embedding, Value: value}
	return nil
}

func (m *mockBackend[K, V]) Get(_ context.Context, key K) (V, bool, error) {
	if m.shouldErr {
		var zero V
		return zero, false, &testError{"mock backend error"}
	}
	e, ok := m.data[key]
	if !ok {
		var zero V
		return zero, false, nil
	}
	return e.Value, true, nil
}

func (m *mockBackend[K, V]) Delete(_ context.Context, key K) error {
	if m.shouldErr {
		return &testError{"mock backend error"}
	}
	delete(m.data, key)
	return nil
}

func (m *mockBackend[K, V]) Contains(_ context.Context, key K) (bool, error) {
	if m.shouldErr {
		return false, &testError{"mock backend error"}
	}
	_, ok := m.data[key]
	return ok, nil
}

func (m *mockBackend[K, V]) Flush(_ context.Context) error {
	if m.shouldErr {
		return &testError{"mock backend error"}
	}
	m.data = make(map[K]types.Entry[V])
	return nil
}

func (m *mockBackend[K, V]) Len(_ context.Context) (int, error) {
	if m.shouldErr {
		return 0, &testError{"mock backend error"}
	}
	return len(m.data), nil
}

func (m *mockBackend[K, V]) Close() error { return nil }

func (m *mockBackend[K, V]) Keys(_ context.Context) ([]K, error) {
	if m.shouldErr {
		return nil, &testError{"mock backend error"}
	}
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockBackend[K, V]) GetEmbedding(_ context.Context, key K) ([]float64, bool, error) {
	if m.shouldErr {
		return nil, false, &testError{"mock backend error"}
	}
	e, ok := m.data[key]
	if !ok {
		return nil, false, nil
	}
	return e.Embedding, true, nil
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

// ---------- tests ----------

func TestCacheCreation(t *testing.T) {
	t.Run("NewWithOptions", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		if cache == nil {
			t.Error("Expected non-nil cache")
		}
	})

	t.Run("NewSemanticCache", func(t *testing.T) {
		cache, err := NewSemanticCache(
			newMockBackend[string, string](),
			newMockProvider(),
			similarity.CosineSimilarity,
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		if cache == nil {
			t.Error("Expected non-nil cache")
		}
	})

	t.Run("MissingBackend", func(t *testing.T) {
		_, err := New(
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err == nil {
			t.Error("Expected error for missing backend")
		}
	})

	t.Run("MissingProvider", func(t *testing.T) {
		_, err := New(
			options.WithLRUBackend[string, string](10),
		)
		if err == nil {
			t.Error("Expected error for missing provider")
		}
	})

	t.Run("NilBackend", func(t *testing.T) {
		_, err := NewSemanticCache[string, string](nil, newMockProvider(), similarity.CosineSimilarity)
		if err == nil {
			t.Error("Expected error for nil backend")
		}
	})

	t.Run("NilProvider", func(t *testing.T) {
		_, err := NewSemanticCache(newMockBackend[string, string](), nil, similarity.CosineSimilarity)
		if err == nil {
			t.Error("Expected error for nil provider")
		}
	})

	t.Run("NilComparator", func(t *testing.T) {
		_, err := NewSemanticCache(newMockBackend[string, string](), newMockProvider(), nil)
		if err == nil {
			t.Error("Expected error for nil comparator")
		}
	})
}

func TestCacheOperations(t *testing.T) {
	cache, err := New(
		options.WithCustomBackend(newMockBackend[string, string]()),
		options.WithCustomProvider[string, string](newMockProvider()),
		options.WithSimilarityComparator[string, string](similarity.CosineSimilarity),
	)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	ctx := context.Background()

	t.Run("SetAndGet", func(t *testing.T) {
		if err := cache.Set(ctx, "key1", "hello", "value1"); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		value, found, err := cache.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !found || value != "value1" {
			t.Errorf("expected value1, got %s (found=%v)", value, found)
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		_, found, err := cache.Get(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if found {
			t.Error("expected not found")
		}
	})

	t.Run("Contains", func(t *testing.T) {
		_ = cache.Set(ctx, "key2", "world", "value2")
		found, err := cache.Contains(ctx, "key2")
		if err != nil {
			t.Fatalf("Contains failed: %v", err)
		}
		if !found {
			t.Error("expected to find key2")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		_ = cache.Set(ctx, "key3", "test", "value3")
		if err := cache.Delete(ctx, "key3"); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		found, _ := cache.Contains(ctx, "key3")
		if found {
			t.Error("expected key3 to be deleted")
		}
	})

	t.Run("Len", func(t *testing.T) {
		_ = cache.Flush(ctx)
		_ = cache.Set(ctx, "k", "hello", "v")
		n, err := cache.Len(ctx)
		if err != nil {
			t.Fatalf("Len failed: %v", err)
		}
		if n != 1 {
			t.Errorf("expected 1, got %d", n)
		}
	})

	t.Run("Flush", func(t *testing.T) {
		_ = cache.Flush(ctx)
		n, _ := cache.Len(ctx)
		if n != 0 {
			t.Errorf("expected 0 after flush, got %d", n)
		}
	})

	t.Run("ZeroValueKey", func(t *testing.T) {
		err := cache.Set(ctx, "", "hello", "value")
		if err == nil {
			t.Error("expected error for zero value key")
		}
	})
}

func TestSemanticSearch(t *testing.T) {
	cache, err := New(
		options.WithCustomBackend(newMockBackend[string, string]()),
		options.WithCustomProvider[string, string](newMockProvider()),
		options.WithSimilarityComparator[string, string](similarity.CosineSimilarity),
	)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "hello", "greeting")
	_ = cache.Set(ctx, "key2", "world", "planet")
	_ = cache.Set(ctx, "key3", "test", "examination")

	t.Run("Lookup", func(t *testing.T) {
		match, err := cache.Lookup(ctx, "similar to hello", 0.5)
		if err != nil {
			t.Fatalf("Lookup failed: %v", err)
		}
		if match == nil {
			t.Fatal("expected a match")
		}
		if match.Value != "greeting" {
			t.Errorf("expected greeting, got %s", match.Value)
		}
	})

	t.Run("LookupNoMatch", func(t *testing.T) {
		match, err := cache.Lookup(ctx, "completely different", 0.9)
		if err != nil {
			t.Fatalf("Lookup failed: %v", err)
		}
		if match != nil {
			t.Error("expected no match")
		}
	})

	t.Run("TopMatches", func(t *testing.T) {
		matches, err := cache.TopMatches(ctx, "hello", 2)
		if err != nil {
			t.Fatalf("TopMatches failed: %v", err)
		}
		if len(matches) == 0 {
			t.Error("expected at least one match")
		}
		if len(matches) > 2 {
			t.Errorf("expected at most 2, got %d", len(matches))
		}
		for i := 1; i < len(matches); i++ {
			if matches[i-1].Score < matches[i].Score {
				t.Error("expected descending score order")
			}
		}
	})

	t.Run("TopMatchesInvalidN", func(t *testing.T) {
		if _, err := cache.TopMatches(ctx, "hello", 0); err == nil {
			t.Error("expected error for n=0")
		}
		if _, err := cache.TopMatches(ctx, "hello", -1); err == nil {
			t.Error("expected error for n=-1")
		}
	})
}

func TestBatchOperations(t *testing.T) {
	cache, err := New(
		options.WithCustomBackend(newMockBackend[string, string]()),
		options.WithCustomProvider[string, string](newMockProvider()),
	)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	ctx := context.Background()

	t.Run("SetBatch", func(t *testing.T) {
		items := []BatchItem[string, string]{
			{Key: "b1", InputText: "hello", Value: "g1"},
			{Key: "b2", InputText: "world", Value: "p1"},
		}
		if err := cache.SetBatch(ctx, items); err != nil {
			t.Fatalf("SetBatch failed: %v", err)
		}
		v, found, _ := cache.Get(ctx, "b1")
		if !found || v != "g1" {
			t.Errorf("expected g1, got %s", v)
		}
	})

	t.Run("GetBatch", func(t *testing.T) {
		results, err := cache.GetBatch(ctx, []string{"b1", "b2", "missing"})
		if err != nil {
			t.Fatalf("GetBatch failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("DeleteBatch", func(t *testing.T) {
		if err := cache.DeleteBatch(ctx, []string{"b1", "b2"}); err != nil {
			t.Fatalf("DeleteBatch failed: %v", err)
		}
		n, _ := cache.Len(ctx)
		if n != 0 {
			t.Errorf("expected 0, got %d", n)
		}
	})

	t.Run("SetBatchZeroKey", func(t *testing.T) {
		items := []BatchItem[string, string]{{Key: "", InputText: "hello", Value: "v"}}
		if err := cache.SetBatch(ctx, items); err == nil {
			t.Error("expected error for zero-value key")
		}
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("ProviderError", func(t *testing.T) {
		cache, _ := New(
			options.WithCustomBackend(newMockBackend[string, string]()),
			options.WithCustomProvider[string, string](&mockProvider{shouldErr: true}),
		)
		ctx := context.Background()

		if err := cache.Set(ctx, "key", "text", "value"); err == nil {
			t.Error("expected provider error on Set")
		}
		if _, err := cache.Lookup(ctx, "text", 0.5); err == nil {
			t.Error("expected provider error on Lookup")
		}
		if _, err := cache.TopMatches(ctx, "text", 1); err == nil {
			t.Error("expected provider error on TopMatches")
		}
	})

	t.Run("BackendError", func(t *testing.T) {
		backend := &mockBackend[string, string]{
			data:      make(map[string]types.Entry[string]),
			shouldErr: true,
		}
		cache, _ := NewSemanticCache(backend, newMockProvider(), similarity.CosineSimilarity)
		ctx := context.Background()

		if err := cache.Set(ctx, "key", "text", "value"); err == nil {
			t.Error("expected backend error on Set")
		}
		if _, _, err := cache.Get(ctx, "key"); err == nil {
			t.Error("expected backend error on Get")
		}
		if _, err := cache.Contains(ctx, "key"); err == nil {
			t.Error("expected backend error on Contains")
		}
		if err := cache.Delete(ctx, "key"); err == nil {
			t.Error("expected backend error on Delete")
		}
		if _, err := cache.Len(ctx); err == nil {
			t.Error("expected backend error on Len")
		}
		if err := cache.Flush(ctx); err == nil {
			t.Error("expected backend error on Flush")
		}
	})
}

func TestClose(t *testing.T) {
	cache, _ := New(
		options.WithCustomBackend(newMockBackend[string, string]()),
		options.WithCustomProvider[string, string](newMockProvider()),
	)

	if err := cache.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	ctx := context.Background()
	if err := cache.Set(ctx, "k", "t", "v"); err != scerrors.ErrClosed {
		t.Errorf("expected ErrClosed, got %v", err)
	}
	if _, _, err := cache.Get(ctx, "k"); err != scerrors.ErrClosed {
		t.Errorf("expected ErrClosed, got %v", err)
	}

	// Double-close should be no-op.
	if err := cache.Close(); err != nil {
		t.Errorf("double Close returned error: %v", err)
	}
}

func TestWithDifferentTypes(t *testing.T) {
	t.Run("IntKey", func(t *testing.T) {
		cache, _ := New(
			options.WithCustomBackend(newMockBackend[int, string]()),
			options.WithCustomProvider[int, string](newMockProvider()),
		)
		ctx := context.Background()
		_ = cache.Set(ctx, 123, "hello", "world")
		v, found, _ := cache.Get(ctx, 123)
		if !found || v != "world" {
			t.Errorf("expected world, got %s", v)
		}
	})

	t.Run("StructValue", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		cache, _ := New(
			options.WithCustomBackend(newMockBackend[string, TestStruct]()),
			options.WithCustomProvider[string, TestStruct](newMockProvider()),
		)
		ctx := context.Background()
		_ = cache.Set(ctx, "person", "hello", TestStruct{Name: "John", Age: 30})
		v, found, _ := cache.Get(ctx, "person")
		if !found || v.Name != "John" || v.Age != 30 {
			t.Errorf("expected {John 30}, got %+v", v)
		}
	})
}
