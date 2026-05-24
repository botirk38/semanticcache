// Package options provides functional options for configuring SemanticCache.
package options

import (
	"errors"

	"github.com/botirk38/semanticcache/backends/inmemory"
	"github.com/botirk38/semanticcache/backends/remote"
	"github.com/botirk38/semanticcache/providers/local"
	"github.com/botirk38/semanticcache/providers/openai"
	"github.com/botirk38/semanticcache/similarity"
	"github.com/botirk38/semanticcache/types"
)

// Sentinel errors for configuration validation.
var (
	// ErrNilBackend is returned when a nil backend is provided.
	ErrNilBackend = errors.New("options: backend cannot be nil")

	// ErrNilProvider is returned when a nil embedding provider is provided.
	ErrNilProvider = errors.New("options: embedding provider cannot be nil")

	// ErrNilComparator is returned when a nil similarity function is provided.
	ErrNilComparator = errors.New("options: similarity comparator cannot be nil")
)

// Option configures a cache instance.
type Option[K comparable, V any] func(*Config[K, V]) error

// Config holds the resolved configuration for building a cache.
type Config[K comparable, V any] struct {
	Backend    types.Backend[K, V]
	Provider   types.EmbeddingProvider
	Comparator similarity.SimilarityFunc
}

// NewConfig returns a Config with sensible defaults.
func NewConfig[K comparable, V any]() *Config[K, V] {
	return &Config[K, V]{
		Comparator: similarity.CosineSimilarity,
	}
}

// Apply applies all options to the config.
func (c *Config[K, V]) Apply(opts ...Option[K, V]) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks that the config has the required fields.
func (c *Config[K, V]) Validate() error {
	if c.Backend == nil {
		return ErrNilBackend
	}
	if c.Provider == nil {
		return ErrNilProvider
	}
	return nil
}

// ---------- backend options ----------

// WithLRUBackend sets up an LRU in-memory backend.
func WithLRUBackend[K comparable, V any](capacity int) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		b, err := inmemory.NewLRUBackend[K, V](capacity)
		if err != nil {
			return err
		}
		cfg.Backend = b
		return nil
	}
}

// WithFIFOBackend sets up a FIFO in-memory backend.
func WithFIFOBackend[K comparable, V any](capacity int) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		b, err := inmemory.NewFIFOBackend[K, V](capacity)
		if err != nil {
			return err
		}
		cfg.Backend = b
		return nil
	}
}

// WithLFUBackend sets up an LFU in-memory backend.
func WithLFUBackend[K comparable, V any](capacity int) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		b, err := inmemory.NewLFUBackend[K, V](capacity)
		if err != nil {
			return err
		}
		cfg.Backend = b
		return nil
	}
}

// WithRedisBackend sets up a Redis backend. addr can be "host:port" or a
// redis:// URL. Use remote.With* options for password, prefix, etc.
func WithRedisBackend[K comparable, V any](addr string, opts ...remote.RedisOption) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		b, err := remote.NewRedisBackend[K, V](addr, opts...)
		if err != nil {
			return err
		}
		cfg.Backend = b
		return nil
	}
}

// WithCustomBackend uses a pre-constructed backend.
func WithCustomBackend[K comparable, V any](backend types.Backend[K, V]) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		if backend == nil {
			return ErrNilBackend
		}
		cfg.Backend = backend
		return nil
	}
}

// ---------- provider options ----------

// WithOpenAIProvider sets up an OpenAI embedding provider.
func WithOpenAIProvider[K comparable, V any](apiKey string, model ...string) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		c := openai.OpenAIConfig{APIKey: apiKey}
		if len(model) > 0 {
			c.Model = model[0]
		}
		p, err := openai.NewOpenAIProvider(c)
		if err != nil {
			return err
		}
		cfg.Provider = p
		return nil
	}
}

// WithLocalProvider sets up a hash-based provider for testing (no API key needed).
// dimensions controls the vector size (default 128 if <= 0).
func WithLocalProvider[K comparable, V any](dimensions int) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		cfg.Provider = local.New(dimensions)
		return nil
	}
}

// WithCustomProvider uses a pre-constructed embedding provider.
func WithCustomProvider[K comparable, V any](provider types.EmbeddingProvider) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		if provider == nil {
			return ErrNilProvider
		}
		cfg.Provider = provider
		return nil
	}
}

// ---------- similarity options ----------

// WithSimilarityComparator sets a custom similarity function.
func WithSimilarityComparator[K comparable, V any](comparator similarity.SimilarityFunc) Option[K, V] {
	return func(cfg *Config[K, V]) error {
		if comparator == nil {
			return ErrNilComparator
		}
		cfg.Comparator = comparator
		return nil
	}
}
