// Package backends re-exports the concrete backend constructors for convenience.
package backends

import (
	"github.com/botirk38/semanticcache/backends/inmemory"
	"github.com/botirk38/semanticcache/backends/remote"
	"github.com/botirk38/semanticcache/types"
)

// NewLRUBackend creates a new LRU in-memory backend.
func NewLRUBackend[K comparable, V any](capacity int) (types.Backend[K, V], error) {
	return inmemory.NewLRUBackend[K, V](capacity)
}

// NewFIFOBackend creates a new FIFO in-memory backend.
func NewFIFOBackend[K comparable, V any](capacity int) (types.Backend[K, V], error) {
	return inmemory.NewFIFOBackend[K, V](capacity)
}

// NewLFUBackend creates a new LFU in-memory backend.
func NewLFUBackend[K comparable, V any](capacity int) (types.Backend[K, V], error) {
	return inmemory.NewLFUBackend[K, V](capacity)
}

// NewRedisBackend creates a new Redis backend.
func NewRedisBackend[K comparable, V any](addr string, opts ...remote.RedisOption) (types.Backend[K, V], error) {
	return remote.NewRedisBackend[K, V](addr, opts...)
}
