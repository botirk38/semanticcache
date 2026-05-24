package semanticcache

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"

	scerrors "github.com/botirk38/semanticcache/errors"
	"github.com/botirk38/semanticcache/options"
	"github.com/botirk38/semanticcache/similarity"
	"github.com/botirk38/semanticcache/types"
)

// Cache is a synchronous semantic cache.
type Cache[K comparable, V any] struct {
	backend    types.Backend[K, V]
	provider   types.EmbeddingProvider
	comparator similarity.SimilarityFunc
	closed     atomic.Bool
}

// Match represents a semantic search result with its similarity score.
type Match[V any] struct {
	Value V       `json:"value"`
	Score float64 `json:"score"`
}

// BatchItem represents an item to be stored via SetBatch.
type BatchItem[K comparable, V any] struct {
	Key       K
	InputText string
	Value     V
}

// New creates a Cache with functional options.
func New[K comparable, V any](opts ...options.Option[K, V]) (*Cache[K, V], error) {
	cfg := options.NewConfig[K, V]()
	if err := cfg.Apply(opts...); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Cache[K, V]{
		backend:    cfg.Backend,
		provider:   cfg.Provider,
		comparator: cfg.Comparator,
	}, nil
}

// NewSemanticCache creates a Cache from explicit components.
func NewSemanticCache[K comparable, V any](
	backend types.Backend[K, V],
	provider types.EmbeddingProvider,
	comparator similarity.SimilarityFunc,
) (*Cache[K, V], error) {
	if backend == nil {
		return nil, scerrors.ErrNilBackend
	}
	if provider == nil {
		return nil, scerrors.ErrNilProvider
	}
	if comparator == nil {
		return nil, scerrors.ErrNilComparator
	}
	return &Cache[K, V]{
		backend:    backend,
		provider:   provider,
		comparator: comparator,
	}, nil
}

func (c *Cache[K, V]) checkClosed() error {
	if c.closed.Load() {
		return scerrors.ErrClosed
	}
	return nil
}

// Set stores a value keyed by key, computing the embedding from inputText.
func (c *Cache[K, V]) Set(ctx context.Context, key K, inputText string, value V) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	if key == *new(K) {
		return scerrors.ErrZeroKey
	}
	embedding, err := c.provider.EmbedText(ctx, inputText)
	if err != nil {
		return err
	}
	return c.backend.Set(ctx, key, embedding, value)
}

// Get retrieves the value for key.
func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, bool, error) {
	if err := c.checkClosed(); err != nil {
		var zero V
		return zero, false, err
	}
	return c.backend.Get(ctx, key)
}

// Contains checks for key presence.
func (c *Cache[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	if err := c.checkClosed(); err != nil {
		return false, err
	}
	return c.backend.Contains(ctx, key)
}

// Delete removes the entry for key.
func (c *Cache[K, V]) Delete(ctx context.Context, key K) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.backend.Delete(ctx, key)
}

// Flush clears all entries.
func (c *Cache[K, V]) Flush(ctx context.Context) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.backend.Flush(ctx)
}

// Len returns the number of cached entries.
func (c *Cache[K, V]) Len(ctx context.Context) (int, error) {
	if err := c.checkClosed(); err != nil {
		return 0, err
	}
	return c.backend.Len(ctx)
}

// Lookup returns the best match whose similarity >= threshold, or nil.
// If the backend implements VectorSearcher, server-side search is used;
// otherwise all keys are scanned client-side.
func (c *Cache[K, V]) Lookup(ctx context.Context, inputText string, threshold float64) (*Match[V], error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	embedding, err := c.provider.EmbedText(ctx, inputText)
	if err != nil {
		return nil, err
	}

	if vs, ok := c.backend.(types.VectorSearcher[K, V]); ok {
		results, err := vs.VectorSearch(ctx, embedding, threshold, 1)
		if err != nil {
			return nil, err
		}
		if len(results) == 0 {
			return nil, nil
		}
		return &Match[V]{Value: results[0].Value, Score: results[0].Score}, nil
	}

	return c.lookupScan(ctx, embedding, threshold)
}

func (c *Cache[K, V]) lookupScan(ctx context.Context, query []float64, threshold float64) (*Match[V], error) {
	keys, err := c.backend.Keys(ctx)
	if err != nil {
		return nil, err
	}

	var best *Match[V]
	bestScore := threshold

	for _, key := range keys {
		emb, found, err := c.backend.GetEmbedding(ctx, key)
		if err != nil || !found {
			continue
		}
		score := c.comparator(query, emb)
		if score >= bestScore {
			value, found, err := c.backend.Get(ctx, key)
			if err == nil && found {
				best = &Match[V]{Value: value, Score: score}
				bestScore = score
			}
		}
	}
	return best, nil
}

// TopMatches returns up to n matches sorted by descending similarity.
// If the backend implements VectorSearcher, server-side search is used;
// otherwise all keys are scanned client-side.
func (c *Cache[K, V]) TopMatches(ctx context.Context, inputText string, n int) ([]Match[V], error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, scerrors.ErrInvalidN
	}

	embedding, err := c.provider.EmbedText(ctx, inputText)
	if err != nil {
		return nil, err
	}

	if vs, ok := c.backend.(types.VectorSearcher[K, V]); ok {
		results, err := vs.VectorSearch(ctx, embedding, 0, n)
		if err != nil {
			return nil, err
		}
		matches := make([]Match[V], len(results))
		for i, r := range results {
			matches[i] = Match[V]{Value: r.Value, Score: r.Score}
		}
		return matches, nil
	}

	return c.topMatchesScan(ctx, embedding, n)
}

func (c *Cache[K, V]) topMatchesScan(ctx context.Context, query []float64, n int) ([]Match[V], error) {
	keys, err := c.backend.Keys(ctx)
	if err != nil {
		return nil, err
	}

	matches := make([]Match[V], 0, len(keys))
	for _, key := range keys {
		emb, found, err := c.backend.GetEmbedding(ctx, key)
		if err != nil || !found {
			continue
		}
		score := c.comparator(query, emb)
		value, found, err := c.backend.Get(ctx, key)
		if err == nil && found {
			matches = append(matches, Match[V]{Value: value, Score: score})
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

// SetBatch stores multiple entries.
func (c *Cache[K, V]) SetBatch(ctx context.Context, items []BatchItem[K, V]) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	for _, item := range items {
		if item.Key == *new(K) {
			return scerrors.ErrZeroKey
		}
		if err := c.Set(ctx, item.Key, item.InputText, item.Value); err != nil {
			return err
		}
	}
	return nil
}

// GetBatch retrieves multiple values by key.
func (c *Cache[K, V]) GetBatch(ctx context.Context, keys []K) (map[K]V, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	result := make(map[K]V)
	for _, key := range keys {
		if value, found, err := c.backend.Get(ctx, key); err != nil {
			return nil, err
		} else if found {
			result[key] = value
		}
	}
	return result, nil
}

// DeleteBatch removes multiple entries.
func (c *Cache[K, V]) DeleteBatch(ctx context.Context, keys []K) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	for _, key := range keys {
		if err := c.backend.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Close closes both the backend and the provider.
func (c *Cache[K, V]) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	provErr := c.provider.Close()
	backendErr := c.backend.Close()
	if provErr != nil {
		return fmt.Errorf("provider close: %w", provErr)
	}
	return backendErr
}
