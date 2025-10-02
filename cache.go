package semanticcache

import (
	"context"
	"errors"
	"sort"

	"github.com/botirk38/semanticcache/options"
	"github.com/botirk38/semanticcache/similarity"
	"github.com/botirk38/semanticcache/types"
)

// SemanticCache represents the semantic cache with configurable backend and embedding provider.
type SemanticCache[K comparable, V any] struct {
	backend    types.CacheBackend[K, V]
	provider   types.EmbeddingProvider
	comparator similarity.SimilarityFunc
}

// Match represents a semantic search result with its similarity score.
type Match[V any] struct {
	Value V       `json:"value"`
	Score float32 `json:"score"`
}

// BatchItem represents an item to be set in batch operations.
type BatchItem[K comparable, V any] struct {
	Key       K
	InputText string
	Value     V
}

// New creates a SemanticCache with functional options.
// This provides a more ergonomic API compared to NewSemanticCache.
func New[K comparable, V any](opts ...options.Option[K, V]) (*SemanticCache[K, V], error) {
	cfg := options.NewConfig[K, V]()

	if err := cfg.Apply(opts...); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return NewSemanticCache(cfg.Backend, cfg.Provider, cfg.Comparator)
}

// NewSemanticCache creates a new semantic cache with the given backend, provider, and comparator.
func NewSemanticCache[K comparable, V any](backend types.CacheBackend[K, V], provider types.EmbeddingProvider, comparator similarity.SimilarityFunc) (*SemanticCache[K, V], error) {
	if backend == nil {
		return nil, errors.New("backend cannot be nil")
	}
	if provider == nil {
		return nil, errors.New("provider cannot be nil")
	}
	if comparator == nil {
		return nil, errors.New("comparator cannot be nil")
	}

	return &SemanticCache[K, V]{
		backend:    backend,
		provider:   provider,
		comparator: comparator,
	}, nil
}

// Set stores or updates the entry for key with embedding computed from inputText.
func (sc *SemanticCache[K, V]) Set(ctx context.Context, key K, inputText string, value V) error {
	if key == *new(K) { // Zero value check for K
		return errors.New("key cannot be zero value")
	}
	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		return err
	}
	entry := types.Entry[V]{Embedding: embedding, Value: value}
	return sc.backend.Set(ctx, key, entry)
}

// Get retrieves the value for key, if present.
func (sc *SemanticCache[K, V]) Get(ctx context.Context, key K) (V, bool, error) {
	entry, found, err := sc.backend.Get(ctx, key)
	if err != nil {
		var zero V
		return zero, false, err
	}
	if !found {
		var zero V
		return zero, false, nil
	}
	return entry.Value, true, nil
}

// Contains checks for key presence without affecting recency.
func (sc *SemanticCache[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	return sc.backend.Contains(ctx, key)
}

// Delete removes the entry for key.
func (sc *SemanticCache[K, V]) Delete(ctx context.Context, key K) error {
	return sc.backend.Delete(ctx, key)
}

// Flush clears all entries from the cache and the index.
func (sc *SemanticCache[K, V]) Flush(ctx context.Context) error {
	return sc.backend.Flush(ctx)
}

// Len returns the number of items in the cache.
func (sc *SemanticCache[K, V]) Len(ctx context.Context) (int, error) {
	return sc.backend.Len(ctx)
}

// Lookup returns the first value whose embedding similarity >= threshold.
func (sc *SemanticCache[K, V]) Lookup(ctx context.Context, inputText string, threshold float32) (*Match[V], error) {
	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		return nil, err
	}

	keys, err := sc.backend.Keys(ctx)
	if err != nil {
		return nil, err
	}

	bestMatch := (*Match[V])(nil)
	bestScore := threshold

	for _, key := range keys {
		emb, found, err := sc.backend.GetEmbedding(ctx, key)
		if err != nil || !found {
			continue
		}

		score := sc.comparator(embedding, emb)
		if score >= bestScore {
			entry, found, err := sc.backend.Get(ctx, key)
			if err == nil && found {
				bestMatch = &Match[V]{Value: entry.Value, Score: score}
				bestScore = score // Update threshold to find even better matches
			}
		}
	}

	return bestMatch, nil
}

