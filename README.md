# semanticcache

A Go library for **semantic caching** with pluggable backends and embedding providers, supporting vector-based similarity search for AI applications.
Useful for AI, retrieval-augmented generation, chatbot memory, and any workload where you want to cache by semantic similarity—not just string keys.

---

## Project Overview

`semanticcache` is a pluggable semantic cache for Go with a scalable architecture supporting multiple storage backends and embedding providers.
It lets you cache arbitrary values indexed by **embeddings** (vector representations of text, etc.), with support for various eviction policies and storage systems.

---

## Key Features

- **Semantic Cache:** Store and retrieve values by semantic similarity, not just by string key.
- **Pluggable Backends:** Support for in-memory (LRU, FIFO, LFU) and remote backends (Redis).
- **Pluggable Embedding Providers:** OpenAI, with architecture ready for Anthropic, Ollama, and more.
- **Scalable Architecture:** Clean separation of concerns with backend and provider interfaces.
- **Fast Similarity Search:** Lookup or rank by vector similarity (defaults to cosine).
- **Type Safety:** Full generics support for keys and values.
- **Context Support:** Context-aware operations for timeouts and cancellation.

---

## Installation

```sh
go get github.com/botirk38/semanticcache
```

---

## Architecture

The library is built with a pluggable architecture that separates concerns:

- **Cache Layer:** High-level semantic operations (`SemanticCache`)
- **Backend Layer:** Storage implementations (in-memory, Redis, etc.)
- **Provider Layer:** Embedding generation (OpenAI, future providers)
- **Types Layer:** Shared interfaces and types

```
┌─────────────────┐
│  SemanticCache  │  ← High-level API
├─────────────────┤
│  CacheBackend   │  ← Storage interface (LRU/FIFO/LFU/Redis)
├─────────────────┤
│EmbeddingProvider│  ← Embedding interface (OpenAI/Anthropic/etc)
└─────────────────┘
```

---

## Usage

### 1. Create a backend

#### **In-Memory Backends**

```go
import (
    "github.com/botirk38/semanticcache/backends"
    "github.com/botirk38/semanticcache/types"
)

// LRU backend
config := types.BackendConfig{Capacity: 1000}
factory := &backends.BackendFactory[string, string]{}
backend, err := factory.NewBackend(types.BackendLRU, config)
if err != nil {
    panic(err)
}
defer backend.Close()

// Other options: BackendFIFO, BackendLFU
```

#### **Redis Backend**

```go
config := types.BackendConfig{
    ConnectionString: "localhost:6379",
    Database:         0,
    // Username/Password if needed
}
backend, err := factory.NewBackend(types.BackendRedis, config)
if err != nil {
    panic(err)
}
defer backend.Close()
```

### 2. Create an embedding provider

```go
import "github.com/botirk38/semanticcache/providers/openai"

config := openai.OpenAIConfig{
    APIKey: "your-api-key",
    Model:  "text-embedding-3-small", // or text-embedding-ada-002
}
provider, err := openai.NewOpenAIProvider(config)
if err != nil {
    panic(err)
}
defer provider.Close()
```

### 3. Create your semantic cache

```go
import "github.com/botirk38/semanticcache"

cache, err := semanticcache.NewSemanticCache(backend, provider, nil)
if err != nil {
    panic(err)
}
defer cache.Close()
```

### 4. Use the cache

#### Store and retrieve by key

```go
ctx := context.Background()

// Store a value
err := cache.Set("france-capital", "What is the capital of France?", "Paris")
if err != nil {
    panic(err)
}

// Retrieve by exact key
value, ok := cache.Get("france-capital")
if ok {
    fmt.Println(value) // "Paris"
}
```

#### Semantic search

```go
// Find semantically similar entries
value, ok, err := cache.Lookup("Which city is France's capital?", 0.8)
if err != nil {
    panic(err)
}
if ok {
    fmt.Println(value) // "Paris" (if similarity > 0.8)
}

// Get top N matches
matches, err := cache.TopMatches("french capital city", 3)
if err != nil {
    panic(err)
}
for _, match := range matches {
    fmt.Printf("Value: %v, Score: %.3f\n", match.Value, match.Score)
}
```

---

## API Overview

### Core Types

- `SemanticCache[K, V]`: The main cache with generics support
- `CacheBackend[K, V]`: Interface for storage backends
- `EmbeddingProvider`: Interface for embedding generation
- `Entry[V]`: Holds embedding and value
- `Match[V]`: A value and similarity score for ranked results

