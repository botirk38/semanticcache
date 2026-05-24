# SemanticCache -- Development Instructions

## Overview
Semantic caching library for Go. Stores values with embedding vectors, retrieves by similarity. Sync-only `Cache[K, V]` type with pluggable backends and embedding providers.

## Project structure
```
semanticcache/
  cache.go, errors.go          Cache type, constructors, sentinel errors
  types/                       Backend[K,V] and EmbeddingProvider interfaces
  options/                     Functional options (With* functions), config errors
  backends/
    inmemory/                  LRU, LFU, FIFO (thread-safe)
    remote/                    Redis (JSON storage)
  providers/
    openai/                    OpenAI embeddings (official SDK)
    local/                     Hash-based provider for testing (no API key)
  similarity/                  Cosine, Euclidean, DotProduct, Manhattan, Pearson
  chunker/                     Text chunking utilities
  tokenizer/                   Token counting (OpenAI, Anthropic, Gemini)
```

## Key design decisions
- Sync-only. No async cache.
- `Backend[K, V]` has 9 methods. Every backend implements all of them.
- No centralized errors package. Each package defines its own errors.
- Functional options pattern for configuration.
- `context.Context` on all operations.
- Generics throughout: `Cache[K comparable, V any]`.

## Commands
```
go test ./...             # all tests
go test -race ./...       # with race detector
go test -bench=. ./...    # benchmarks
go vet ./...              # static analysis
gofmt -l .                # formatting check
go build ./...            # build all
```

## Import path
```go
import "github.com/botirk38/semanticcache"
```

All subpackages: `github.com/botirk38/semanticcache/{options,types,backends/inmemory,...}`

## Adding a new backend
1. Implement `types.Backend[K, V]` (9 methods)
2. Add `With*Backend` option in `options/options.go`
3. Add compile-time check: `var _ types.Backend[string, string] = (*YourBackend[string, string])(nil)`
4. Add tests in the backend's package
5. Add a re-export in `backends/backends.go`

## Adding a new provider
1. Implement `types.EmbeddingProvider` (2 methods: `EmbedText`, `Close`)
2. Optionally implement `types.BatchEmbeddingProvider`
3. Add `With*Provider` option in `options/options.go`
4. Add tests (use `httptest` for HTTP providers)
5. Add a re-export in `providers/providers.go`

## Adding a similarity function
1. Create a new file in `similarity/`
2. Implement `func(a, b []float64) float64`
3. Add tests in `similarity_test.go`

## Error conventions
- Each package defines its own sentinel errors.
- Root package: `ErrClosed`, `ErrZeroKey`, `ErrInvalidN`
- Options: `ErrNilBackend`, `ErrNilProvider`, `ErrNilComparator`
- Chunker: `ErrInvalidChunkSize`, `ErrEmptyText`, etc.
