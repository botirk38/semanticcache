package semanticcache

import (
	"context"
	"sync"
	"testing"

	"github.com/botirk38/semanticcache/options"
	"github.com/botirk38/semanticcache/types"
)

// syncOnlyBackend implements only CacheBackend (no AsyncCacheBackend).
// Used to test that async operations fall back to wrapping sync methods.
type syncOnlyBackend[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]types.Entry[V]
}

func newSyncOnlyBackend[K comparable, V any]() *syncOnlyBackend[K, V] {
	return &syncOnlyBackend[K, V]{data: make(map[K]types.Entry[V])}
}

func (b *syncOnlyBackend[K, V]) Set(_ context.Context, key K, entry types.Entry[V]) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[key] = entry
	return nil
}
func (b *syncOnlyBackend[K, V]) Get(_ context.Context, key K) (types.Entry[V], bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	e, ok := b.data[key]
	return e, ok, nil
}
func (b *syncOnlyBackend[K, V]) Delete(_ context.Context, key K) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.data, key)
	return nil
}
func (b *syncOnlyBackend[K, V]) Contains(_ context.Context, key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.data[key]
	return ok, nil
}
func (b *syncOnlyBackend[K, V]) Flush(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = make(map[K]types.Entry[V])
	return nil
}
func (b *syncOnlyBackend[K, V]) Len(_ context.Context) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.data), nil
}
func (b *syncOnlyBackend[K, V]) Keys(_ context.Context) ([]K, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	keys := make([]K, 0, len(b.data))
	for k := range b.data {
		keys = append(keys, k)
	}
	return keys, nil
}
func (b *syncOnlyBackend[K, V]) GetEmbedding(_ context.Context, key K) ([]float64, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	e, ok := b.data[key]
	if !ok {
		return nil, false, nil
	}
	return e.Embedding, true, nil
}
func (b *syncOnlyBackend[K, V]) Close() error { return nil }

func TestSyncOnlyBackendAsyncFallback(t *testing.T) {
	ctx := context.Background()

	cache, err := New(
		options.WithCustomBackend[string, string](newSyncOnlyBackend[string, string]()),
		options.WithCustomProvider[string, string](newMockProvider()),
	)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	t.Run("SetAsync", func(t *testing.T) {
		errCh := cache.SetAsync(ctx, "k1", "hello", "v1")
		if err := <-errCh; err != nil {
			t.Fatalf("SetAsync failed: %v", err)
		}
		val, found, err := cache.Get(ctx, "k1")
		if err != nil || !found || val != "v1" {
			t.Fatalf("Expected v1, got %s (found=%v, err=%v)", val, found, err)
		}
	})

	t.Run("GetAsync", func(t *testing.T) {
		resultCh := cache.GetAsync(ctx, "k1")
		result := <-resultCh
		if result.Error != nil || !result.Found || result.Value != "v1" {
			t.Fatalf("GetAsync failed: %+v", result)
		}
	})

	t.Run("DeleteAsync", func(t *testing.T) {
		errCh := cache.DeleteAsync(ctx, "k1")
		if err := <-errCh; err != nil {
			t.Fatalf("DeleteAsync failed: %v", err)
		}
		_, found, _ := cache.Get(ctx, "k1")
		if found {
			t.Fatal("Expected key to be deleted")
		}
	})

	t.Run("GetBatchAsync", func(t *testing.T) {
		_ = cache.Set(ctx, "a", "hello", "va")
		_ = cache.Set(ctx, "b", "world", "vb")
		resultCh := cache.GetBatchAsync(ctx, []string{"a", "b", "missing"})
		result := <-resultCh
		if result.Error != nil {
			t.Fatalf("GetBatchAsync failed: %v", result.Error)
		}
		if len(result.Values) != 2 {
			t.Fatalf("Expected 2 values, got %d", len(result.Values))
		}
	})

	t.Run("DeleteBatchAsync", func(t *testing.T) {
		errCh := cache.DeleteBatchAsync(ctx, []string{"a", "b"})
		if err := <-errCh; err != nil {
			t.Fatalf("DeleteBatchAsync failed: %v", err)
		}
		l, _ := cache.Len(ctx)
		if l != 0 {
			t.Fatalf("Expected 0 items, got %d", l)
		}
	})
}

func TestAsyncOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("SetAsync", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		errCh := cache.SetAsync(ctx, "key1", "hello", "value1")
		err = <-errCh
		if err != nil {
			t.Errorf("SetAsync failed: %v", err)
		}

		// Verify the value was set
		value, found, err := cache.Get(ctx, "key1")
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if !found {
			t.Error("Expected key to be found")
		}
		if value != "value1" {
			t.Errorf("Expected value1, got %s", value)
		}
	})

	t.Run("GetAsync", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Set a value first
		err = cache.Set(ctx, "key1", "hello", "value1")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Get async
		resultCh := cache.GetAsync(ctx, "key1")
		result := <-resultCh

		if result.Error != nil {
			t.Errorf("GetAsync failed: %v", result.Error)
		}
		if !result.Found {
			t.Error("Expected key to be found")
		}
		if result.Value != "value1" {
			t.Errorf("Expected value1, got %s", result.Value)
		}
	})

	t.Run("GetAsync_NotFound", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		resultCh := cache.GetAsync(ctx, "nonexistent")
		result := <-resultCh

		if result.Error != nil {
			t.Errorf("GetAsync failed: %v", result.Error)
		}
		if result.Found {
			t.Error("Expected key not to be found")
		}
	})

	t.Run("DeleteAsync", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Set a value first
		err = cache.Set(ctx, "key1", "hello", "value1")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Delete async
		errCh := cache.DeleteAsync(ctx, "key1")
		err = <-errCh
		if err != nil {
			t.Errorf("DeleteAsync failed: %v", err)
		}

		// Verify deletion
		_, found, err := cache.Get(ctx, "key1")
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if found {
			t.Error("Expected key to be deleted")
		}
	})

	t.Run("LookupAsync", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Set a value
		err = cache.Set(ctx, "key1", "hello", "value1")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Lookup async with similar text
		resultCh := cache.LookupAsync(ctx, "similar to hello", 0.5)
		result := <-resultCh

		if result.Error != nil {
			t.Errorf("LookupAsync failed: %v", result.Error)
		}
		if result.Match == nil {
			t.Error("Expected to find a match")
		}
	})

	t.Run("TopMatchesAsync", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Set multiple values
		err = cache.Set(ctx, "key1", "hello", "value1")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		err = cache.Set(ctx, "key2", "world", "value2")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Get top matches async
		resultCh := cache.TopMatchesAsync(ctx, "test", 2)
		result := <-resultCh

		if result.Error != nil {
			t.Errorf("TopMatchesAsync failed: %v", result.Error)
		}
		if len(result.Matches) == 0 {
			t.Error("Expected to find matches")
		}
	})

	t.Run("SetBatchAsync", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		items := []BatchItem[string, string]{
			{Key: "key1", InputText: "hello", Value: "value1"},
			{Key: "key2", InputText: "world", Value: "value2"},
		}

		errCh := cache.SetBatchAsync(ctx, items)
		err = <-errCh
		if err != nil {
			t.Errorf("SetBatchAsync failed: %v", err)
		}

		// Verify both values were set
		for i, item := range items {
			value, found, err := cache.Get(ctx, item.Key)
			if err != nil {
				t.Errorf("Get failed for item %d: %v", i, err)
			}
			if !found {
				t.Errorf("Expected key %s to be found", item.Key)
			}
			if value != item.Value {
				t.Errorf("Expected %s, got %s", item.Value, value)
			}
		}
	})

	t.Run("GetBatchAsync", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Set values first
		err = cache.Set(ctx, "key1", "hello", "value1")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		err = cache.Set(ctx, "key2", "world", "value2")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Get batch async
		resultCh := cache.GetBatchAsync(ctx, []string{"key1", "key2", "key3"})
		result := <-resultCh

		if result.Error != nil {
			t.Errorf("GetBatchAsync failed: %v", result.Error)
		}
		if len(result.Values) != 2 {
			t.Errorf("Expected 2 values, got %d", len(result.Values))
		}
		if result.Values["key1"] != "value1" {
			t.Errorf("Expected value1, got %s", result.Values["key1"])
		}
		if result.Values["key2"] != "value2" {
			t.Errorf("Expected value2, got %s", result.Values["key2"])
		}
	})

	t.Run("DeleteBatchAsync", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](10),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Set values first
		err = cache.Set(ctx, "key1", "hello", "value1")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}
		err = cache.Set(ctx, "key2", "world", "value2")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Delete batch async
		errCh := cache.DeleteBatchAsync(ctx, []string{"key1", "key2"})
		err = <-errCh
		if err != nil {
			t.Errorf("DeleteBatchAsync failed: %v", err)
		}

		// Verify deletion
		_, found, err := cache.Get(ctx, "key1")
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if found {
			t.Error("Expected key1 to be deleted")
		}

		_, found, err = cache.Get(ctx, "key2")
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if found {
			t.Error("Expected key2 to be deleted")
		}
	})

	t.Run("ConcurrentAsyncOperations", func(t *testing.T) {
		cache, err := New(
			options.WithLRUBackend[string, string](100),
			options.WithCustomProvider[string, string](newMockProvider()),
		)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		// Launch multiple async operations concurrently
		errCh1 := cache.SetAsync(ctx, "key1", "hello", "value1")
		errCh2 := cache.SetAsync(ctx, "key2", "world", "value2")
		errCh3 := cache.SetAsync(ctx, "key3", "test", "value3")

		// Wait for all to complete
		if err := <-errCh1; err != nil {
			t.Errorf("SetAsync 1 failed: %v", err)
		}
		if err := <-errCh2; err != nil {
			t.Errorf("SetAsync 2 failed: %v", err)
		}
		if err := <-errCh3; err != nil {
			t.Errorf("SetAsync 3 failed: %v", err)
		}

		// Verify all values
		len, err := cache.Len(ctx)
		if err != nil {
			t.Errorf("Len failed: %v", err)
		}
		if len != 3 {
			t.Errorf("Expected 3 items, got %d", len)
		}
	})
}
