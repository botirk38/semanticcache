package inmemory

import (
	"context"
	"sync"

	"github.com/botirk38/semanticcache/types"
)

// LFUEntry wraps an entry with frequency tracking
type LFUEntry[V any] struct {
	Entry     types.Entry[V]
	Frequency int
}

// LFUBackend implements CacheBackend using LFU (Least Frequently Used) eviction policy
type LFUBackend[K comparable, V any] struct {
	mu       *sync.RWMutex
	entries  map[K]*LFUEntry[V]
	index    map[K][]float32
	capacity int
}

// NewLFUBackend creates a new LFU backend
func NewLFUBackend[K comparable, V any](config types.BackendConfig) (*LFUBackend[K, V], error) {
	return &LFUBackend[K, V]{
		mu:       &sync.RWMutex{},
		entries:  make(map[K]*LFUEntry[V]),
		index:    make(map[K][]float32),
		capacity: config.Capacity,
	}, nil
}

// Set stores an entry in the LFU cache
func (b *LFUBackend[K, V]) Set(ctx context.Context, key K, entry types.Entry[V]) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If key already exists, update it and increment frequency
	if existingEntry, exists := b.entries[key]; exists {
		existingEntry.Entry = entry
		existingEntry.Frequency++
		b.index[key] = entry.Embedding
		return nil
	}

	// If at capacity, evict the least frequently used entry
	if len(b.entries) >= b.capacity && b.capacity > 0 {
		b.evictLFU()
	}

	// Add new entry with frequency 1
	b.entries[key] = &LFUEntry[V]{
		Entry:     entry,
		Frequency: 1,
	}
	b.index[key] = entry.Embedding
	return nil
}

// evictLFU removes the least frequently used entry
func (b *LFUBackend[K, V]) evictLFU() {
	var lfuKey K
	minFreq := int(^uint(0) >> 1) // Max int value

	for key, entry := range b.entries {
		if entry.Frequency < minFreq {
			minFreq = entry.Frequency
			lfuKey = key
		}
	}

	delete(b.entries, lfuKey)
	delete(b.index, lfuKey)
}

// Get retrieves an entry from the LFU cache and increments its frequency
func (b *LFUBackend[K, V]) Get(ctx context.Context, key K) (types.Entry[V], bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if entry, ok := b.entries[key]; ok {
		entry.Frequency++
		return entry.Entry, true, nil
	}
	return types.Entry[V]{}, false, nil
}

// Delete removes an entry from the LFU cache
func (b *LFUBackend[K, V]) Delete(ctx context.Context, key K) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.entries, key)
	delete(b.index, key)
	return nil
}

// Contains checks if a key exists in the LFU cache (without incrementing frequency)
func (b *LFUBackend[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, exists := b.entries[key]
	return exists, nil
}

// Flush clears all entries from the LFU cache
func (b *LFUBackend[K, V]) Flush(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries = make(map[K]*LFUEntry[V])
	b.index = make(map[K][]float32)
	return nil
}

// Len returns the number of entries in the LFU cache
func (b *LFUBackend[K, V]) Len(ctx context.Context) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.entries), nil
}

// Keys returns all keys in the LFU cache
func (b *LFUBackend[K, V]) Keys(ctx context.Context) ([]K, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	keys := make([]K, 0, len(b.index))
	for key := range b.index {
		keys = append(keys, key)
	}
	return keys, nil
}

// GetEmbedding retrieves just the embedding for a key (without incrementing frequency)
func (b *LFUBackend[K, V]) GetEmbedding(ctx context.Context, key K) ([]float32, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if embedding, ok := b.index[key]; ok {
		return embedding, true, nil
	}
	return nil, false, nil
}

// Close closes the LFU backend (no-op for in-memory)
func (b *LFUBackend[K, V]) Close() error {
	return nil
}
