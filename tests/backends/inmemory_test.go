package backends_test

import (
	"context"
	"testing"

	"github.com/botirk38/semanticcache/backends"
	"github.com/botirk38/semanticcache/types"
)

// Test suite for all in-memory backends
func TestInMemoryBackends(t *testing.T) {
	backendTypes := []struct {
		name        string
		backendType types.BackendType
	}{
		{"LRU", types.BackendLRU},
		{"FIFO", types.BackendFIFO},
		{"LFU", types.BackendLFU},
	}

	for _, bt := range backendTypes {
		t.Run(bt.name, func(t *testing.T) {
			config := types.BackendConfig{Capacity: 3}
			factory := &backends.BackendFactory[string, string]{}
			backend, err := factory.NewBackend(bt.backendType, config)
			if err != nil {
				t.Fatalf("Failed to create %s backend: %v", bt.name, err)
			}
			defer func() { _ = backend.Close() }()

			testBasicOperations(t, backend)
			testCapacityLimits(t, backend, 3)
		})
	}
}

func testBasicOperations(t *testing.T, backend types.CacheBackend[string, string]) {
	ctx := context.Background()

	// Test initial state
	if len, _ := backend.Len(ctx); len != 0 {
		t.Errorf("Expected empty backend, got length %d", len)
	}

	// Test Set and Get
	entry1 := types.Entry[string]{
		Embedding: []float32{0.1, 0.2, 0.3},
		Value:     "value1",
	}

	err := backend.Set(ctx, "key1", entry1)
	if err != nil {
		t.Fatalf("Failed to set entry: %v", err)
	}

	// Test Get
	retrieved, found, err := backend.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if !found {
		t.Error("Expected to find key1")
	}
	if retrieved.Value != "value1" {
		t.Errorf("Expected value1, got %s", retrieved.Value)
	}

	// Test Contains
	exists, err := backend.Contains(ctx, "key1")
	if err != nil {
		t.Fatalf("Failed to check contains: %v", err)
	}
	if !exists {
		t.Error("Expected key1 to exist")
	}

	// Test GetEmbedding
	embedding, found, err := backend.GetEmbedding(ctx, "key1")
	if err != nil {
		t.Fatalf("Failed to get embedding: %v", err)
	}
	if !found {
		t.Error("Expected to find embedding for key1")
	}
	if len(embedding) != 3 || embedding[0] != 0.1 {
		t.Errorf("Expected embedding [0.1, 0.2, 0.3], got %v", embedding)
	}

	// Test Keys
	keys, err := backend.Keys(ctx)
	if err != nil {
		t.Fatalf("Failed to get keys: %v", err)
	}
	if len(keys) != 1 || keys[0] != "key1" {
		t.Errorf("Expected [key1], got %v", keys)
	}

	// Test Len
	if len, _ := backend.Len(ctx); len != 1 {
		t.Errorf("Expected length 1, got %d", len)
	}

	// Test Delete
	err = backend.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("Failed to delete entry: %v", err)
	}

	// Verify deletion
	exists, _ = backend.Contains(ctx, "key1")
	if exists {
		t.Error("Expected key1 to be deleted")
	}

	// Test Flush
	_ = backend.Set(ctx, "key2", entry1)
	err = backend.Flush(ctx)
	if err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	if len, _ := backend.Len(ctx); len != 0 {
		t.Errorf("Expected empty backend after flush, got length %d", len)
	}
}

func testCapacityLimits(t *testing.T, backend types.CacheBackend[string, string], capacity int) {
	ctx := context.Background()

	// Fill beyond capacity
	for i := 0; i < capacity+2; i++ {
		entry := types.Entry[string]{
			Embedding: []float32{float32(i), 0.2, 0.3},
			Value:     "value" + string(rune(i+'0')),
		}
		key := "key" + string(rune(i+'0'))
		err := backend.Set(ctx, key, entry)
		if err != nil {
			t.Fatalf("Failed to set entry %s: %v", key, err)
		}
	}

	// Check that capacity is respected
	length, err := backend.Len(ctx)
	if err != nil {
		t.Fatalf("Failed to get length: %v", err)
	}
	if length > capacity {
		t.Errorf("Expected length <= %d, got %d", capacity, length)
	}
}

func TestBackendErrorCases(t *testing.T) {
	factory := &backends.BackendFactory[string, string]{}

	// Test unsupported backend type
	_, err := factory.NewBackend("unsupported", types.BackendConfig{})
	if err == nil {
		t.Error("Expected error for unsupported backend type")
	}
}

func TestConcurrentAccess(t *testing.T) {
	config := types.BackendConfig{Capacity: 10}
	factory := &backends.BackendFactory[string, string]{}
	backend, err := factory.NewBackend(types.BackendLRU, config)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}
	defer func() { _ = backend.Close() }()

	ctx := context.Background()

	// Test concurrent writes and reads
	done := make(chan bool, 20)

	// Start 10 writers
	for i := range 10 {
		go func(id int) {
			entry := types.Entry[string]{
				Embedding: []float32{float32(id), 0.2, 0.3},
				Value:     "value" + string(rune(id+'0')),
			}
			key := "key" + string(rune(id+'0'))
			_ = backend.Set(ctx, key, entry)
			done <- true
		}(i)
	}

	// Start 10 readers
	for i := range 10 {
		go func(id int) {
			key := "key" + string(rune(id+'0'))
			_, _, _ = backend.Get(ctx, key)
			_, _ = backend.Contains(ctx, key)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range 20 {
		<-done
	}

	// Verify final state is consistent
	length, err := backend.Len(ctx)
	if err != nil {
		t.Fatalf("Failed to get final length: %v", err)
	}
	if length < 0 || length > 10 {
		t.Errorf("Expected length between 0-10, got %d", length)
	}
}
