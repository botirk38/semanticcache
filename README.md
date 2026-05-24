# SemanticCache

A Go library for semantic caching. Store values alongside their embedding vectors and retrieve them by meaning rather than exact key match.

Useful for LLM response caching, search deduplication, and any system where you want to find "close enough" matches.

## Install

```
go get github.com/botirk38/semanticcache
```

Requires Go 1.21+.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
)

func main() {
    cache, err := semanticcache.New[string, string](
        options.WithLRUBackend[string, string](1000),
        options.WithOpenAIProvider[string, string]("your-api-key"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()

    ctx := context.Background()
    _ = cache.Set(ctx, "greeting", "Hello, how are you?", "I'm doing well, thanks!")

    match, _ := cache.Lookup(ctx, "Hi, how's it going?", 0.8)
    if match != nil {
        fmt.Printf("Hit: %s (score: %.2f)\n", match.Value, match.Score)
    }
}
```

## API

All methods take `context.Context` as the first argument.

### Core operations

| Method | Description |
|--------|-------------|
| `Set(ctx, key, inputText, value)` | Store a value. The embedding is computed from `inputText`. |
| `Get(ctx, key)` | Retrieve by exact key. Returns `(value, found, error)`. |
| `Delete(ctx, key)` | Remove an entry. |
| `Contains(ctx, key)` | Check if a key exists. |
| `Flush(ctx)` | Remove all entries. |
| `Len(ctx)` | Count of stored entries. |
| `Close()` | Release backend and provider resources. |

### Semantic search

| Method | Description |
|--------|-------------|
| `Lookup(ctx, text, threshold)` | Best match above the similarity threshold. Returns `nil` if nothing qualifies. |
| `TopMatches(ctx, text, n)` | Top `n` matches sorted by descending similarity. |

### Batch operations

| Method | Description |
|--------|-------------|
| `SetBatch(ctx, items)` | Store multiple items. |
| `GetBatch(ctx, keys)` | Retrieve multiple values. Missing keys are omitted. |
| `DeleteBatch(ctx, keys)` | Remove multiple entries. |

## Configuration

Configuration uses functional options passed to `semanticcache.New`.

### Backends

```go
options.WithLRUBackend[K, V](capacity)           // Least Recently Used
options.WithLFUBackend[K, V](capacity)           // Least Frequently Used
options.WithFIFOBackend[K, V](capacity)          // First In, First Out
options.WithRedisBackend[K, V](addr, redisOpts...)  // Redis (JSON storage)
options.WithCustomBackend[K, V](backend)         // Your own Backend implementation
```

Redis options: `remote.WithPassword`, `remote.WithDB`, `remote.WithPrefix`, `remote.WithUsername`, `remote.WithTLS`.

### Embedding providers

```go
options.WithOpenAIProvider[K, V]("api-key")               // text-embedding-3-small (default)
options.WithOpenAIProvider[K, V]("api-key", "model-name")  // custom model
options.WithCustomProvider[K, V](provider)                 // your own EmbeddingProvider
```

A local hash-based provider is available for testing (not semantically meaningful):

```go
import "github.com/botirk38/semanticcache/providers/local"

provider := local.New(128)  // 128-dimensional vectors
```

### Similarity functions

```go
options.WithSimilarityComparator[K, V](similarity.CosineSimilarity)           // default
options.WithSimilarityComparator[K, V](similarity.EuclideanSimilarity)
options.WithSimilarityComparator[K, V](similarity.DotProductSimilarity)
options.WithSimilarityComparator[K, V](similarity.ManhattanSimilarity)
options.WithSimilarityComparator[K, V](similarity.PearsonCorrelationSimilarity)
```

## Architecture

```
semanticcache/         Root package: Cache type + constructors
  options/             Functional options (WithLRUBackend, WithOpenAIProvider, etc.)
  types/               Backend and EmbeddingProvider interfaces
  backends/
    inmemory/          LRU, LFU, FIFO backends
    remote/            Redis backend
  providers/
    openai/            OpenAI embedding provider
    local/             Hash-based provider for testing
  similarity/          Cosine, Euclidean, DotProduct, Manhattan, Pearson
  chunker/             Text chunking utilities
  tokenizer/           Token counting (OpenAI, Anthropic, Gemini)
```

The `Backend[K, V]` interface (9 methods) is in `types/`. Any type implementing it can be used as a cache backend. `EmbeddingProvider` (2 methods: `EmbedText`, `Close`) turns text into vectors.

## Implementing a custom backend

Implement `types.Backend[K, V]`:

```go
type Backend[K comparable, V any] interface {
    Set(ctx context.Context, key K, embedding []float64, value V) error
    Get(ctx context.Context, key K) (V, bool, error)
    Delete(ctx context.Context, key K) error
    Contains(ctx context.Context, key K) (bool, error)
    Keys(ctx context.Context) ([]K, error)
    GetEmbedding(ctx context.Context, key K) ([]float64, bool, error)
    Flush(ctx context.Context) error
    Len(ctx context.Context) (int, error)
    Close() error
}
```

## Implementing a custom provider

Implement `types.EmbeddingProvider`:

```go
type EmbeddingProvider interface {
    EmbedText(ctx context.Context, text string) ([]float64, error)
    Close() error
}
```

Optionally implement `types.BatchEmbeddingProvider` for batch support.

## Development

```
go test ./...           # run all tests
go test -race ./...     # with race detector
go test -bench=. ./...  # run benchmarks
go vet ./...            # static analysis
gofmt -l .              # check formatting
```

## License

MIT
