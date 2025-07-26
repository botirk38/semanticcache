# SemanticCache

A Go library for semantic caching that stores and retrieves values by meaning, not just exact string matches. Perfect for AI applications, chatbots, and retrieval-augmented generation.

## Quick Start

```sh
go get github.com/botirk38/semanticcache
```

```go
import (
    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/backends"
    "github.com/botirk38/semanticcache/providers/openai"
    "github.com/botirk38/semanticcache/types"
)

// 1. Create backend and provider
backend, _ := backends.NewLRUBackend[string, string](1000)
provider, _ := openai.NewOpenAIProvider(openai.OpenAIConfig{
    APIKey: "your-openai-key",
    Model:  "text-embedding-3-small",
})

// 2. Create cache
cache, _ := semanticcache.NewSemanticCache(backend, provider, nil)
defer cache.Close()

// 3. Use semantic caching
cache.Set("q1", "What is the capital of France?", "Paris")

// Find similar questions
value, found, _ := cache.Lookup("French capital city?", 0.8)
if found {
    fmt.Println(value) // "Paris"
}
```

## Features

- **Semantic Search**: Find cached values by meaning, not exact keys
- **Multiple Backends**: In-memory (LRU/FIFO/LFU) and Redis support  
- **OpenAI Integration**: Built-in support for OpenAI embeddings
- **Type Safe**: Full generics support for keys and values
- **Pluggable**: Easy to extend with new backends and embedding providers

## Backends

Choose from multiple storage options:

```go
// In-memory LRU cache (fastest)
backend, _ := backends.NewLRUBackend[string, string](1000)

// Redis backend (persistent)
config := types.BackendConfig{
    ConnectionString: "localhost:6379",
    Database: 0,
}
factory := &backends.BackendFactory[string, string]{}
backend, _ := factory.NewBackend(types.BackendRedis, config)
```

## API

```go
// Store with semantic text
cache.Set("key1", "What is the capital of France?", "Paris")

// Get by exact key
value, found := cache.Get("key1")

// Find by semantic similarity
value, found, _ := cache.Lookup("French capital?", 0.8)

// Get top matches ranked by similarity
matches, _ := cache.TopMatches("European capitals", 5)
```

## Development

```bash
# Run tests
go test -v ./...

# Run benchmarks  
go test -v ./tests/benchmarks -bench=.
```

## License

MIT License

---

*Built with ❤️ for the Go community*