// TopMatches returns up to n matches sorted by descending similarity.
func (sc *SemanticCache[K, V]) TopMatches(ctx context.Context, inputText string, n int) ([]Match[V], error) {
	if n <= 0 {
		return nil, errors.New("n must be positive")
	}

	embedding, err := sc.provider.EmbedText(inputText)
	if err != nil {
		return nil, err
	}

	keys, err := sc.backend.Keys(ctx)
	if err != nil {
		return nil, err
	}

	matches := make([]Match[V], 0, len(keys))
	for _, key := range keys {
		emb, found, err := sc.backend.GetEmbedding(ctx, key)
		if err != nil || !found {
			continue
		}

		score := sc.comparator(embedding, emb)
		entry, found, err := sc.backend.Get(ctx, key)
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

// SetBatch stores multiple entries efficiently in a single operation.
func (sc *SemanticCache[K, V]) SetBatch(ctx context.Context, items []BatchItem[K, V]) error {
	if len(items) == 0 {
		return nil
	}

	for _, item := range items {
		if item.Key == *new(K) {
			return errors.New("key cannot be zero value")
		}
		if err := sc.Set(ctx, item.Key, item.InputText, item.Value); err != nil {
			return err
		}
	}
	return nil
}

// GetBatch retrieves multiple values efficiently in a single operation.
func (sc *SemanticCache[K, V]) GetBatch(ctx context.Context, keys []K) (map[K]V, error) {
	result := make(map[K]V)
	for _, key := range keys {
		if value, found, err := sc.Get(ctx, key); err != nil {
			return nil, err
		} else if found {
			result[key] = value
		}
	}
	return result, nil
}

// DeleteBatch removes multiple entries efficiently in a single operation.
func (sc *SemanticCache[K, V]) DeleteBatch(ctx context.Context, keys []K) error {
	for _, key := range keys {
		if err := sc.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the underlying backend and provider.
func (sc *SemanticCache[K, V]) Close() error {
	sc.provider.Close()
	return nil
}

// SetAsync stores or updates the entry asynchronously using backend async capabilities.
// Returns a channel that will receive an error or nil when complete.
func (sc *SemanticCache[K, V]) SetAsync(ctx context.Context, key K, inputText string, value V) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if key == *new(K) {
			errCh <- errors.New("key cannot be zero value")
			return
		}
		embedding, err := sc.provider.EmbedText(inputText)
		if err != nil {
			errCh <- err
			return
		}
		entry := types.Entry[V]{Embedding: embedding, Value: value}

		// Use backend's async method
		backendErrCh := sc.backend.SetAsync(ctx, key, entry)
		errCh <- <-backendErrCh
	}()
	return errCh
}

// GetResult holds the result of an async Get operation.
type GetResult[V any] struct {
	Value V
	Found bool
	Error error
}

// GetAsync retrieves the value asynchronously using backend async capabilities.
// Returns a channel that will receive the result when complete.
func (sc *SemanticCache[K, V]) GetAsync(ctx context.Context, key K) <-chan GetResult[V] {
	resultCh := make(chan GetResult[V], 1)
	go func() {
		defer close(resultCh)

		// Use backend's async method
		backendResultCh := sc.backend.GetAsync(ctx, key)
		backendResult := <-backendResultCh

		if backendResult.Error != nil {
			resultCh <- GetResult[V]{Error: backendResult.Error}
			return
		}

		resultCh <- GetResult[V]{
			Value: backendResult.Entry.Value,
			Found: backendResult.Found,
			Error: nil,
		}
	}()
	return resultCh
}

// DeleteAsync removes the entry asynchronously using backend async capabilities.
// Returns a channel that will receive an error or nil when complete.
func (sc *SemanticCache[K, V]) DeleteAsync(ctx context.Context, key K) <-chan error {
	return sc.backend.DeleteAsync(ctx, key)
}

// LookupResult holds the result of an async Lookup operation.
type LookupResult[V any] struct {
	Match *Match[V]
	Error error
}

// LookupAsync performs semantic lookup asynchronously.
// Returns a channel that will receive the result when complete.
func (sc *SemanticCache[K, V]) LookupAsync(ctx context.Context, inputText string, threshold float32) <-chan LookupResult[V] {
	resultCh := make(chan LookupResult[V], 1)
	go func() {
		defer close(resultCh)
		match, err := sc.Lookup(ctx, inputText, threshold)
		resultCh <- LookupResult[V]{Match: match, Error: err}
	}()
	return resultCh
}

// TopMatchesResult holds the result of an async TopMatches operation.
type TopMatchesResult[V any] struct {
	Matches []Match[V]
	Error   error
}

// TopMatchesAsync returns top matches asynchronously.
// Returns a channel that will receive the result when complete.
func (sc *SemanticCache[K, V]) TopMatchesAsync(ctx context.Context, inputText string, n int) <-chan TopMatchesResult[V] {
	resultCh := make(chan TopMatchesResult[V], 1)
	go func() {
		defer close(resultCh)
		matches, err := sc.TopMatches(ctx, inputText, n)
		resultCh <- TopMatchesResult[V]{Matches: matches, Error: err}
	}()
	return resultCh
}

// SetBatchAsync stores multiple entries asynchronously.
// Returns a channel that will receive an error or nil when complete.
func (sc *SemanticCache[K, V]) SetBatchAsync(ctx context.Context, items []BatchItem[K, V]) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if len(items) == 0 {
			errCh <- nil
			return
		}

		for _, item := range items {
			if item.Key == *new(K) {
				errCh <- errors.New("key cannot be zero value")
				return
			}
		}

		// Process all items and use backend async for each
		type setResult struct {
			err error
		}
		resultCh := make(chan setResult, len(items))

		for _, item := range items {
			go func(it BatchItem[K, V]) {
				embedding, err := sc.provider.EmbedText(it.InputText)
				if err != nil {
					resultCh <- setResult{err: err}
					return
				}
				entry := types.Entry[V]{Embedding: embedding, Value: it.Value}
				backendErrCh := sc.backend.SetAsync(ctx, it.Key, entry)
				resultCh <- setResult{err: <-backendErrCh}
			}(item)
		}

		// Wait for all to complete
		for range items {
			result := <-resultCh
			if result.err != nil {
				errCh <- result.err
				return
			}
		}
		errCh <- nil
	}()
	return errCh
}

