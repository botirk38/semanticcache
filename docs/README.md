# SemanticCache Documentation

Welcome to the SemanticCache documentation! This guide will help you get the most out of this high-performance semantic caching library for Go.

## 📚 Documentation Index

### Getting Started

- **[Main README](../README.md)** - Quick start guide and basic usage
- **[Redis Setup Guide](redis-setup.md)** - Detailed Redis backend configuration

### Core Documentation

#### [API Reference](api-reference.md)
Complete API documentation covering:
- Core types and interfaces
- Synchronous operations
- Asynchronous operations
- Batch operations
- Configuration options
- Error handling

#### [Architecture & Design](architecture.md)
Deep dive into the library's architecture:
- High-level architecture overview
- Design patterns used
- Component details (backends, providers, similarity)
- Data flow diagrams
- Concurrency model
- Extension points

#### [Performance Guide](performance.md)
Optimization and best practices:
- Performance characteristics
- Backend selection guide
- Optimization strategies
- Async vs sync operations
- Best practices
- Common pitfalls
- Benchmarking guide

#### [Examples & Tutorials](examples.md)
Practical code examples:
- Basic usage examples (FAQ, LLM caching, product search)
- Advanced use cases (async batch processing, multi-similarity)
- Integration examples (HTTP API server)
- Custom implementations (Ollama provider, SQLite backend, monitoring)

---

## Quick Navigation

### By Topic

