package semanticcache

import (
	"errors"
	"sort"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// Embedding represents a semantic vector.
type Embedding = []float32

// Comparator defines a function to compute similarity between two embeddings.
type Comparator func(a, b Embedding) float32

// Entry holds an embedding and its associated response.
type Entry struct {
	Embedding Embedding
	Response  any
}

// Match represents a cache hit with its response and similarity score.
type Match struct {
	Response any
	Score    float32
}

// SemanticCache is an in-memory semantic cache with LRU eviction.
type SemanticCache struct {
	mu         sync.RWMutex
	cache      *lru.Cache[string, Entry]
	comparator Comparator
	capacity   int
}

// New creates a SemanticCache with the given capacity and comparator function.
func New(capacity int, comparator Comparator) (*SemanticCache, error) {
	if comparator == nil {
		return nil, errors.New("comparator function cannot be nil")
	}
	lruCache, err := lru.New[string, Entry](capacity)
	if err != nil {
		return nil, err
	}
	return &SemanticCache{
		cache:      lruCache,
		comparator: comparator,
		capacity:   capacity,
	}, nil
}

// Set stores or updates the entry for key.
func (sc *SemanticCache) Set(key string, embedding Embedding, response any) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.cache.Add(key, Entry{Embedding: embedding, Response: response})
	return nil
}

// Get retrieves the response for key, if present.
func (sc *SemanticCache) Get(key string) (any, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if entry, ok := sc.cache.Get(key); ok {
		return entry.Response, true
	}
	return nil, false
}

// Contains checks for key presence without affecting recency.
func (sc *SemanticCache) Contains(key string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.cache.Contains(key)
}

// Delete removes the entry for key.
func (sc *SemanticCache) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.cache.Remove(key)
}

// Flush clears all entries from the cache.
func (sc *SemanticCache) Flush() error {
	newCache, err := lru.New[string, Entry](sc.capacity)
	if err != nil {
		return err
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.cache = newCache
	return nil
}

// Len returns the number of items in the cache.
func (sc *SemanticCache) Len() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.cache.Len()
}

// Lookup returns the first response whose embedding similarity >= threshold.
func (sc *SemanticCache) Lookup(embedding Embedding, threshold float32) (any, bool) {
	sc.mu.RLock()
	keys := sc.cache.Keys()
	sc.mu.RUnlock()

	for _, key := range keys {
		sc.mu.RLock()
		entry, ok := sc.cache.Peek(key)
		sc.mu.RUnlock()
		if ok && sc.comparator(embedding, entry.Embedding) >= threshold {
			sc.mu.Lock()
			defer sc.mu.Unlock()
			if entry, ok := sc.cache.Get(key); ok {
				return entry.Response, true
			}
		}
	}
	return nil, false
}

// TopMatches returns up to n matches sorted by descending similarity.
func (sc *SemanticCache) TopMatches(embedding Embedding, n int) []Match {
	sc.mu.RLock()
	keys := sc.cache.Keys()
	sc.mu.RUnlock()

	matches := make([]Match, 0, len(keys))
	for _, key := range keys {
		sc.mu.RLock()
		entry, ok := sc.cache.Peek(key)
		sc.mu.RUnlock()
		if ok {
			score := sc.comparator(embedding, entry.Embedding)
			matches = append(matches, Match{Response: entry.Response, Score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	if len(matches) > n {
		return matches[:n]
	}
	return matches
}
