package types

import (
	"context"
	"time"
)

// Entry holds an embedding and its associated value.
type Entry[V any] struct {
	Embedding []float32
	Value     V
}

// CacheBackend defines the interface for different cache storage backends.
// This allows for pluggable storage systems including in-memory and Redis.
type CacheBackend[K comparable, V any] interface {
	// Set stores a value with its embedding in the cache
	Set(ctx context.Context, key K, entry Entry[V]) error

	// Get retrieves an entry by key
	Get(ctx context.Context, key K) (Entry[V], bool, error)

	// Delete removes an entry by key
	Delete(ctx context.Context, key K) error

	// Contains checks if a key exists without retrieving the value
	Contains(ctx context.Context, key K) (bool, error)

	// Flush clears all entries from the cache
	Flush(ctx context.Context) error

	// Len returns the number of entries in the cache
	Len(ctx context.Context) (int, error)

	// Keys returns all keys in the cache (for semantic search)
	Keys(ctx context.Context) ([]K, error)

	// GetEmbedding retrieves just the embedding for a key
	GetEmbedding(ctx context.Context, key K) ([]float32, bool, error)

	// Close closes the backend and releases resources
	Close() error

	// Async operations
	// SetAsync stores a value asynchronously
	SetAsync(ctx context.Context, key K, entry Entry[V]) <-chan error

	// GetAsync retrieves an entry asynchronously
	GetAsync(ctx context.Context, key K) <-chan AsyncGetResult[V]

	// DeleteAsync removes an entry asynchronously
	DeleteAsync(ctx context.Context, key K) <-chan error

	// GetBatchAsync retrieves multiple entries asynchronously
	GetBatchAsync(ctx context.Context, keys []K) <-chan AsyncBatchResult[K, V]
}

// AsyncGetResult holds the result of an async Get operation at the backend level.
type AsyncGetResult[V any] struct {
	Entry Entry[V]
	Found bool
	Error error
}

// AsyncBatchResult holds the result of an async batch operation.
type AsyncBatchResult[K comparable, V any] struct {
	Entries map[K]Entry[V]
	Error   error
}

// BackendConfig provides configuration options for backends
type BackendConfig struct {
	// For in-memory caches
	Capacity int
	TTL      time.Duration

	// For Redis
	ConnectionString string
	Username         string
	Password         string
	Database         int

	// Additional options
	Options map[string]any
}

// BackendType represents the type of cache backend
type BackendType string

const (
	BackendLRU   BackendType = "lru"
	BackendFIFO  BackendType = "fifo"
	BackendLFU   BackendType = "lfu"
	BackendRedis BackendType = "redis"
)

// EmbeddingProvider defines the interface all embedding providers must satisfy.
type EmbeddingProvider interface {
	// EmbedText turns a piece of text into its embedding vector.
	EmbedText(text string) ([]float32, error)
	// Close frees any resources held by the provider.
	Close()
}

// ProviderType represents the type of embedding provider
type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	// Add more providers as needed:
	// ProviderAnthropic   ProviderType = "anthropic"
	// ProviderOllama      ProviderType = "ollama"
	// ProviderHuggingFace ProviderType = "huggingface"
	// ProviderCohere      ProviderType = "cohere"
)