// GetBatchResult holds the result of an async GetBatch operation.
type GetBatchResult[K comparable, V any] struct {
	Values map[K]V
	Error  error
}

// GetBatchAsync retrieves multiple values asynchronously using backend async capabilities.
// Returns a channel that will receive the result when complete.
func (sc *SemanticCache[K, V]) GetBatchAsync(ctx context.Context, keys []K) <-chan GetBatchResult[K, V] {
	resultCh := make(chan GetBatchResult[K, V], 1)
	go func() {
		defer close(resultCh)

		// Use backend's async batch method
		backendResultCh := sc.backend.GetBatchAsync(ctx, keys)
		backendResult := <-backendResultCh

		if backendResult.Error != nil {
			resultCh <- GetBatchResult[K, V]{Error: backendResult.Error}
			return
		}

		// Extract values from entries
		values := make(map[K]V)
		for key, entry := range backendResult.Entries {
			values[key] = entry.Value
		}

		resultCh <- GetBatchResult[K, V]{Values: values, Error: nil}
	}()
	return resultCh
}

// DeleteBatchAsync removes multiple entries asynchronously using backend async capabilities.
// Returns a channel that will receive an error or nil when complete.
func (sc *SemanticCache[K, V]) DeleteBatchAsync(ctx context.Context, keys []K) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)

		// Use goroutines to delete keys concurrently
		type delResult struct {
			err error
		}
		resultCh := make(chan delResult, len(keys))

		for _, key := range keys {
			go func(k K) {
				backendErrCh := sc.backend.DeleteAsync(ctx, k)
				resultCh <- delResult{err: <-backendErrCh}
			}(key)
		}

		// Wait for all to complete
		for range keys {
			result := <-resultCh
			if result.err != nil {
				errCh <- result.err
				return
			}
		}
		errCh <- nil
	}()
	return errCh
}