### Backend Types

- `BackendLRU`: Least Recently Used eviction
- `BackendFIFO`: First In, First Out eviction  
- `BackendLFU`: Least Frequently Used eviction
- `BackendRedis`: Redis with vector search

### SemanticCache Methods

- `Set(key K, text string, value V) error`
- `Get(key K) (V, bool)`
- `Lookup(text string, threshold float32) (V, bool, error)`
- `TopMatches(text string, n int) ([]Match[V], error)`
- `Flush() error`
- `Len() int`
- `Close()`

### Backend Interface

```go
type CacheBackend[K comparable, V any] interface {
    Set(ctx context.Context, key K, entry Entry[V]) error
    Get(ctx context.Context, key K) (Entry[V], bool, error)
    Delete(ctx context.Context, key K) error
    Contains(ctx context.Context, key K) (bool, error)
    Flush(ctx context.Context) error
    Len(ctx context.Context) (int, error)
    Keys(ctx context.Context) ([]K, error)
    GetEmbedding(ctx context.Context, key K) ([]float32, bool, error)
    Close() error
}
```

---

## Testing

The project includes comprehensive tests organized in the `tests/` directory:

### Run All Tests

```bash
go test -v ./...
```

### Test Categories

#### Backend Tests
```bash
# Test in-memory backends (LRU, FIFO, LFU)
go test -v ./tests/backends

# Test specific backend
go test -v ./tests/backends -run TestLRUBackend
```

#### Provider Tests
```bash
# Test embedding providers
go test -v ./tests/providers

# Test with mocks (no API key required)
go test -v ./tests/providers -run TestOpenAIProvider
```

#### Integration Tests
```bash
# End-to-end semantic cache tests
go test -v ./tests/integration
```

#### Benchmark Tests
```bash
# Performance benchmarks
go test -v ./tests/benchmarks -bench=.

# Specific benchmarks
go test -v ./tests/benchmarks -bench=BenchmarkCacheSet
go test -v ./tests/benchmarks -bench=BenchmarkCacheLookup
```

#### Redis Integration Tests
```bash
# Requires Redis server running on localhost:6379
go test -v ./tests/backends -run TestRedisBackend
```

### Test Structure

```
tests/
├── backends/           # Backend implementation tests
│   ├── inmemory_test.go
│   └── redis_test.go
├── providers/          # Provider tests with mocks
│   └── openai_test.go
├── integration/        # End-to-end tests
│   └── cache_test.go
└── benchmarks/         # Performance tests
    └── cache_bench_test.go
```

---

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Write tests and make your changes.
4. Open a pull request!

---

## License

MIT License – see LICENSE file.

---

## Dependencies

### Core Dependencies
- [github.com/hashicorp/golang-lru/v2](https://github.com/hashicorp/golang-lru) – LRU cache implementation
- [github.com/openai/openai-go](https://github.com/openai/openai-go) – OpenAI embedding provider
- [github.com/redis/go-redis/v9](https://github.com/redis/go-redis) – Redis client for remote backend

### Development Dependencies  
- [github.com/stretchr/testify](https://github.com/stretchr/testify) – Testing framework and mocks
- [github.com/alicebob/miniredis/v2](https://github.com/alicebob/miniredis) – In-memory Redis for testing

---

## Compatibility

- Requires Go 1.18 or newer.
- Supports Linux, macOS, Windows (embedding provider support may vary).

---

## Extending the Library

### Adding New Backends

1. Implement the `CacheBackend[K, V]` interface in `backends/`
2. Add your backend type to `types/types.go`
3. Update the factory in `backends/backends.go`
4. Add comprehensive tests in `tests/backends/`

### Adding New Providers

1. Create a new package in `providers/`
2. Implement the `EmbeddingProvider` interface
3. Use provider-specific configuration (not generic)
4. Add tests with mocks in `tests/providers/`

### Provider Roadmap

- ✅ OpenAI (`text-embedding-3-small`, `text-embedding-ada-002`)
- 🔄 Anthropic (planned)
- 🔄 Ollama (planned)
- 🔄 Cohere (planned)
- 🔄 HuggingFace (planned)

---

**Questions?** Open an issue or discussion on GitHub!
**Need a new provider or backend? PRs welcome!**
