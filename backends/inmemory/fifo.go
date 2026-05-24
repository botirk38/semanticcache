package inmemory

import (
	"context"
	"sync"

	"github.com/botirk38/semanticcache/types"
)

// FIFOBackend implements Backend and EmbeddingStore using FIFO eviction.
type FIFOBackend[K comparable, V any] struct {
	mu       sync.RWMutex
	entries  map[K]types.Entry[V]
	queue    []K
	capacity int
}

// NewFIFOBackend creates a new FIFO backend with the given capacity.
func NewFIFOBackend[K comparable, V any](capacity int) (*FIFOBackend[K, V], error) {
	return &FIFOBackend[K, V]{
		entries:  make(map[K]types.Entry[V]),
		queue:    make([]K, 0, capacity),
		capacity: capacity,
	}, nil
}

// Set stores a value with its embedding.
func (b *FIFOBackend[K, V]) Set(_ context.Context, key K, embedding []float64, value V) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	entry := types.Entry[V]{Embedding: embedding, Value: value}

	if _, ok := b.entries[key]; ok {
		b.entries[key] = entry
		return nil
	}

	if len(b.entries) >= b.capacity && b.capacity > 0 {
		oldest := b.queue[0]
		b.queue = b.queue[1:]
		delete(b.entries, oldest)
	}

	b.entries[key] = entry
	b.queue = append(b.queue, key)
	return nil
}

// Get retrieves the value for a key.
func (b *FIFOBackend[K, V]) Get(_ context.Context, key K) (V, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if e, ok := b.entries[key]; ok {
		return e.Value, true, nil
	}
	var zero V
	return zero, false, nil
}

// Delete removes an entry by key.
func (b *FIFOBackend[K, V]) Delete(_ context.Context, key K) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.entries[key]; !ok {
		return nil
	}
	delete(b.entries, key)

	for i, k := range b.queue {
		if k == key {
			b.queue = append(b.queue[:i], b.queue[i+1:]...)
			break
		}
	}
	return nil
}

// Contains checks whether a key exists.
func (b *FIFOBackend[K, V]) Contains(_ context.Context, key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.entries[key]
	return ok, nil
}

// Flush removes all entries.
func (b *FIFOBackend[K, V]) Flush(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = make(map[K]types.Entry[V])
	b.queue = make([]K, 0, b.capacity)
	return nil
}

// Len returns the number of stored entries.
func (b *FIFOBackend[K, V]) Len(_ context.Context) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.entries), nil
}

// Close is a no-op for in-memory backends.
func (b *FIFOBackend[K, V]) Close() error { return nil }

// Keys returns all keys in the cache.
func (b *FIFOBackend[K, V]) Keys(_ context.Context) ([]K, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	keys := make([]K, 0, len(b.entries))
	for k := range b.entries {
		keys = append(keys, k)
	}
	return keys, nil
}

// GetEmbedding retrieves the embedding for a key.
func (b *FIFOBackend[K, V]) GetEmbedding(_ context.Context, key K) ([]float64, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if e, ok := b.entries[key]; ok {
		return e.Embedding, true, nil
	}
	return nil, false, nil
}
