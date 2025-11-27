package inmemory

import (
	"context"
	"sync"

	"github.com/botirk38/semanticcache/types"
	lru "github.com/hashicorp/golang-lru/v2"
)

// LRUBackend implements CacheBackend using LRU eviction policy
type LRUBackend[K comparable, V any] struct {
	mu    *sync.RWMutex
	cache *lru.Cache[K, types.Entry[V]]
	index map[K][]float64
}

// NewLRUBackend creates a new LRU backend
func NewLRUBackend[K comparable, V any](config types.BackendConfig) (*LRUBackend[K, V], error) {
	lruCache, err := lru.New[K, types.Entry[V]](config.Capacity)
	if err != nil {
		return nil, err
	}

	return &LRUBackend[K, V]{
		mu:    &sync.RWMutex{},
		cache: lruCache,
		index: make(map[K][]float64),
	}, nil
}

// Set stores an entry in the LRU cache
func (b *LRUBackend[K, V]) Set(ctx context.Context, key K, entry types.Entry[V]) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cache.Add(key, entry)
	b.index[key] = entry.Embedding
	return nil
}

// Get retrieves an entry from the LRU cache
func (b *LRUBackend[K, V]) Get(ctx context.Context, key K) (types.Entry[V], bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if entry, ok := b.cache.Get(key); ok {
		return entry, true, nil
	}
	return types.Entry[V]{}, false, nil
}

// Delete removes an entry from the LRU cache
func (b *LRUBackend[K, V]) Delete(ctx context.Context, key K) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cache.Remove(key)
	delete(b.index, key)
	return nil
}

// Contains checks if a key exists in the LRU cache
func (b *LRUBackend[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.cache.Contains(key), nil
}

// Flush clears all entries from the LRU cache
func (b *LRUBackend[K, V]) Flush(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cache.Purge()
	b.index = make(map[K][]float64)
	return nil
}

// Len returns the number of entries in the LRU cache
func (b *LRUBackend[K, V]) Len(ctx context.Context) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.cache.Len(), nil
}

// Keys returns all keys in the LRU cache
func (b *LRUBackend[K, V]) Keys(ctx context.Context) ([]K, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clean up stale index entries and collect valid keys
	keys := make([]K, 0, b.cache.Len())
	validIndex := make(map[K][]float64)

	for key, embedding := range b.index {
		if b.cache.Contains(key) {
			keys = append(keys, key)
			validIndex[key] = embedding
		}
	}

	// Replace the index with the cleaned version
	b.index = validIndex
	return keys, nil
}

// GetEmbedding retrieves just the embedding for a key
func (b *LRUBackend[K, V]) GetEmbedding(ctx context.Context, key K) ([]float64, bool, error) {
	b.mu.RLock()
	embedding, hasEmbedding := b.index[key]
	cacheHasKey := b.cache.Contains(key)
	b.mu.RUnlock()

	if hasEmbedding && cacheHasKey {
		return embedding, true, nil
	}

	// Clean up stale index entry if needed
	if hasEmbedding && !cacheHasKey {
		b.mu.Lock()
		delete(b.index, key)
		b.mu.Unlock()
	}

	return nil, false, nil
}

// Close closes the LRU backend (no-op for in-memory)
func (b *LRUBackend[K, V]) Close() error {
	return nil
}

// SetAsync stores an entry asynchronously
func (b *LRUBackend[K, V]) SetAsync(ctx context.Context, key K, entry types.Entry[V]) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- b.Set(ctx, key, entry)
	}()
	return errCh
}

// GetAsync retrieves an entry asynchronously
func (b *LRUBackend[K, V]) GetAsync(ctx context.Context, key K) <-chan types.AsyncGetResult[V] {
	resultCh := make(chan types.AsyncGetResult[V], 1)
	go func() {
		defer close(resultCh)
		entry, found, err := b.Get(ctx, key)
		resultCh <- types.AsyncGetResult[V]{
			Entry: entry,
			Found: found,
			Error: err,
		}
	}()
	return resultCh
}

// DeleteAsync removes an entry asynchronously
func (b *LRUBackend[K, V]) DeleteAsync(ctx context.Context, key K) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- b.Delete(ctx, key)
	}()
	return errCh
}

// GetBatchAsync retrieves multiple entries asynchronously
func (b *LRUBackend[K, V]) GetBatchAsync(ctx context.Context, keys []K) <-chan types.AsyncBatchResult[K, V] {
	resultCh := make(chan types.AsyncBatchResult[K, V], 1)
	go func() {
		defer close(resultCh)
		entries := make(map[K]types.Entry[V])
		for _, key := range keys {
			if entry, found, err := b.Get(ctx, key); err == nil && found {
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
