package backends

import (
	"errors"

	"github.com/botirk38/semanticcache/backends/inmemory"
	"github.com/botirk38/semanticcache/backends/remote"
	"github.com/botirk38/semanticcache/types"
)

var ErrUnsupportedBackend = errors.New("unsupported backend type")

// BackendFactory creates cache backends based on type and configuration
type BackendFactory[K comparable, V any] struct{}

// NewBackend creates a new cache backend of the specified type
func (f *BackendFactory[K, V]) NewBackend(backendType types.BackendType, config types.BackendConfig) (types.CacheBackend[K, V], error) {
	switch backendType {
	case types.BackendLRU:
		return NewLRUBackend[K, V](config)
	case types.BackendFIFO:
		return NewFIFOBackend[K, V](config)
	case types.BackendLFU:
		return NewLFUBackend[K, V](config)
	case types.BackendRedis:
		return NewRedisBackend[K, V](config)
	default:
		return nil, ErrUnsupportedBackend
	}
}

// NewLRUBackend creates a new LRU backend
func NewLRUBackend[K comparable, V any](config types.BackendConfig) (types.CacheBackend[K, V], error) {
	return inmemory.NewLRUBackend[K, V](config)
}

// NewFIFOBackend creates a new FIFO backend
func NewFIFOBackend[K comparable, V any](config types.BackendConfig) (types.CacheBackend[K, V], error) {
	return inmemory.NewFIFOBackend[K, V](config)
}

// NewLFUBackend creates a new LFU backend
func NewLFUBackend[K comparable, V any](config types.BackendConfig) (types.CacheBackend[K, V], error) {
	return inmemory.NewLFUBackend[K, V](config)
}

// NewRedisBackend creates a new Redis backend
func NewRedisBackend[K comparable, V any](config types.BackendConfig) (types.CacheBackend[K, V], error) {
	return remote.NewRedisBackend[K, V](config)
}
