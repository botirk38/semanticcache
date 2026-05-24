package inmemory

import (
	"context"
	"sync"

	"github.com/botirk38/semanticcache/types"
	lru "github.com/hashicorp/golang-lru/v2"
)

// LRUBackend implements Backend using LRU eviction.
type LRUBackend[K comparable, V any] struct {
	mu    sync.RWMutex
	cache *lru.Cache[K, types.Entry[V]]
}

// NewLRUBackend creates a new LRU backend with the given capacity.
func NewLRUBackend[K comparable, V any](capacity int) (*LRUBackend[K, V], error) {
	c, err := lru.New[K, types.Entry[V]](capacity)
	if err != nil {
		return nil, err
	}
	return &LRUBackend[K, V]{cache: c}, nil
}

// Set stores a value with its embedding.
func (b *LRUBackend[K, V]) Set(_ context.Context, key K, embedding []float64, value V) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cache.Add(key, types.Entry[V]{Embedding: embedding, Value: value})
	return nil
}

// Get retrieves the value for a key.
func (b *LRUBackend[K, V]) Get(_ context.Context, key K) (V, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if entry, ok := b.cache.Get(key); ok {
		return entry.Value, true, nil
	}
	var zero V
	return zero, false, nil
}

// Delete removes an entry by key.
func (b *LRUBackend[K, V]) Delete(_ context.Context, key K) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cache.Remove(key)
	return nil
}

// Contains checks whether a key exists.
func (b *LRUBackend[K, V]) Contains(_ context.Context, key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cache.Contains(key), nil
}

// Flush removes all entries.
func (b *LRUBackend[K, V]) Flush(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cache.Purge()
	return nil
}

// Len returns the number of stored entries.
func (b *LRUBackend[K, V]) Len(_ context.Context) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cache.Len(), nil
}

// Close is a no-op for in-memory backends.
func (b *LRUBackend[K, V]) Close() error { return nil }

// Keys returns all keys in the cache.
func (b *LRUBackend[K, V]) Keys(_ context.Context) ([]K, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cache.Keys(), nil
}

// GetEmbedding retrieves the embedding for a key.
func (b *LRUBackend[K, V]) GetEmbedding(_ context.Context, key K) ([]float64, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if entry, ok := b.cache.Peek(key); ok {
		return entry.Embedding, true, nil
	}
	return nil, false, nil
}
