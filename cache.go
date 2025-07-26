package semanticcache

import (
	"context"
	"errors"
	"sort"

	"github.com/botirk38/semanticcache/types"
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

// SemanticCache is a semantic cache with pluggable backends and embedding-based lookup.
type SemanticCache[K comparable, V any] struct {
	backend    types.CacheBackend[K, V]
	provider   types.EmbeddingProvider
	comparator func(a, b []float32) float32
}

// NewSemanticCache creates a SemanticCache with the given backend and embedding provider.
// comparator is optional; if nil, defaults to cosine similarity.
func NewSemanticCache[K comparable, V any](
	backend types.CacheBackend[K, V],
	provider types.EmbeddingProvider,
	comparator func(a, b []float32) float32,
) (*SemanticCache[K, V], error) {
	if backend == nil {
		return nil, errors.New("backend cannot be nil")
	}
	if provider == nil {
		return nil, errors.New("embedding provider cannot be nil")
	}
	if comparator == nil {
		comparator = CosineSimilarity
	}

	sc := &SemanticCache[K, V]{
		backend:    backend,
		provider:   provider,
		comparator: comparator,
	}

	return sc, nil
}

// Set stores or updates the entry for key with embedding computed from inputText.
func (sc *SemanticCache[K, V]) Set(key K, inputText string, value V) error {
	if key == *new(K) { // Zero value check for K
		return errors.New("key cannot be zero value")
	}
	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		return err
	}
	entry := types.Entry[V]{Embedding: embedding, Value: value}
	return sc.backend.Set(context.Background(), key, entry)
}

// Get retrieves the value for key, if present.
func (sc *SemanticCache[K, V]) Get(key K) (V, bool) {
	entry, found, err := sc.backend.Get(context.Background(), key)
	if err != nil || !found {
		var zero V
		return zero, false
	}
	return entry.Value, true
}

// Contains checks for key presence without affecting recency.
func (sc *SemanticCache[K, V]) Contains(key K) bool {
	exists, err := sc.backend.Contains(context.Background(), key)
	return err == nil && exists
}

// Delete removes the entry for key.
func (sc *SemanticCache[K, V]) Delete(key K) {
	_ = sc.backend.Delete(context.Background(), key)
}

// Flush clears all entries from the cache and the index.
func (sc *SemanticCache[K, V]) Flush() error {
	return sc.backend.Flush(context.Background())
}

// Len returns the number of items in the cache.
func (sc *SemanticCache[K, V]) Len() int {
	count, err := sc.backend.Len(context.Background())
	if err != nil {
		return 0
	}
	return count
}

// Lookup returns the first value whose embedding similarity >= threshold.
func (sc *SemanticCache[K, V]) Lookup(inputText string, threshold float32) (V, bool, error) {
	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		var zero V
		return zero, false, err
	}

	keys, err := sc.backend.Keys(context.Background())
	if err != nil {
		var zero V
		return zero, false, err
	}

	for _, key := range keys {
		emb, found, err := sc.backend.GetEmbedding(context.Background(), key)
		if err != nil || !found {
			continue
		}

		score := sc.comparator(embedding, emb)
		if score >= threshold {
			entry, found, err := sc.backend.Get(context.Background(), key)
			if err == nil && found {
				return entry.Value, true, nil
			}
		}
	}
	var zero V
	return zero, false, nil
}

// TopMatches returns up to n matches sorted by descending similarity.
func (sc *SemanticCache[K, V]) TopMatches(inputText string, n int) ([]Match[V], error) {
	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		return nil, err
	}

	keys, err := sc.backend.Keys(context.Background())
	if err != nil {
		return nil, err
	}

	matches := make([]Match[V], 0, len(keys))
	for _, key := range keys {
		emb, found, err := sc.backend.GetEmbedding(context.Background(), key)
		if err != nil || !found {
			continue
		}

		score := sc.comparator(embedding, emb)
		entry, found, err := sc.backend.Get(context.Background(), key)
		if err == nil && found {
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

// Close frees resources used by the provider and backend.
func (sc *SemanticCache[K, V]) Close() {
	if sc.provider != nil {
		sc.provider.Close()
	}
	if sc.backend != nil {
		_ = sc.backend.Close()
	}
}

// CosineSimilarity computes the cosine similarity between two float32 vectors.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float32
	for i := range a {
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
	for range 8 {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
