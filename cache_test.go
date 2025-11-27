package semanticcache

import (
	"context"
	"testing"

	"github.com/botirk38/semanticcache/options"
	"github.com/botirk38/semanticcache/similarity"
	"github.com/botirk38/semanticcache/types"
)

// Mock provider for testing
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

func (m *mockProvider) EmbedText(text string) ([]float64, error) {
	if m.shouldErr {
		return nil, &testError{"mock embedding error"}
	}
	if embedding, exists := m.embeddings[text]; exists {
		return embedding, nil
	}
	// Default embedding for unknown text
	return []float64{0.5, 0.5, 0.5}, nil
}

func (m *mockProvider) Close() {}

// Mock backend for testing
type mockBackend[K comparable, V any] struct {
	data      map[K]types.Entry[V]
	shouldErr bool
}

func newMockBackend[K comparable, V any]() *mockBackend[K, V] {
	return &mockBackend[K, V]{
		data: make(map[K]types.Entry[V]),
	}
}

func (m *mockBackend[K, V]) Set(ctx context.Context, key K, entry types.Entry[V]) error {
	if m.shouldErr {
		return &testError{"mock backend error"}
	}
	m.data[key] = entry
	return nil
}

func (m *mockBackend[K, V]) Get(ctx context.Context, key K) (types.Entry[V], bool, error) {
	if m.shouldErr {
		return types.Entry[V]{}, false, &testError{"mock backend error"}
	}
	entry, found := m.data[key]
	return entry, found, nil
}

func (m *mockBackend[K, V]) Delete(ctx context.Context, key K) error {
	if m.shouldErr {
		return &testError{"mock backend error"}
	}
	delete(m.data, key)
	return nil
}

func (m *mockBackend[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	if m.shouldErr {
		return false, &testError{"mock backend error"}
	}
	_, found := m.data[key]
	return found, nil
}

func (m *mockBackend[K, V]) Keys(ctx context.Context) ([]K, error) {
	if m.shouldErr {
		return nil, &testError{"mock backend error"}
	}
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockBackend[K, V]) Len(ctx context.Context) (int, error) {
	if m.shouldErr {
		return 0, &testError{"mock backend error"}
	}
	return len(m.data), nil
}

func (m *mockBackend[K, V]) Flush(ctx context.Context) error {
	if m.shouldErr {
		return &testError{"mock backend error"}
	}
	m.data = make(map[K]types.Entry[V])
	return nil
}

func (m *mockBackend[K, V]) GetEmbedding(ctx context.Context, key K) ([]float64, bool, error) {
	if m.shouldErr {
		return nil, false, &testError{"mock backend error"}
	}
	entry, found := m.data[key]
	if !found {
		return nil, false, nil
	}
	return entry.Embedding, true, nil
}

func (m *mockBackend[K, V]) Close() error {
	return nil
}

// Async methods for mockBackend
func (m *mockBackend[K, V]) SetAsync(ctx context.Context, key K, entry types.Entry[V]) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- m.Set(ctx, key, entry)
	}()
	return errCh
}

func (m *mockBackend[K, V]) GetAsync(ctx context.Context, key K) <-chan types.AsyncGetResult[V] {
	resultCh := make(chan types.AsyncGetResult[V], 1)
	go func() {
		defer close(resultCh)
		entry, found, err := m.Get(ctx, key)
		resultCh <- types.AsyncGetResult[V]{
			Entry: entry,
			Found: found,
			Error: err,
		}
	}()
	return resultCh
}

func (m *mockBackend[K, V]) DeleteAsync(ctx context.Context, key K) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- m.Delete(ctx, key)
	}()
	return errCh
}

