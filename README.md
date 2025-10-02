# Semantic Cache

A high-performance semantic caching library for Go that uses vector embeddings to find semantically similar content. Perfect for LLM applications, search systems, and any use case where semantic similarity matters.

## Features

- 🚀 **Multiple Backend Support**: In-memory (LRU, LFU, FIFO) and Redis
- 🤖 **OpenAI Integration**: Built-in support for OpenAI's embedding models
- 🎯 **Semantic Search**: Find similar content using vector similarity
- ⚡ **High Performance**: Optimized similarity algorithms with async support
- 🛠️ **Extensible**: Pluggable backends and embedding providers
- 📦 **Type Safe**: Full generic support for any key/value types
- 🔄 **Context Aware**: Built-in context support for all operations
- 📊 **Batch Operations**: Efficient bulk operations (sync & async)
- ⚙️ **Async API**: Non-blocking operations with channel-based results

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/botirkhaltaev/semanticcache"
    "github.com/botirkhaltaev/semanticcache/options"
    "github.com/botirkhaltaev/semanticcache/similarity"
)

func main() {
    // Create a semantic cache with functional options
    cache, err := semanticcache.New[string, string](
        options.WithOpenAIProvider("your-api-key"),
        options.WithLRUBackend(1000), // 1000 items capacity
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // Store some data
    cache.Set(ctx, "user1", "The weather is sunny today", "Great day for outdoor activities")
    cache.Set(ctx, "user2", "It's raining heavily outside", "Perfect day to stay indoors")

    // Find semantically similar content
    match, err := cache.Lookup(ctx, "Nice weather outdoors", 0.8)
    if err != nil {
        log.Fatal(err)
    }
    if match != nil {
        fmt.Printf("Found similar content: %s (score: %.2f)\n", match.Value, match.Score)
    }

    // Get top similar matches
    matches, err := cache.TopMatches(ctx, "rainy day inside", 3)
    if err != nil {
        log.Fatal(err)
    }
    for _, match := range matches {
        fmt.Printf("Match: %s (score: %.2f)\n", match.Value, match.Score)
    }
}
```

## Installation

```bash
go get github.com/botirk38/semanticcache
```

## Configuration Options

### Backend Options

```go
// In-memory backends
options.WithLRUBackend(capacity)    // Least Recently Used
options.WithLFUBackend(capacity)    // Least Frequently Used  
options.WithFIFOBackend(capacity)   // First In, First Out

// Redis backend
options.WithRedisBackend("localhost:6379", 0) // addr, db

// Custom backend
options.WithCustomBackend(yourBackend)
```

### Embedding Provider Options

```go
// OpenAI (with optional model specification)
options.WithOpenAIProvider("api-key")
options.WithOpenAIProvider("api-key", "text-embedding-3-large")

// Custom provider
options.WithCustomProvider(yourProvider)
```

### Similarity Functions

```go
// Built-in similarity functions
options.WithSimilarityComparator(similarity.CosineSimilarity)        // Default
options.WithSimilarityComparator(similarity.EuclideanSimilarity)
options.WithSimilarityComparator(similarity.DotProductSimilarity)
options.WithSimilarityComparator(similarity.ManhattanSimilarity)
options.WithSimilarityComparator(similarity.PearsonCorrelationSimilarity)

// Custom similarity function
options.WithSimilarityComparator(func(a, b []float32) float32 {
    // Your custom similarity logic
    return similarity
})
```

## API Reference

### Core Operations

```go
// Basic CRUD operations
err := cache.Set(ctx, key, inputText, value)
value, found, err := cache.Get(ctx, key)
exists, err := cache.Contains(ctx, key)
err := cache.Delete(ctx, key)

// Cache management
err := cache.Flush(ctx)           // Clear all entries
count, err := cache.Len(ctx)      // Get cache size
err := cache.Close()              // Close resources
```

### Semantic Search

```go
// Find first match above threshold
match, err := cache.Lookup(ctx, "search text", 0.8)
if match != nil {
    fmt.Printf("Found: %v (score: %.2f)\n", match.Value, match.Score)
}

// Get top N matches
matches, err := cache.TopMatches(ctx, "search text", 5)
for _, match := range matches {
    fmt.Printf("Match: %v (score: %.2f)\n", match.Value, match.Score)
}
```

### Batch Operations

```go
// Batch set
items := []semanticcache.BatchItem[string, string]{
    {Key: "key1", InputText: "text1", Value: "value1"},
    {Key: "key2", InputText: "text2", Value: "value2"},
}
err := cache.SetBatch(ctx, items)

// Batch get
values, err := cache.GetBatch(ctx, []string{"key1", "key2"})

// Batch delete
err := cache.DeleteBatch(ctx, []string{"key1", "key2"})
```

### Async Operations

All cache operations have async variants that return channels for non-blocking execution:

```go
// Async set - returns immediately
errCh := cache.SetAsync(ctx, "key", "input text", "value")
// Do other work...
if err := <-errCh; err != nil {
    log.Printf("Set failed: %v", err)
}

// Async get
resultCh := cache.GetAsync(ctx, "key")
result := <-resultCh
if result.Error != nil {
    log.Printf("Get failed: %v", result.Error)
}
if result.Found {
    fmt.Printf("Value: %v\n", result.Value)
}

// Async semantic search
lookupCh := cache.LookupAsync(ctx, "search query", 0.8)
lookupResult := <-lookupCh
if lookupResult.Match != nil {
    fmt.Printf("Found: %v (score: %.2f)\n",
        lookupResult.Match.Value, lookupResult.Match.Score)
}

// Async batch operations for concurrent processing
errCh = cache.SetBatchAsync(ctx, items)
err = <-errCh

valuesCh := cache.GetBatchAsync(ctx, []string{"key1", "key2", "key3"})
batchResult := <-valuesCh
if batchResult.Error == nil {
    for key, value := range batchResult.Values {
        fmt.Printf("%s: %v\n", key, value)
    }
}

// Concurrent async operations
errCh1 := cache.SetAsync(ctx, "key1", "text1", "value1")
errCh2 := cache.SetAsync(ctx, "key2", "text2", "value2")
errCh3 := cache.SetAsync(ctx, "key3", "text3", "value3")

// Wait for all to complete
for _, ch := range []<-chan error{errCh1, errCh2, errCh3} {
    if err := <-ch; err != nil {
        log.Printf("Operation failed: %v", err)
    }
}
```

## Advanced Usage

### Custom Embedding Provider

```go
type MyProvider struct{}

func (p *MyProvider) EmbedText(text string) ([]float32, error) {
    // Your embedding logic
    return embedding, nil
}

func (p *MyProvider) Close() {}

// Use with cache
cache, err := semanticcache.New[string, string](
    options.WithCustomProvider(&MyProvider{}),
    options.WithLRUBackend(1000),
)
```

### Custom Backend

```go
type MyBackend struct{}

func (b *MyBackend) Set(ctx context.Context, key string, entry types.Entry[string]) error {
    // Your storage logic
    return nil
}

// Implement other required methods...

// Use with cache
cache, err := semanticcache.New[string, string](
    options.WithOpenAIProvider("api-key"),
    options.WithCustomBackend(&MyBackend{}),
)
```

### Redis Configuration

For Redis backend with JSON support (RedisJSON module):

```go
cache, err := semanticcache.New[string, MyStruct](
    options.WithOpenAIProvider("api-key"),
    options.WithRedisBackend("localhost:6379", 0),
)
```

## Performance Tips

1. **Choose the right backend**: LRU for time-locality, LFU for frequency-based access
2. **Batch operations**: Use `SetBatch`, `GetBatch` for multiple operations
3. **Adjust similarity threshold**: Higher thresholds = fewer, more precise matches
4. **Use context**: Enable request cancellation and timeouts

## Error Handling

The library returns descriptive errors for common issues:

```go
cache, err := semanticcache.New[string, string](
    options.WithOpenAIProvider("invalid-key"),
    options.WithLRUBackend(1000),
)
if err != nil {
    // Handle configuration errors
    log.Printf("Cache creation failed: %v", err)
}

match, err := cache.Lookup(ctx, "query", 0.8)
if err != nil {
    // Handle runtime errors (network, API limits, etc.)
    log.Printf("Lookup failed: %v", err)
}
```

## Examples

Check out the `examples/` directory for complete examples:

- Basic usage with OpenAI
- Custom similarity functions
- Redis backend setup
- Batch operations
- LLM response caching

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.