**Getting Started**
- [Installation](#installation)
- [Quick Start Example](#quick-start)
- [Configuration Options](#configuration)

**Core Concepts**
- [Semantic Search](api-reference.md#semantic-search)
- [Similarity Functions](architecture.md#similarity-functions)
- [Backends](architecture.md#backend-system)
- [Embedding Providers](architecture.md#provider-openai)

**Operations**
- [Basic CRUD](api-reference.md#synchronous-operations)
- [Async Operations](api-reference.md#asynchronous-operations)
- [Batch Operations](api-reference.md#batch-operations)

**Advanced**
- [Custom Backends](examples.md#example-8-custom-backend-sqlite)
- [Custom Providers](examples.md#example-7-custom-embedding-provider-ollama)
- [Performance Tuning](performance.md#optimization-strategies)

### By Use Case

**I want to...**

- **Cache LLM responses** → [LLM Caching Example](examples.md#example-2-llm-response-caching)
- **Build a FAQ system** → [FAQ Example](examples.md#example-1-simple-faq-system)
- **Search products semantically** → [Product Search Example](examples.md#example-3-product-search-with-semantic-matching)
- **Process large datasets** → [Async Batch Example](examples.md#example-4-async-batch-processing)
- **Deploy to production** → [Performance Guide](performance.md#best-practices)
- **Use with Redis** → [Redis Setup](redis-setup.md)
- **Add monitoring** → [Monitoring Example](examples.md#example-9-monitoring-wrapper)
- **Understand internals** → [Architecture](architecture.md)

---

## Installation

```bash
go get github.com/botirk38/semanticcache
```

**Requirements:**
- Go 1.18+ (for generics support)
- Optional: Redis 7+ with RedisJSON module

---

## Quick Start

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
    // Create semantic cache
    cache, err := semanticcache.New[string, string](
        options.WithOpenAIProvider("your-api-key"),
        options.WithLRUBackend(1000),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // Store data
    cache.Set(ctx, "q1", "How do I reset my password?", "Go to Settings > Reset Password")

    // Semantic search
    match, err := cache.Lookup(ctx, "password reset help", 0.8)
    if match != nil {
        fmt.Printf("Found: %s (score: %.2f)\n", match.Value, match.Score)
    }
}
```

---

## Configuration

### Backend Options

```go
// In-memory backends
options.WithLRUBackend(capacity)    // Least Recently Used
options.WithLFUBackend(capacity)    // Least Frequently Used
options.WithFIFOBackend(capacity)   // First In, First Out

// Redis backend
options.WithRedisBackend("localhost:6379", 0)

// Custom backend
options.WithCustomBackend(yourBackend)
```

See [Backend Selection Guide](performance.md#backend-selection) for choosing the right backend.

### Provider Options

```go
// OpenAI (default model: text-embedding-3-small)
options.WithOpenAIProvider("api-key")
options.WithOpenAIProvider("api-key", "text-embedding-3-large")

// Custom provider
options.WithCustomProvider(yourProvider)
```

See [Custom Provider Example](examples.md#example-7-custom-embedding-provider-ollama) for implementing custom providers.

### Similarity Options

```go
// Built-in similarity functions
options.WithSimilarityComparator(similarity.CosineSimilarity)        // Default, [-1,1]
options.WithSimilarityComparator(similarity.EuclideanSimilarity)     // [0,1]
options.WithSimilarityComparator(similarity.DotProductSimilarity)    // Unbounded
options.WithSimilarityComparator(similarity.ManhattanSimilarity)     // [0,1]
options.WithSimilarityComparator(similarity.PearsonCorrelationSimilarity)  // [-1,1]
```

See [Similarity Functions](architecture.md#similarity-functions) for details on each algorithm.

---

## Key Features

### 🔍 Semantic Search

Find semantically similar content, not just exact matches:

```go
// Find first match above threshold
match, err := cache.Lookup(ctx, "query text", 0.8)

// Get top N matches, sorted by similarity
matches, err := cache.TopMatches(ctx, "query text", 5)
```

[Learn more about semantic search →](api-reference.md#semantic-search)

### ⚡ Async Operations

Non-blocking operations for better performance:

```go
// Returns immediately, executes in background
errCh := cache.SetAsync(ctx, key, text, value)

// Do other work...

// Wait for result when needed
if err := <-errCh; err != nil {
    log.Printf("Error: %v", err)
}
```

[Learn more about async operations →](api-reference.md#asynchronous-operations)

### 📦 Batch Operations

Efficient bulk operations:

```go
// Sync batch
items := []semanticcache.BatchItem[string, string]{...}
cache.SetBatch(ctx, items)

// Async batch (parallel processing)
errCh := cache.SetBatchAsync(ctx, items)
```

[Learn more about batch operations →](api-reference.md#batch-operations)

### 🔌 Extensible

Implement custom backends, providers, or similarity functions:

```go
// Custom backend
type MyBackend struct{}
func (b *MyBackend) Set(...) error { /* ... */ }
// Implement other methods...

cache := semanticcache.New[string, string](
    options.WithCustomBackend(&MyBackend{}),
    options.WithOpenAIProvider("key"),
)
```

[See custom implementation examples →](examples.md#custom-implementations)

---

## Performance Tips

### 1. Choose the Right Backend

- **LRU**: Time-based access patterns (recent items)
- **LFU**: Frequency-based patterns (popular items)
- **Redis**: Shared cache, persistence, large datasets

[Detailed backend comparison →](performance.md#backend-selection)

### 2. Use Async for Concurrency

```go
// Concurrent operations
errCh1 := cache.SetAsync(ctx, "key1", "text1", "value1")
errCh2 := cache.SetAsync(ctx, "key2", "text2", "value2")

// Wait for all
for _, ch := range []<-chan error{errCh1, errCh2} {
    <-ch
}
```

[Async best practices →](performance.md#async-vs-sync-operations)

### 3. Batch Operations

```go
// Instead of N API calls
for _, item := range items {
    cache.Set(ctx, item.Key, item.Text, item.Value)  // Slow
}

// Use batch (1 operation)
cache.SetBatch(ctx, batchItems)  // Fast
```

[Optimization strategies →](performance.md#optimization-strategies)

### 4. Tune Similarity Threshold

```go
// High precision (exact matches)
match, _ := cache.Lookup(ctx, query, 0.95)

// Balanced (recommended)
match, _ := cache.Lookup(ctx, query, 0.85)

// High recall (fuzzy matches)
match, _ := cache.Lookup(ctx, query, 0.75)
```

[Threshold selection guide →](performance.md#optimize-threshold-selection)

---

## Common Use Cases

### LLM Response Caching

Avoid redundant LLM API calls by caching responses:

```go
match, _ := cache.Lookup(ctx, userPrompt, 0.85)
if match != nil {
    return match.Value  // Cache hit
}

// Cache miss - call LLM
response := callLLM(userPrompt)
cache.Set(ctx, promptID, userPrompt, response)
return response
```

[Full LLM caching example →](examples.md#example-2-llm-response-caching)

### FAQ Systems

Semantic FAQ lookup:

```go
// Index FAQs
cache.Set(ctx, "faq1", "How do I reset my password?", answer1)
cache.Set(ctx, "faq2", "How do I contact support?", answer2)

// User query (different wording)
match, _ := cache.Lookup(ctx, "password reset help", 0.8)
// Returns answer1
```

[Full FAQ example →](examples.md#example-1-simple-faq-system)

### Product Search

Natural language product search:

```go
// Index products by description
cache.Set(ctx, productID, productDescription, product)

// Natural language search
matches, _ := cache.TopMatches(ctx, "laptop for programming", 5)
```

[Full product search example →](examples.md#example-3-product-search-with-semantic-matching)

---

## Troubleshooting

### Common Issues

**Q: "key cannot be zero value" error**

A: Ensure keys are not empty/zero values:
```go
if key == "" {
    return errors.New("invalid key")
}
cache.Set(ctx, key, text, value)
```

**Q: Lookup returns nil unexpectedly**

A: Check threshold and similarity function:
```go
// Try lower threshold
match, _ := cache.Lookup(ctx, query, 0.7)  // Instead of 0.9

// Or different similarity function
options.WithSimilarityComparator(similarity.CosineSimilarity)
```

**Q: Redis connection errors**

A: Verify Redis is running and accessible:
```bash
redis-cli ping  # Should return PONG
```

See [Redis Setup Guide](redis-setup.md) for detailed configuration.

**Q: Slow semantic search**

A: Limit cache size or use Redis with vector indexing:
```go
// Limit cache size
options.WithLRUBackend(10000)  // Not 1000000

// Or use Redis (future: vector indexing)
options.WithRedisBackend("localhost:6379", 0)
```

See [Performance Guide](performance.md) for optimization tips.

---

## API Quick Reference

### Cache Operations

```go
// Basic CRUD
err := cache.Set(ctx, key, inputText, value)
value, found, err := cache.Get(ctx, key)
exists, err := cache.Contains(ctx, key)
err := cache.Delete(ctx, key)

// Semantic search
match, err := cache.Lookup(ctx, inputText, threshold)
matches, err := cache.TopMatches(ctx, inputText, n)

// Management
err := cache.Flush(ctx)
count, err := cache.Len(ctx)
err := cache.Close()
```

[Full API reference →](api-reference.md)

### Async Operations

```go
errCh := cache.SetAsync(ctx, key, text, value)
resultCh := cache.GetAsync(ctx, key)
errCh := cache.DeleteAsync(ctx, key)
resultCh := cache.LookupAsync(ctx, text, threshold)
resultCh := cache.TopMatchesAsync(ctx, text, n)
```

[Async API reference →](api-reference.md#asynchronous-operations)

### Batch Operations

```go
items := []semanticcache.BatchItem[K, V]{...}
err := cache.SetBatch(ctx, items)
values, err := cache.GetBatch(ctx, keys)
err := cache.DeleteBatch(ctx, keys)

// Async variants
errCh := cache.SetBatchAsync(ctx, items)
resultCh := cache.GetBatchAsync(ctx, keys)
errCh := cache.DeleteBatchAsync(ctx, keys)
```

[Batch API reference →](api-reference.md#batch-operations)

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

**Areas for contribution:**
- Vector indexing for faster search
- Additional embedding providers (Anthropic, Cohere, HuggingFace)
- More backends (PostgreSQL with pgvector, etc.)
- Performance benchmarks
- Documentation improvements

---

## Community & Support

- **Issues**: [GitHub Issues](https://github.com/botirk38/semanticcache/issues)
- **Discussions**: [GitHub Discussions](https://github.com/botirk38/semanticcache/discussions)
- **Examples**: [docs/examples.md](examples.md)

---

## License

MIT License - see [LICENSE](../LICENSE) for details.

---

## Changelog

See [Releases](https://github.com/botirk38/semanticcache/releases) for version history and updates.

---

**Happy caching! 🚀**
