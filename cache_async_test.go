package semanticcache

import (
	"context"
	"testing"

	"github.com/botirk38/semanticcache/options"
)

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