func (m *mockBackend[K, V]) GetBatchAsync(ctx context.Context, keys []K) <-chan types.AsyncBatchResult[K, V] {
	resultCh := make(chan types.AsyncBatchResult[K, V], 1)
	go func() {
		defer close(resultCh)
		entries := make(map[K]types.Entry[V])
		for _, key := range keys {
			if entry, found, err := m.Get(ctx, key); err == nil && found {
				entries[key] = entry
			}
		}
		resultCh <- types.AsyncBatchResult[K, V]{
			Entries: entries,
			Error:   nil,
		}
	}()
	return resultCh
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestSemanticCacheCreation(t *testing.T) {
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
		backend := newMockBackend[string, string]()
		provider := newMockProvider()
		comparator := similarity.CosineSimilarity

		cache, err := NewSemanticCache(backend, provider, comparator)
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

func TestSemanticCacheOperations(t *testing.T) {
	cache, err := New(
		options.WithCustomBackend(newMockBackend[string, string]()),
		options.WithCustomProvider[string, string](newMockProvider()),
		options.WithSimilarityComparator[string, string](similarity.CosineSimilarity),
	)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		err := cache.Set(ctx, "key1", "hello", "value1")
		if err != nil {
			t.Fatalf("Failed to set: %v", err)
		}

		value, found, err := cache.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Failed to get: %v", err)
		}
		if !found {
			t.Error("Expected to find key1")
		}
		if value != "value1" {
			t.Errorf("Expected 'value1', got %s", value)
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		value, found, err := cache.Get(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if found {
			t.Error("Expected not to find nonexistent key")
		}
		if value != "" {
			t.Errorf("Expected empty value, got %s", value)
		}
	})

	t.Run("Contains", func(t *testing.T) {
		err := cache.Set(ctx, "key2", "world", "value2")
		if err != nil {
			t.Fatalf("Failed to set: %v", err)
		}

		found, err := cache.Contains(ctx, "key2")
		if err != nil {
			t.Fatalf("Failed to check contains: %v", err)
		}
		if !found {
			t.Error("Expected to find key2")
		}

		found, err = cache.Contains(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("Failed to check contains: %v", err)
		}
		if found {
			t.Error("Expected not to find nonexistent key")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := cache.Set(ctx, "key3", "test", "value3")
		if err != nil {
			t.Fatalf("Failed to set: %v", err)
		}

		err = cache.Delete(ctx, "key3")
		if err != nil {
			t.Fatalf("Failed to delete: %v", err)
		}

		found, err := cache.Contains(ctx, "key3")
		if err != nil {
			t.Fatalf("Failed to check contains: %v", err)
		}
		if found {
			t.Error("Expected key3 to be deleted")
		}
	})

	t.Run("Len", func(t *testing.T) {
		err := cache.Flush(ctx)
		if err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}

		length, err := cache.Len(ctx)
		if err != nil {
			t.Fatalf("Failed to get length: %v", err)
		}
		if length != 0 {
			t.Errorf("Expected length 0, got %d", length)
		}

		err = cache.Set(ctx, "key4", "hello", "value4")
		if err != nil {
			t.Fatalf("Failed to set: %v", err)
		}

		length, err = cache.Len(ctx)
		if err != nil {
			t.Fatalf("Failed to get length: %v", err)
		}
		if length != 1 {
			t.Errorf("Expected length 1, got %d", length)
		}
	})

	t.Run("Flush", func(t *testing.T) {
		err := cache.Set(ctx, "key5", "hello", "value5")
		if err != nil {
			t.Fatalf("Failed to set: %v", err)
		}

		err = cache.Flush(ctx)
		if err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}

		length, err := cache.Len(ctx)
		if err != nil {
			t.Fatalf("Failed to get length: %v", err)
		}
		if length != 0 {
			t.Errorf("Expected length 0 after flush, got %d", length)
		}
	})

	t.Run("ZeroValueKey", func(t *testing.T) {
		err := cache.Set(ctx, "", "hello", "value")
		if err == nil {
			t.Error("Expected error for zero value key")
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

	// Set up test data
	err = cache.Set(ctx, "key1", "hello", "greeting")
	if err != nil {
		t.Fatalf("Failed to set key1: %v", err)
	}

	err = cache.Set(ctx, "key2", "world", "planet")
	if err != nil {
		t.Fatalf("Failed to set key2: %v", err)
	}

	err = cache.Set(ctx, "key3", "test", "examination")
	if err != nil {
		t.Fatalf("Failed to set key3: %v", err)
	}

	t.Run("Lookup", func(t *testing.T) {
		match, err := cache.Lookup(ctx, "similar to hello", 0.5)
		if err != nil {
			t.Fatalf("Failed to lookup: %v", err)
		}
		if match == nil {
			t.Error("Expected to find a match")
		} else {
			if match.Value != "greeting" {
				t.Errorf("Expected 'greeting', got %s", match.Value)
			}
			if match.Score < 0.5 {
				t.Errorf("Expected score >= 0.5, got %f", match.Score)
			}
		}
	})

	t.Run("LookupNoMatch", func(t *testing.T) {
		match, err := cache.Lookup(ctx, "completely different", 0.9)
		if err != nil {
			t.Fatalf("Failed to lookup: %v", err)
		}
		if match != nil {
			t.Error("Expected no match for high threshold")
		}
	})

	t.Run("TopMatches", func(t *testing.T) {
		matches, err := cache.TopMatches(ctx, "hello", 2)
		if err != nil {
			t.Fatalf("Failed to get top matches: %v", err)
		}
		if len(matches) == 0 {
			t.Error("Expected at least one match")
		}
		if len(matches) > 2 {
			t.Errorf("Expected at most 2 matches, got %d", len(matches))
		}

		// Check that matches are sorted by score (descending)
		for i := 1; i < len(matches); i++ {
			if matches[i-1].Score < matches[i].Score {
				t.Error("Expected matches to be sorted by descending score")
			}
		}
	})

	t.Run("TopMatchesInvalidN", func(t *testing.T) {
		_, err := cache.TopMatches(ctx, "hello", 0)
		if err == nil {
			t.Error("Expected error for n <= 0")
		}

		_, err = cache.TopMatches(ctx, "hello", -1)
		if err == nil {
			t.Error("Expected error for n <= 0")
		}
	})
}

func TestBatchOperations(t *testing.T) {
	cache, err := New(
		options.WithCustomBackend(newMockBackend[string, string]()),
		options.WithCustomProvider[string, string](newMockProvider()),
		options.WithSimilarityComparator[string, string](similarity.CosineSimilarity),
	)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	ctx := context.Background()

	t.Run("SetBatch", func(t *testing.T) {
		items := []BatchItem[string, string]{
			{Key: "batch1", InputText: "hello", Value: "greeting1"},
			{Key: "batch2", InputText: "world", Value: "planet1"},
			{Key: "batch3", InputText: "test", Value: "examination1"},
		}

		err := cache.SetBatch(ctx, items)
		if err != nil {
			t.Fatalf("Failed to set batch: %v", err)
		}

		// Verify all items were set
		for _, item := range items {
			value, found, err := cache.Get(ctx, item.Key)
			if err != nil {
				t.Fatalf("Failed to get %s: %v", item.Key, err)
			}
			if !found {
				t.Errorf("Expected to find %s", item.Key)
			}
			if value != item.Value {
				t.Errorf("Expected %s, got %s", item.Value, value)
			}
		}
	})

	t.Run("GetBatch", func(t *testing.T) {
		keys := []string{"batch1", "batch2", "nonexistent"}
		results, err := cache.GetBatch(ctx, keys)
		if err != nil {
			t.Fatalf("Failed to get batch: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		if results["batch1"] != "greeting1" {
			t.Errorf("Expected 'greeting1', got %s", results["batch1"])
		}

		if results["batch2"] != "planet1" {
			t.Errorf("Expected 'planet1', got %s", results["batch2"])
		}

		if _, exists := results["nonexistent"]; exists {
			t.Error("Expected nonexistent key to not be in results")
		}
	})

	t.Run("DeleteBatch", func(t *testing.T) {
		keys := []string{"batch1", "batch2"}
		err := cache.DeleteBatch(ctx, keys)
		if err != nil {
			t.Fatalf("Failed to delete batch: %v", err)
		}

		// Verify deletion
		for _, key := range keys {
			found, err := cache.Contains(ctx, key)
			if err != nil {
				t.Fatalf("Failed to check contains for %s: %v", key, err)
			}
			if found {
				t.Errorf("Expected %s to be deleted", key)
			}
		}
	})

	t.Run("SetBatchEmpty", func(t *testing.T) {
		err := cache.SetBatch(ctx, []BatchItem[string, string]{})
		if err != nil {
			t.Fatalf("Failed to set empty batch: %v", err)
		}
	})

	t.Run("SetBatchZeroKey", func(t *testing.T) {
		items := []BatchItem[string, string]{
			{Key: "", InputText: "hello", Value: "greeting"},
		}

		err := cache.SetBatch(ctx, items)
		if err == nil {
			t.Error("Expected error for zero value key in batch")
		}
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("ProviderError", func(t *testing.T) {
		provider := &mockProvider{shouldErr: true}
		cache, err := New(
			options.WithCustomBackend(newMockBackend[string, string]()),
			options.WithCustomProvider[string, string](provider),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		ctx := context.Background()
		err = cache.Set(ctx, "key", "text", "value")
		if err == nil {
			t.Error("Expected error from provider")
		}

		_, err = cache.Lookup(ctx, "text", 0.5)
		if err == nil {
			t.Error("Expected error from provider in lookup")
		}

		_, err = cache.TopMatches(ctx, "text", 1)
		if err == nil {
			t.Error("Expected error from provider in top matches")
		}
	})

	t.Run("BackendError", func(t *testing.T) {
		backend := &mockBackend[string, string]{shouldErr: true}
		cache, err := NewSemanticCache(backend, newMockProvider(), similarity.CosineSimilarity)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		ctx := context.Background()

		err = cache.Set(ctx, "key", "text", "value")
		if err == nil {
			t.Error("Expected error from backend")
		}

		_, _, err = cache.Get(ctx, "key")
		if err == nil {
			t.Error("Expected error from backend")
		}

		_, err = cache.Contains(ctx, "key")
		if err == nil {
			t.Error("Expected error from backend")
		}

		err = cache.Delete(ctx, "key")
		if err == nil {
			t.Error("Expected error from backend")
		}

		_, err = cache.Len(ctx)
		if err == nil {
			t.Error("Expected error from backend")
		}

		err = cache.Flush(ctx)
		if err == nil {
			t.Error("Expected error from backend")
		}
	})
}

func TestClose(t *testing.T) {
	cache, err := New(
		options.WithCustomBackend(newMockBackend[string, string]()),
		options.WithCustomProvider[string, string](newMockProvider()),
	)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	err = cache.Close()
	if err != nil {
		t.Fatalf("Failed to close cache: %v", err)
	}
}

func TestWithDifferentTypes(t *testing.T) {
	t.Run("IntKey", func(t *testing.T) {
		cache, err := New(
			options.WithCustomBackend(newMockBackend[int, string]()),
			options.WithCustomProvider[int, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		ctx := context.Background()
		err = cache.Set(ctx, 123, "hello", "world")
		if err != nil {
			t.Fatalf("Failed to set: %v", err)
		}

		value, found, err := cache.Get(ctx, 123)
		if err != nil {
			t.Fatalf("Failed to get: %v", err)
		}
		if !found {
			t.Error("Expected to find key")
		}
		if value != "world" {
			t.Errorf("Expected 'world', got %s", value)
		}
	})

	t.Run("StructValue", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}

		cache, err := New(
			options.WithCustomBackend(newMockBackend[string, TestStruct]()),
			options.WithCustomProvider[string, TestStruct](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		ctx := context.Background()
		testValue := TestStruct{Name: "John", Age: 30}
		err = cache.Set(ctx, "person", "hello", testValue)
		if err != nil {
			t.Fatalf("Failed to set: %v", err)
		}

		value, found, err := cache.Get(ctx, "person")
		if err != nil {
			t.Fatalf("Failed to get: %v", err)
		}
		if !found {
			t.Error("Expected to find key")
		}
		if value.Name != "John" || value.Age != 30 {
			t.Errorf("Expected {John 30}, got %+v", value)
		}
	})
}

// Test chunking functionality
func TestChunkingIntegration(t *testing.T) {
	t.Run("ChunkingEnabledByDefault", func(t *testing.T) {
		cache, err := New[string, string](
			options.WithLRUBackend[string, string](100),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		defer cache.Close()

		// Chunking should be enabled by default
		if !cache.enableChunking {
			t.Error("Expected chunking to be enabled by default")
		}
		if cache.chunker == nil {
			t.Error("Expected chunker to be initialized")
		}
	})

	t.Run("ChunkingCanBeDisabled", func(t *testing.T) {
		cache, err := New[string, string](
			options.WithLRUBackend[string, string](100),
			options.WithCustomProvider[string, string](newMockProvider()),
			options.WithoutChunking[string, string](),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}
		defer cache.Close()

		// Chunking should be disabled
		if cache.enableChunking {
			t.Error("Expected chunking to be disabled")
		}
	})

	t.Run("AggregateEmbeddingsEmpty", func(t *testing.T) {
		cache := &SemanticCache[string, string]{}

		result := cache.aggregateEmbeddings(nil)
		if result != nil {
			t.Error("Expected nil for nil embeddings")
		}

		result = cache.aggregateEmbeddings([][]float64{})
		if result != nil {
			t.Error("Expected nil for empty slice")
		}
	})

	t.Run("AggregateEmbeddingsSingle", func(t *testing.T) {
		cache := &SemanticCache[string, string]{}

		single := [][]float64{{0.1, 0.2, 0.3}}
		result := cache.aggregateEmbeddings(single)

		if len(result) != 3 {
			t.Fatalf("Expected length 3, got %d", len(result))
		}
		for i, v := range result {
			if v != single[0][i] {
				t.Errorf("Expected %f at index %d, got %f", single[0][i], i, v)
			}
		}
	})

	t.Run("AggregateEmbeddingsMultiple", func(t *testing.T) {
		cache := &SemanticCache[string, string]{}

		embeddings := [][]float64{
			{0.2, 0.4, 0.6},
			{0.4, 0.6, 0.8},
			{0.6, 0.8, 1.0},
		}
		result := cache.aggregateEmbeddings(embeddings)

		if len(result) != 3 {
			t.Fatalf("Expected length 3, got %d", len(result))
		}

		// Expected: average of each dimension
		expected := []float64{0.4, 0.6, 0.8}
		epsilon := 0.0001
		for i, v := range result {
			diff := v - expected[i]
			if diff < 0 {
				diff = -diff
			}
			if diff > epsilon {
				t.Errorf("Expected %f at index %d, got %f", expected[i], i, v)
			}
		}
	})
}
