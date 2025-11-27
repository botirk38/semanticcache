package inmemory

import (
	"context"
	"sync"

	"github.com/botirk38/semanticcache/types"
)

// FIFOBackend implements CacheBackend using FIFO (First In, First Out) eviction policy
type FIFOBackend[K comparable, V any] struct {
	mu       *sync.RWMutex
	entries  map[K]types.Entry[V]
	index    map[K][]float64
	queue    []K
	capacity int
}

// NewFIFOBackend creates a new FIFO backend
func NewFIFOBackend[K comparable, V any](config types.BackendConfig) (*FIFOBackend[K, V], error) {
	return &FIFOBackend[K, V]{
		mu:       &sync.RWMutex{},
		entries:  make(map[K]types.Entry[V]),
		index:    make(map[K][]float64),
		queue:    make([]K, 0, config.Capacity),
		capacity: config.Capacity,
	}, nil
}

// Set stores an entry in the FIFO cache
func (b *FIFOBackend[K, V]) Set(ctx context.Context, key K, entry types.Entry[V]) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If key already exists, update it
	if _, exists := b.entries[key]; exists {
		b.entries[key] = entry
		b.index[key] = entry.Embedding
		return nil
	}

	// If at capacity, evict the oldest entry (FIFO)
	if len(b.entries) >= b.capacity && b.capacity > 0 {
		oldestKey := b.queue[0]
		b.queue = b.queue[1:]
		delete(b.entries, oldestKey)
		delete(b.index, oldestKey)
	}

	// Add new entry
	b.entries[key] = entry
	b.index[key] = entry.Embedding
	b.queue = append(b.queue, key)
	return nil
}

// Get retrieves an entry from the FIFO cache
func (b *FIFOBackend[K, V]) Get(ctx context.Context, key K) (types.Entry[V], bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if entry, ok := b.entries[key]; ok {
		return entry, true, nil
	}
	return types.Entry[V]{}, false, nil
}

// Delete removes an entry from the FIFO cache
func (b *FIFOBackend[K, V]) Delete(ctx context.Context, key K) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.entries[key]; !exists {
		return nil
	}

	delete(b.entries, key)
	delete(b.index, key)

	// Remove from queue
	for i, qKey := range b.queue {
		if qKey == key {
			b.queue = append(b.queue[:i], b.queue[i+1:]...)
			break
		}
	}
	return nil
}

// Contains checks if a key exists in the FIFO cache
func (b *FIFOBackend[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, exists := b.entries[key]
	return exists, nil
}

// Flush clears all entries from the FIFO cache
func (b *FIFOBackend[K, V]) Flush(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries = make(map[K]types.Entry[V])
	b.index = make(map[K][]float64)
	b.queue = make([]K, 0, b.capacity)
	return nil
}

// Len returns the number of entries in the FIFO cache
func (b *FIFOBackend[K, V]) Len(ctx context.Context) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.entries), nil
}

// Keys returns all keys in the FIFO cache
func (b *FIFOBackend[K, V]) Keys(ctx context.Context) ([]K, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	keys := make([]K, 0, len(b.index))
	for key := range b.index {
		keys = append(keys, key)
	}
	return keys, nil
}

// GetEmbedding retrieves just the embedding for a key
func (b *FIFOBackend[K, V]) GetEmbedding(ctx context.Context, key K) ([]float64, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if embedding, ok := b.index[key]; ok {
		return embedding, true, nil
	}
	return nil, false, nil
}

// Close closes the FIFO backend (no-op for in-memory)
func (b *FIFOBackend[K, V]) Close() error {
	return nil
}

// SetAsync stores an entry asynchronously
func (b *FIFOBackend[K, V]) SetAsync(ctx context.Context, key K, entry types.Entry[V]) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- b.Set(ctx, key, entry)
	}()
	return errCh
}

// GetAsync retrieves an entry asynchronously
func (b *FIFOBackend[K, V]) GetAsync(ctx context.Context, key K) <-chan types.AsyncGetResult[V] {
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
func (b *FIFOBackend[K, V]) DeleteAsync(ctx context.Context, key K) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- b.Delete(ctx, key)
	}()
	return errCh
}

// GetBatchAsync retrieves multiple entries asynchronously
func (b *FIFOBackend[K, V]) GetBatchAsync(ctx context.Context, keys []K) <-chan types.AsyncBatchResult[K, V] {
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
