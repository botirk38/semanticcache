// Package types defines the core interfaces and types for the semantic cache.
package types

import "context"

// Entry holds an embedding vector alongside its cached value.
type Entry[V any] struct {
	Embedding []float64
	Value     V
}

// Backend is the storage interface that every cache backend must implement.
type Backend[K comparable, V any] interface {
	// Set stores a value with its embedding vector.
	Set(ctx context.Context, key K, embedding []float64, value V) error

	// Get retrieves the value for a key.
	Get(ctx context.Context, key K) (V, bool, error)

	// Delete removes an entry by key.
	Delete(ctx context.Context, key K) error

	// Contains checks whether a key exists without retrieving the value.
	Contains(ctx context.Context, key K) (bool, error)

	// Keys returns all keys currently stored.
	Keys(ctx context.Context) ([]K, error)

	// GetEmbedding retrieves the embedding vector for a key.
	GetEmbedding(ctx context.Context, key K) ([]float64, bool, error)

	// Flush removes all entries.
	Flush(ctx context.Context) error

	// Len returns the number of stored entries.
	Len(ctx context.Context) (int, error)

	// Close releases any resources held by the backend.
	Close() error
}

// VectorSearchResult is a single hit returned by a VectorSearcher.
type VectorSearchResult[K comparable, V any] struct {
	Key   K
	Value V
	Score float64
}

// VectorSearcher performs server-side similarity search. Backends like
// Redis that have native vector-search capabilities implement this
// in addition to Backend. When present, Lookup and TopMatches use it
// instead of scanning all keys client-side.
type VectorSearcher[K comparable, V any] interface {
	// VectorSearch returns up to limit results whose similarity to query
	// meets or exceeds threshold, sorted by descending score.
	VectorSearch(ctx context.Context, query []float64, threshold float64, limit int) ([]VectorSearchResult[K, V], error)
}

// EmbeddingProvider turns text into embedding vectors.
type EmbeddingProvider interface {
	// EmbedText computes the embedding vector for a single piece of text.
	EmbedText(ctx context.Context, text string) ([]float64, error)

	// Close releases any resources held by the provider.
	Close() error
}

// BatchEmbeddingProvider is an optional extension for providers that
// support embedding multiple texts in a single API call.
type BatchEmbeddingProvider interface {
	EmbeddingProvider

	// EmbedBatch embeds multiple texts in one operation.
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}
