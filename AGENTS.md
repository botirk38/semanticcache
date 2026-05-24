# SemanticCache -- Agent Instructions

Semantic caching library for Go. Sync-only `Cache[K, V]` with pluggable backends and embedding providers.

## Commands
```
go test ./...             # all tests
go test -race ./...       # race detector
go test -bench=. ./...    # benchmarks
go test -cover ./...      # coverage
go vet ./...              # static analysis
gofmt -l .                # formatting check
go build ./...            # build all
```

## Import path
```go
import "github.com/botirk38/semanticcache"
```

Subpackages: `options`, `types`, `backends/inmemory`, `backends/remote`, `providers/openai`, `providers/local`, `similarity`, `chunker`, `tokenizer`.

## Architecture
- `cache.go` + `errors.go` -- `Cache[K, V]` type, constructors, sentinel errors (`ErrClosed`, `ErrZeroKey`, `ErrInvalidN`)
- `types/` -- `Backend[K, V]` interface (9 methods), `EmbeddingProvider`, `BatchEmbeddingProvider`
- `options/` -- functional options (`With*` functions), config errors (`ErrNilBackend`, `ErrNilProvider`, `ErrNilComparator`)
- `backends/inmemory/` -- LRU, LFU, FIFO (thread-safe via `sync.RWMutex`)
- `backends/remote/` -- Redis (JSON storage, requires RedisJSON or Redis 7.2+)
- `providers/openai/` -- OpenAI SDK, default model `text-embedding-3-small`
- `providers/local/` -- hash-based provider for testing (no API key, not semantically meaningful)
- `similarity/` -- `func(a, b []float64) float64` functions (cosine, euclidean, dot, manhattan, pearson)
- `chunker/` -- text chunking with configurable strategy, its own errors
- `tokenizer/` -- token counting for OpenAI (local), Anthropic (API), Gemini (API)

## Error conventions
Each package defines its own errors. No centralized errors package.

## Adding a backend
1. Implement all 9 methods of `types.Backend[K, V]`
2. Add `With*Backend` option in `options/options.go`
3. Add compile-time check: `var _ types.Backend[string, string] = (*YourBackend[string, string])(nil)`
4. Add tests + benchmarks
5. Add re-export in `backends/backends.go`
6. Add `README.md` and `AGENTS.md` in the new package

## Adding a provider
1. Implement `types.EmbeddingProvider` (`EmbedText`, `Close`)
2. Optionally implement `types.BatchEmbeddingProvider`
3. Add `With*Provider` option in `options/options.go`
4. Add tests (use `httptest` for HTTP-based providers)
5. Add re-export in `providers/providers.go`
6. Add `README.md` and `AGENTS.md` in the new package

## Adding a similarity function
1. New file in `similarity/`, signature `func(a, b []float64) float64`
2. Return 0 for mismatched lengths or empty vectors
3. Add tests in `similarity_test.go`

## Code style
- Standard Go conventions. `gofmt` enforced in CI.
- Functional options pattern for configuration.
- `context.Context` first argument on all I/O methods.
- Generics: `[K comparable, V any]` throughout.
- Each module has its own `README.md` (for humans) and `AGENTS.md` (for agents).
