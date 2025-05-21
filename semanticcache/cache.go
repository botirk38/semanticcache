package semanticcache

import (
	"errors"
	"sort"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

// Entry holds an embedding and its associated value.
type Entry[V any] struct {
	Embedding []float32
	Value     V
}

// Match represents a cache hit with its value and similarity score.
type Match[V any] struct {
	Value V
	Score float32
}

// SemanticCache is an in-memory semantic cache with LRU eviction and embedding-based lookup.
type SemanticCache[V any] struct {
	mu         sync.RWMutex
	cache      *lru.Cache[string, Entry[V]]
	index      map[string][]float32 // key -> embedding, for brute force search
	provider   EmbeddingProvider    // interface for embedding text
	capacity   int
	comparator func(a, b []float32) float32 // Similarity function
}

// NewSemanticCache creates a SemanticCache with the given capacity and embedding provider.
// comparator is optional; if nil, defaults to cosine similarity.
func NewSemanticCache[V any](capacity int, provider EmbeddingProvider, comparator func(a, b []float32) float32) (*SemanticCache[V], error) {
	if provider == nil {
		return nil, errors.New("embedding provider cannot be nil")
	}
	if comparator == nil {
		comparator = CosineSimilarity
	}
	lruCache, err := lru.New[string, Entry[V]](capacity)
	if err != nil {
		return nil, err
	}
	return &SemanticCache[V]{
		cache:      lruCache,
		index:      make(map[string][]float32),
		provider:   provider,
		capacity:   capacity,
		comparator: comparator,
	}, nil
}

// Set stores or updates the entry for key with embedding computed from inputText.
func (sc *SemanticCache[V]) Set(key, inputText string, value V) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		return err
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	entry := Entry[V]{Embedding: embedding, Value: value}
	sc.cache.Add(key, entry)
	sc.index[key] = embedding
	return nil
}

// Get retrieves the value for key, if present.
func (sc *SemanticCache[V]) Get(key string) (V, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if entry, ok := sc.cache.Get(key); ok {
		return entry.Value, true
	}
	var zero V
	return zero, false
}

// Contains checks for key presence without affecting recency.
func (sc *SemanticCache[V]) Contains(key string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.cache.Contains(key)
}

// Delete removes the entry for key.
func (sc *SemanticCache[V]) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.cache.Remove(key)
	delete(sc.index, key)
}

// Flush clears all entries from the cache.
func (sc *SemanticCache[V]) Flush() error {
	newCache, err := lru.New[string, Entry[V]](sc.capacity)
	if err != nil {
		return err
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.cache = newCache
	sc.index = make(map[string][]float32)
	return nil
}

// Len returns the number of items in the cache.
func (sc *SemanticCache[V]) Len() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.cache.Len()
}

// Lookup returns the first value whose embedding similarity >= threshold.
func (sc *SemanticCache[V]) Lookup(inputText string, threshold float32) (V, bool, error) {
	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		var zero V
		return zero, false, err
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()
	for key, emb := range sc.index {
		score := sc.comparator(embedding, emb)
		if score >= threshold {
			if entry, ok := sc.cache.Get(key); ok {
				return entry.Value, true, nil
			}
		}
	}
	var zero V
	return zero, false, nil
}

// TopMatches returns up to n matches sorted by descending similarity.
func (sc *SemanticCache[V]) TopMatches(inputText string, n int) ([]Match[V], error) {
	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		return nil, err
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	matches := make([]Match[V], 0, len(sc.index))
	for key, emb := range sc.index {
		score := sc.comparator(embedding, emb)
		if entry, ok := sc.cache.Get(key); ok {
			matches = append(matches, Match[V]{Value: entry.Value, Score: score})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	if len(matches) > n {
		return matches[:n], nil
	}
	return matches, nil
}

// Close frees resources used by the provider.
func (sc *SemanticCache[V]) Close() {
	if sc.provider != nil {
		sc.provider.Close()
	}
}

// CosineSimilarity computes the cosine similarity between two float32 vectors.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float32
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (sqrt(normA) * sqrt(normB))
}

// sqrt returns the square root of a float32 (simple Newton's method).
func sqrt(x float32) float32 {
	if x == 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 8; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
