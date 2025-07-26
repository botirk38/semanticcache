package backends_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/botirk38/semanticcache/backends/remote"
	"github.com/botirk38/semanticcache/types"
)

// TestRedisBackend tests Redis backend functionality
// Requires Redis to be running on localhost:6379
func TestRedisBackend(t *testing.T) {
	// Skip if Redis is not available
	if testing.Short() {
		t.Skip("Skipping Redis tests in short mode")
	}

	// Check if Redis connection string is provided
	connStr := os.Getenv("REDIS_URL")
	if connStr == "" {
		connStr = "localhost:6379"
	}

	config := types.BackendConfig{
		ConnectionString: connStr,
		Options: map[string]any{
			"prefix":     "test_cache:",
			"dimensions": 3,
		},
	}

	backend, err := remote.NewRedisBackend[string, string](config)
	if err != nil {
		t.Skipf("Redis not available, skipping Redis tests: %v", err)
	}
	defer func() { _ = backend.Close() }()

	// Clean up before test
	ctx := context.Background()
	_ = backend.Flush(ctx)

	t.Run("BasicOperations", func(t *testing.T) {
		testRedisBasicOperations(t, backend)
	})

	t.Run("VectorSearch", func(t *testing.T) {
		testRedisVectorSearch(t, backend)
	})

	t.Run("Persistence", func(t *testing.T) {
		testRedisPersistence(t, backend, config)
	})
}

func testRedisBasicOperations(t *testing.T, backend *remote.RedisBackend[string, string]) {
	ctx := context.Background()

	// Test Set and Get
	entry := types.Entry[string]{
		Embedding: []float32{0.1, 0.2, 0.3},
		Value:     "redis_value",
	}

	err := backend.Set(ctx, "redis_key", entry)
	if err != nil {
		t.Fatalf("Failed to set entry: %v", err)
	}

	// Test Get
	retrieved, found, err := backend.Get(ctx, "redis_key")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if !found {
		t.Error("Expected to find redis_key")
	}
	if retrieved.Value != "redis_value" {
		t.Errorf("Expected redis_value, got %s", retrieved.Value)
	}

	// Test embedding retrieval
	if len(retrieved.Embedding) != 3 {
		t.Errorf("Expected 3D embedding, got %d dimensions", len(retrieved.Embedding))
	}

	// Test Contains
	exists, err := backend.Contains(ctx, "redis_key")
	if err != nil {
		t.Fatalf("Failed to check contains: %v", err)
	}
	if !exists {
		t.Error("Expected redis_key to exist")
	}

	// Test GetEmbedding
	embedding, found, err := backend.GetEmbedding(ctx, "redis_key")
	if err != nil {
		t.Fatalf("Failed to get embedding: %v", err)
	}
	if !found {
		t.Error("Expected to find embedding")
	}
	if len(embedding) != 3 {
		t.Errorf("Expected 3D embedding, got %v", embedding)
	}

	// Test Delete
	err = backend.Delete(ctx, "redis_key")
	if err != nil {
		t.Fatalf("Failed to delete entry: %v", err)
	}

	// Verify deletion
	exists, _ = backend.Contains(ctx, "redis_key")
	if exists {
		t.Error("Expected redis_key to be deleted")
	}
}

func testRedisVectorSearch(t *testing.T, backend *remote.RedisBackend[string, string]) {
	ctx := context.Background()

	// Set up test data with known embeddings
	testData := []struct {
		key       string
		embedding []float32
		value     string
	}{
		{"vec1", []float32{1.0, 0.0, 0.0}, "first"},
		{"vec2", []float32{0.0, 1.0, 0.0}, "second"},
		{"vec3", []float32{0.0, 0.0, 1.0}, "third"},
		{"vec4", []float32{0.7, 0.7, 0.0}, "fourth"}, // Similar to vec1 and vec2
	}

	// Insert test data
	for _, data := range testData {
		entry := types.Entry[string]{
			Embedding: data.embedding,
			Value:     data.value,
		}
		err := backend.Set(ctx, data.key, entry)
		if err != nil {
			t.Fatalf("Failed to set test data %s: %v", data.key, err)
		}
	}

	// Wait a bit for Redis to index
	time.Sleep(100 * time.Millisecond)

	// Test vector search
	queryEmbedding := []float32{0.9, 0.1, 0.0} // Should be closest to vec1
	results, err := backend.VectorSearch(ctx, queryEmbedding, 0.5, 2)
	if err != nil {
		t.Fatalf("Vector search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result from vector search")
	}

	// Clean up
	for _, data := range testData {
		_ = backend.Delete(ctx, data.key)
	}
}

func testRedisPersistence(t *testing.T, backend *remote.RedisBackend[string, string], config types.BackendConfig) {
	ctx := context.Background()

	// Set data
	entry := types.Entry[string]{
		Embedding: []float32{0.5, 0.5, 0.5},
		Value:     "persistent_value",
	}
	err := backend.Set(ctx, "persist_key", entry)
	if err != nil {
		t.Fatalf("Failed to set persistent entry: %v", err)
	}

	// Create new backend instance (simulating restart)
	newBackend, err := remote.NewRedisBackend[string, string](config)
	if err != nil {
		t.Fatalf("Failed to create new backend: %v", err)
	}
	defer func() { _ = newBackend.Close() }()

	// Check if data persisted
	retrieved, found, err := newBackend.Get(ctx, "persist_key")
	if err != nil {
		t.Fatalf("Failed to get persistent entry: %v", err)
	}
	if !found {
		t.Error("Expected persistent data to be found")
	}
	if retrieved.Value != "persistent_value" {
		t.Errorf("Expected persistent_value, got %s", retrieved.Value)
	}

	// Clean up
	_ = newBackend.Delete(ctx, "persist_key")
}

func TestRedisConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      types.BackendConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: types.BackendConfig{
				ConnectionString: "localhost:6379",
				Options: map[string]any{
					"prefix":     "test:",
					"dimensions": 128,
				},
			},
			expectError: true, // Expect error if no Redis server is running
		},
		{
			name: "invalid connection",
			config: types.BackendConfig{
				ConnectionString: "invalid:99999",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := remote.NewRedisBackend[string, string](tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					if backend != nil {
						_ = backend.Close()
					}
				}
			} else {
				if err != nil && !testing.Short() {
					t.Errorf("Expected no error but got: %v", err)
				}
				if backend != nil {
					_ = backend.Close()
				}
			}
		})
	}
}
