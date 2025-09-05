// Package options provides functional options for configuring SemanticCache instances.
package options

import (
	"errors"

	"github.com/botirk38/semanticcache/backends"
	"github.com/botirk38/semanticcache/providers/openai"
	"github.com/botirk38/semanticcache/similarity"
	"github.com/botirk38/semanticcache/types"
)

// Option represents a configuration option for SemanticCache
type Option[K comparable, V any] func(*Config[K, V]) error

// Config holds the configuration for building a SemanticCache
type Config[K comparable, V any] struct {
	Backend    types.CacheBackend[K, V]
	Provider   types.EmbeddingProvider
	Comparator similarity.SimilarityFunc
}

// NewConfig creates a new configuration with default values
func NewConfig[K comparable, V any]() *Config[K, V] {
	return &Config[K, V]{
		Comparator: similarity.CosineSimilarity,
	}
}

// Apply applies all the given options to the config
func (c *Config[K, V]) Apply(opts ...Option[K, V]) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks if the configuration is valid
func (c *Config[K, V]) Validate() error {
	if c.Backend == nil {
		return errors.New("backend is required - use WithLRUBackend, WithRedisBackend, etc.")
	}
	if c.Provider == nil {
		return errors.New("embedding provider is required - use WithOpenAIProvider, etc.")
	}
	return nil
}

// WithLRUBackend sets up an LRU in-memory backend
func WithLRUBackend[K comparable, V any](capacity int) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		backend, err := backends.NewLRUBackend[K, V](types.BackendConfig{
			Capacity: capacity,
		})
		if err != nil {
			return err
		}
		cfg.Backend = backend
		return nil
	}
}

// WithFIFOBackend sets up a FIFO in-memory backend
func WithFIFOBackend[K comparable, V any](capacity int) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		backend, err := backends.NewFIFOBackend[K, V](types.BackendConfig{
			Capacity: capacity,
		})
		if err != nil {
			return err
		}
		cfg.Backend = backend
		return nil
	}
}

// WithLFUBackend sets up an LFU in-memory backend
func WithLFUBackend[K comparable, V any](capacity int) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		backend, err := backends.NewLFUBackend[K, V](types.BackendConfig{
			Capacity: capacity,
		})
		if err != nil {
			return err
		}
		cfg.Backend = backend
		return nil
	}
}

// WithRedisBackend sets up a Redis backend
func WithRedisBackend[K comparable, V any](addr string, db int) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		backend, err := backends.NewRedisBackend[K, V](types.BackendConfig{
			ConnectionString: addr,
			Database:         db,
		})
		if err != nil {
			return err
		}
		cfg.Backend = backend
		return nil
	}
}

// WithCustomBackend allows using a pre-configured backend
func WithCustomBackend[K comparable, V any](backend types.CacheBackend[K, V]) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		if backend == nil {
			return errors.New("backend cannot be nil")
		}
		cfg.Backend = backend
		return nil
	}
}

// WithOpenAIProvider sets up OpenAI embedding provider
func WithOpenAIProvider[K comparable, V any](apiKey string, model ...string) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		config := openai.OpenAIConfig{
			APIKey: apiKey,
		}
		if len(model) > 0 {
			config.Model = model[0]
		}

		provider, err := openai.NewOpenAIProvider(config)
		if err != nil {
			return err
		}
		cfg.Provider = provider
		return nil
	}
}

// WithCustomProvider allows using a pre-configured embedding provider
func WithCustomProvider[K comparable, V any](provider types.EmbeddingProvider) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		if provider == nil {
			return errors.New("provider cannot be nil")
		}
		cfg.Provider = provider
		return nil
	}
}

// WithSimilarityComparator sets a custom similarity function
func WithSimilarityComparator[K comparable, V any](comparator similarity.SimilarityFunc) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		if comparator == nil {
			return errors.New("comparator cannot be nil")
		}
		cfg.Comparator = comparator
		return nil
	}
}
