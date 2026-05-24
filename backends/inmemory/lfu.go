package inmemory

import (
	"context"
	"sync"

	"github.com/botirk38/semanticcache/types"
)

type lfuEntry[V any] struct {
	entry     types.Entry[V]
	frequency int
}

// LFUBackend implements Backend using LFU eviction.
type LFUBackend[K comparable, V any] struct {
	mu       sync.RWMutex
	entries  map[K]*lfuEntry[V]
	capacity int
}

// NewLFUBackend creates a new LFU backend with the given capacity.
func NewLFUBackend[K comparable, V any](capacity int) (*LFUBackend[K, V], error) {
	return &LFUBackend[K, V]{
		entries:  make(map[K]*lfuEntry[V]),
		capacity: capacity,
	}, nil
}

// Set stores a value with its embedding.
func (b *LFUBackend[K, V]) Set(_ context.Context, key K, embedding []float64, value V) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if e, ok := b.entries[key]; ok {
		e.entry = types.Entry[V]{Embedding: embedding, Value: value}
		e.frequency++
		return nil
	}

	if len(b.entries) >= b.capacity && b.capacity > 0 {
		b.evict()
	}

	b.entries[key] = &lfuEntry[V]{
		entry:     types.Entry[V]{Embedding: embedding, Value: value},
		frequency: 1,
	}
	return nil
}

func (b *LFUBackend[K, V]) evict() {
	var victim K
	minFreq := int(^uint(0) >> 1)
	for k, e := range b.entries {
		if e.frequency < minFreq {
			minFreq = e.frequency
			victim = k
		}
	}
	delete(b.entries, victim)
}

// Get retrieves the value for a key and increments its frequency.
func (b *LFUBackend[K, V]) Get(_ context.Context, key K) (V, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if e, ok := b.entries[key]; ok {
		e.frequency++
		return e.entry.Value, true, nil
	}
	var zero V
	return zero, false, nil
}

// Delete removes an entry by key.
func (b *LFUBackend[K, V]) Delete(_ context.Context, key K) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.entries, key)
	return nil
}

// Contains checks whether a key exists without incrementing frequency.
func (b *LFUBackend[K, V]) Contains(_ context.Context, key K) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.entries[key]
	return ok, nil
}

// Flush removes all entries.
func (b *LFUBackend[K, V]) Flush(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = make(map[K]*lfuEntry[V])
	return nil
}

// Len returns the number of stored entries.
func (b *LFUBackend[K, V]) Len(_ context.Context) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.entries), nil
}

// Close is a no-op for in-memory backends.
func (b *LFUBackend[K, V]) Close() error { return nil }

// Keys returns all keys in the cache.
func (b *LFUBackend[K, V]) Keys(_ context.Context) ([]K, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	keys := make([]K, 0, len(b.entries))
	for k := range b.entries {
		keys = append(keys, k)
	}
	return keys, nil
}

// GetEmbedding retrieves the embedding for a key without incrementing frequency.
func (b *LFUBackend[K, V]) GetEmbedding(_ context.Context, key K) ([]float64, bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if e, ok := b.entries[key]; ok {
		return e.entry.Embedding, true, nil
	}
	return nil, false, nil
}
