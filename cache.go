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

// Cache is a synchronous semantic cache that stores values alongside
// their embedding vectors and supports similarity-based retrieval.
type Cache[K comparable, V any] struct {
	backend    types.Backend[K, V]
	provider   types.EmbeddingProvider
	comparator similarity.SimilarityFunc
	closed     atomic.Bool
}

// Match is a single semantic search result.
type Match[V any] struct {
	Value V       `json:"value"`
	Score float64 `json:"score"`
}

// BatchItem is an input for SetBatch.
type BatchItem[K comparable, V any] struct {
	Key       K
	InputText string
	Value     V
}

// New creates a Cache using functional options.
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

// Set stores a value, computing the embedding from inputText.
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

// Contains reports whether key exists.
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

// Flush removes all entries.
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

// Lookup finds the single best match whose similarity >= threshold.
// Returns nil when nothing meets the threshold.
func (c *Cache[K, V]) Lookup(ctx context.Context, inputText string, threshold float64) (*Match[V], error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	query, err := c.provider.EmbedText(ctx, inputText)
	if err != nil {
		return nil, err
	}

	keys, err := c.backend.Keys(ctx)
	if err != nil {
		return nil, err
	}

	var best *Match[V]
	bestScore := threshold

	for _, key := range keys {
		emb, ok, err := c.backend.GetEmbedding(ctx, key)
		if err != nil || !ok {
			continue
		}
		score := c.comparator(query, emb)
		if score >= bestScore {
			val, found, err := c.backend.Get(ctx, key)
			if err == nil && found {
				best = &Match[V]{Value: val, Score: score}
				bestScore = score
			}
		}
	}
	return best, nil
}

// TopMatches returns up to n entries sorted by descending similarity.
func (c *Cache[K, V]) TopMatches(ctx context.Context, inputText string, n int) ([]Match[V], error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, scerrors.ErrInvalidN
	}
	query, err := c.provider.EmbedText(ctx, inputText)
	if err != nil {
		return nil, err
	}

	keys, err := c.backend.Keys(ctx)
	if err != nil {
		return nil, err
	}

	matches := make([]Match[V], 0, len(keys))
	for _, key := range keys {
		emb, ok, err := c.backend.GetEmbedding(ctx, key)
		if err != nil || !ok {
			continue
		}
		score := c.comparator(query, emb)
		val, found, err := c.backend.Get(ctx, key)
		if err == nil && found {
			matches = append(matches, Match[V]{Value: val, Score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	if len(matches) > n {
		matches = matches[:n]
	}
	return matches, nil
}

// SetBatch stores multiple items.
func (c *Cache[K, V]) SetBatch(ctx context.Context, items []BatchItem[K, V]) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	for _, item := range items {
		if err := c.Set(ctx, item.Key, item.InputText, item.Value); err != nil {
			return err
		}
	}
	return nil
}

// GetBatch retrieves multiple values. Missing keys are omitted.
func (c *Cache[K, V]) GetBatch(ctx context.Context, keys []K) (map[K]V, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	result := make(map[K]V, len(keys))
	for _, key := range keys {
		val, found, err := c.backend.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		if found {
			result[key] = val
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

// Close releases both the provider and backend.
func (c *Cache[K, V]) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	pErr := c.provider.Close()
	bErr := c.backend.Close()
	if pErr != nil {
		return fmt.Errorf("provider close: %w", pErr)
	}
	return bErr
}
