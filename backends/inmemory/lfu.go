package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/botirk38/semanticcache/types"
)

// LFUEntry wraps an entry with frequency tracking
type LFUEntry[V any] struct {
	Entry     types.Entry[V]
	Frequency int
	ExpiresAt time.Time // zero means no expiry
}

// LFUBackend implements CacheBackend using LFU (Least Frequently Used) eviction policy
type LFUBackend[K comparable, V any] struct {
	mu       *sync.RWMutex
	entries  map[K]*LFUEntry[V]
	index    map[K][]float64
	capacity int
	ttl      time.Duration
}

// NewLFUBackend creates a new LFU backend
func NewLFUBackend[K comparable, V any](config types.BackendConfig) (*LFUBackend[K, V], error) {
	return &LFUBackend[K, V]{
		mu:       &sync.RWMutex{},
		entries:  make(map[K]*LFUEntry[V]),
		index:    make(map[K][]float64),
		capacity: config.Capacity,
		ttl:      config.TTL,
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
	lfuEntry := &LFUEntry[V]{
		Entry:     entry,
		Frequency: 1,
	}
	if b.ttl > 0 {
		lfuEntry.ExpiresAt = time.Now().Add(b.ttl)
	}
	b.entries[key] = lfuEntry
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
		if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
			delete(b.entries, key)
			delete(b.index, key)
			return types.Entry[V]{}, false, nil
		}
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
	b.mu.Lock()
	defer b.mu.Unlock()

	entry, exists := b.entries[key]
	if exists && !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		delete(b.entries, key)
		delete(b.index, key)
		return false, nil
	}
	return exists, nil
}

// Flush clears all entries from the LFU cache
func (b *LFUBackend[K, V]) Flush(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries = make(map[K]*LFUEntry[V])
	b.index = make(map[K][]float64)
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
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	keys := make([]K, 0, len(b.index))
	for key := range b.index {
		if entry, ok := b.entries[key]; ok {
			if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
				delete(b.entries, key)
				delete(b.index, key)
				continue
			}
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// GetEmbedding retrieves just the embedding for a key (without incrementing frequency)
func (b *LFUBackend[K, V]) GetEmbedding(ctx context.Context, key K) ([]float64, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if entry, ok := b.entries[key]; ok {
		if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
			delete(b.entries, key)
			delete(b.index, key)
			return nil, false, nil
		}
	}
	if embedding, ok := b.index[key]; ok {
		return embedding, true, nil
	}
	return nil, false, nil
}

// Close closes the LFU backend (no-op for in-memory)
func (b *LFUBackend[K, V]) Close() error {
	return nil
}

// SetAsync stores an entry asynchronously
func (b *LFUBackend[K, V]) SetAsync(ctx context.Context, key K, entry types.Entry[V]) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- b.Set(ctx, key, entry)
	}()
	return errCh
}

// GetAsync retrieves an entry asynchronously
func (b *LFUBackend[K, V]) GetAsync(ctx context.Context, key K) <-chan types.AsyncGetResult[V] {
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
func (b *LFUBackend[K, V]) DeleteAsync(ctx context.Context, key K) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		errCh <- b.Delete(ctx, key)
	}()
	return errCh
}

// GetBatchAsync retrieves multiple entries asynchronously
func (b *LFUBackend[K, V]) GetBatchAsync(ctx context.Context, keys []K) <-chan types.AsyncBatchResult[K, V] {
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
