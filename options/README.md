# options

Functional options for configuring a semantic cache instance.

## Usage

Pass options to `semanticcache.New`:

```go
cache, err := semanticcache.New[string, string](
    options.WithLRUBackend[string, string](1000),
    options.WithOpenAIProvider[string, string]("api-key"),
    options.WithSimilarityComparator[string, string](similarity.CosineSimilarity),
)
```

## Available options

### Backends

| Option | Description |
|--------|-------------|
| `WithLRUBackend(capacity)` | LRU eviction |
| `WithLFUBackend(capacity)` | LFU eviction |
| `WithFIFOBackend(capacity)` | FIFO eviction |
| `WithRedisBackend(addr, opts...)` | Redis with JSON storage |
| `WithCustomBackend(backend)` | Any `types.Backend` implementation |

### Providers

| Option | Description |
|--------|-------------|
| `WithOpenAIProvider(apiKey, model...)` | OpenAI embeddings (default: text-embedding-3-small) |
| `WithCustomProvider(provider)` | Any `types.EmbeddingProvider` implementation |

### Similarity

| Option | Description |
|--------|-------------|
| `WithSimilarityComparator(fn)` | Custom similarity function (default: cosine) |

## Errors

- `ErrNilBackend` -- nil backend provided
- `ErrNilProvider` -- nil provider provided
- `ErrNilComparator` -- nil similarity function provided
