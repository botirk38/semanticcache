# API Reference

Complete API reference for SemanticCache library.

## Table of Contents

- [Core Types](#core-types)
- [Cache Creation](#cache-creation)
- [Synchronous Operations](#synchronous-operations)
- [Asynchronous Operations](#asynchronous-operations)
- [Batch Operations](#batch-operations)
- [Configuration Options](#configuration-options)
- [Interfaces](#interfaces)

---

## Core Types

### SemanticCache[K, V]

The main cache type with generic key and value types.

```go
type SemanticCache[K comparable, V any] struct {
    // Internal fields (not exported)
}
```

**Type Parameters:**
- `K comparable`: Key type (must be comparable)
- `V any`: Value type (can be any type)

### Match[V]

Represents a semantic search result with similarity score.

```go
type Match[V any] struct {
    Value V       `json:"value"`  // The cached value
    Score float32 `json:"score"`  // Similarity score
}
```

### BatchItem[K, V]

Item for batch operations.

```go
type BatchItem[K comparable, V any] struct {
    Key       K      // Cache key
    InputText string // Text to embed
    Value     V      // Value to cache
}
```

### Result Types (Async Operations)

```go
// Result from GetAsync
type GetResult[V any] struct {
    Value V
    Found bool
    Error error
}

// Result from LookupAsync
type LookupResult[V any] struct {
    Match *Match[V]
    Error error
}

// Result from TopMatchesAsync
type TopMatchesResult[V any] struct {
    Matches []Match[V]
    Error   error
}

// Result from GetBatchAsync
type GetBatchResult[K comparable, V any] struct {
    Values map[K]V
    Error  error
}
```

---

## Cache Creation

### New

Creates a new semantic cache with functional options (recommended).

```go
func New[K comparable, V any](opts ...options.Option[K, V]) (*SemanticCache[K, V], error)
```

**Parameters:**
- `opts`: Variable number of configuration options

**Returns:**
- `*SemanticCache[K, V]`: Configured cache instance
- `error`: Configuration or validation error

**Example:**
```go
cache, err := semanticcache.New[string, string](
    options.WithOpenAIProvider("api-key"),
    options.WithLRUBackend(1000),
    options.WithSimilarityComparator(similarity.CosineSimilarity),
)
```

**Errors:**
- Returns error if required options missing (backend, provider)
- Returns error if option application fails

### NewSemanticCache

Low-level constructor (use `New` instead for better API).

```go
func NewSemanticCache[K comparable, V any](
    backend types.CacheBackend[K, V],
    provider types.EmbeddingProvider,
    comparator similarity.SimilarityFunc,
) (*SemanticCache[K, V], error)
```

**Parameters:**
- `backend`: Storage backend implementation
- `provider`: Embedding provider implementation
- `comparator`: Similarity function (nil uses CosineSimilarity)

**Returns:**
- `*SemanticCache[K, V]`: Cache instance
- `error`: Validation error if any parameter is nil

---

## Synchronous Operations

### Set

Stores a value with its semantic embedding.

```go
func (sc *SemanticCache[K, V]) Set(
    ctx context.Context,
    key K,
    inputText string,
    value V,
) error
```

**Parameters:**
- `ctx`: Context for cancellation/timeout
- `key`: Cache key (cannot be zero value)
- `inputText`: Text to generate embedding from
- `value`: Value to cache

**Returns:**
- `error`: Embedding generation or storage error

**Example:**
```go
err := cache.Set(ctx, "user123", "How do I reset my password?", answerStruct)
```

**Errors:**
- Returns error if `key` is zero value
- Returns error if embedding generation fails
- Returns error if backend storage fails

### Get

Retrieves a value by key.

```go
func (sc *SemanticCache[K, V]) Get(
    ctx context.Context,
    key K,
) (V, bool, error)
```

**Parameters:**
- `ctx`: Context for cancellation/timeout
- `key`: Cache key

**Returns:**
- `V`: Value (zero value if not found)
- `bool`: `true` if found, `false` otherwise
- `error`: Retrieval error

**Example:**
```go
value, found, err := cache.Get(ctx, "user123")
if found {
    fmt.Printf("Value: %v\n", value)
}
```

### Lookup

Finds the first semantically similar entry above threshold.

```go
func (sc *SemanticCache[K, V]) Lookup(
    ctx context.Context,
    inputText string,
    threshold float32,
) (*Match[V], error)
```

**Parameters:**
- `ctx`: Context for cancellation/timeout
- `inputText`: Query text to find similar content
- `threshold`: Minimum similarity score `[0, 1]`

**Returns:**
- `*Match[V]`: First match above threshold (nil if none found)
- `error`: Embedding or search error

**Example:**
```go
match, err := cache.Lookup(ctx, "password reset help", 0.8)
if match != nil {
    fmt.Printf("Found: %v (score: %.2f)\n", match.Value, match.Score)
}
```

**Notes:**
- Returns best match (highest score) >= threshold
- Threshold interpretation depends on similarity function
- Cosine similarity: 0.8-0.9 is typical for semantic matches

### TopMatches

Returns top N semantically similar entries, sorted by descending score.

```go
func (sc *SemanticCache[K, V]) TopMatches(
    ctx context.Context,
    inputText string,
    n int,
) ([]Match[V], error)
```

**Parameters:**
- `ctx`: Context for cancellation/timeout
- `inputText`: Query text
- `n`: Maximum number of matches (must be > 0)

**Returns:**
- `[]Match[V]`: Matches sorted by score (descending)
- `error`: Validation, embedding, or search error

**Example:**
```go
matches, err := cache.TopMatches(ctx, "account issues", 5)
for _, match := range matches {
    fmt.Printf("Score %.2f: %v\n", match.Score, match.Value)
}
```

**Errors:**
- Returns error if `n <= 0`

### Contains

Checks if a key exists without retrieving the value.

```go
func (sc *SemanticCache[K, V]) Contains(
    ctx context.Context,
    key K,
) (bool, error)
```

**Parameters:**
- `ctx`: Context
- `key`: Cache key

**Returns:**
- `bool`: `true` if exists, `false` otherwise
- `error`: Check error

**Example:**
```go
exists, err := cache.Contains(ctx, "user123")
```

### Delete

Removes an entry from the cache.

```go
func (sc *SemanticCache[K, V]) Delete(
    ctx context.Context,
    key K,
) error
```

**Parameters:**
- `ctx`: Context
- `key`: Key to delete

**Returns:**
- `error`: Deletion error

**Example:**
```go
err := cache.Delete(ctx, "user123")
```

### Flush

Clears all entries from the cache.

```go
func (sc *SemanticCache[K, V]) Flush(ctx context.Context) error
```

**Parameters:**
- `ctx`: Context

**Returns:**
- `error`: Flush error

**Example:**
```go
err := cache.Flush(ctx)
```

**Warning:** This operation is irreversible.

### Len

Returns the number of entries in the cache.

```go
func (sc *SemanticCache[K, V]) Len(ctx context.Context) (int, error)
```

**Parameters:**
- `ctx`: Context

**Returns:**
- `int`: Number of entries
- `error`: Count error

**Example:**
```go
count, err := cache.Len(ctx)
fmt.Printf("Cache size: %d\n", count)
```

### Close

Closes the cache and releases resources.

```go
func (sc *SemanticCache[K, V]) Close() error
```

**Returns:**
- `error`: Close error (if any)

**Example:**
```go
defer cache.Close()
```

**Notes:**
- Closes embedding provider
- Backend cleanup (connection pools, etc.)
- Should be called when cache no longer needed

---

## Asynchronous Operations

All async operations return buffered channels (size 1) that deliver results when complete. Operations are non-blocking and execute in goroutines.

### SetAsync

Asynchronously stores a value with embedding.

```go
func (sc *SemanticCache[K, V]) SetAsync(
    ctx context.Context,
    key K,
    inputText string,
    value V,
) <-chan error
```

**Parameters:**
- Same as `Set`

**Returns:**
- `<-chan error`: Receive-only channel delivering result

**Example:**
```go
errCh := cache.SetAsync(ctx, "key", "text", "value")
// Do other work...
if err := <-errCh; err != nil {
    log.Printf("SetAsync failed: %v", err)
}
```

### GetAsync

Asynchronously retrieves a value by key.

```go
func (sc *SemanticCache[K, V]) GetAsync(
    ctx context.Context,
    key K,
) <-chan GetResult[V]
```

**Parameters:**
- Same as `Get`

**Returns:**
- `<-chan GetResult[V]`: Channel delivering result

**Example:**
```go
resultCh := cache.GetAsync(ctx, "key")
result := <-resultCh
if result.Error != nil {
    log.Printf("Error: %v", result.Error)
} else if result.Found {
    fmt.Printf("Value: %v\n", result.Value)
}
```

### DeleteAsync

Asynchronously deletes an entry.

```go
func (sc *SemanticCache[K, V]) DeleteAsync(
    ctx context.Context,
    key K,
) <-chan error
```

**Parameters:**
- Same as `Delete`

**Returns:**
- `<-chan error`: Channel delivering result

**Example:**
```go
errCh := cache.DeleteAsync(ctx, "key")
err := <-errCh
```

### LookupAsync

Asynchronously finds semantically similar content.

```go
func (sc *SemanticCache[K, V]) LookupAsync(
    ctx context.Context,
    inputText string,
    threshold float32,
) <-chan LookupResult[V]
```

**Parameters:**
- Same as `Lookup`

**Returns:**
- `<-chan LookupResult[V]`: Channel delivering result

**Example:**
```go
resultCh := cache.LookupAsync(ctx, "query", 0.8)
result := <-resultCh
if result.Error == nil && result.Match != nil {
    fmt.Printf("Match: %v\n", result.Match.Value)
}
```

### TopMatchesAsync

Asynchronously returns top N matches.

```go
func (sc *SemanticCache[K, V]) TopMatchesAsync(
    ctx context.Context,
    inputText string,
    n int,
) <-chan TopMatchesResult[V]
```

**Parameters:**
- Same as `TopMatches`

**Returns:**
- `<-chan TopMatchesResult[V]`: Channel delivering results

**Example:**
```go
resultCh := cache.TopMatchesAsync(ctx, "query", 5)
result := <-resultCh
if result.Error == nil {
    for _, match := range result.Matches {
        fmt.Printf("Score: %.2f\n", match.Score)
    }
}
```

---

## Batch Operations

### SetBatch

Stores multiple entries efficiently.

```go
func (sc *SemanticCache[K, V]) SetBatch(
    ctx context.Context,
    items []BatchItem[K, V],
) error
```

**Parameters:**
- `ctx`: Context
- `items`: Slice of items to store

**Returns:**
- `error`: First error encountered (fail-fast)

**Example:**
```go
items := []semanticcache.BatchItem[string, string]{
    {Key: "q1", InputText: "What is Go?", Value: "Programming language"},
    {Key: "q2", InputText: "What is Python?", Value: "Scripting language"},
}
err := cache.SetBatch(ctx, items)
```

**Notes:**
- Processes items sequentially
- Stops on first error
- All items validated before processing

### SetBatchAsync

Asynchronously stores multiple entries with concurrent embedding generation.

```go
func (sc *SemanticCache[K, V]) SetBatchAsync(
    ctx context.Context,
    items []BatchItem[K, V],
) <-chan error
```

**Parameters:**
- Same as `SetBatch`

**Returns:**
- `<-chan error`: Channel delivering result

**Example:**
```go
errCh := cache.SetBatchAsync(ctx, items)
if err := <-errCh; err != nil {
    log.Printf("Batch failed: %v", err)
}
```

**Notes:**
- Spawns goroutine per item for parallel embedding
- Fails fast on first error
- More efficient than sequential `SetBatch`

### GetBatch

Retrieves multiple values by keys.

```go
func (sc *SemanticCache[K, V]) GetBatch(
    ctx context.Context,
    keys []K,
) (map[K]V, error)
```

**Parameters:**
- `ctx`: Context
- `keys`: Keys to retrieve

**Returns:**
- `map[K]V`: Map of found key-value pairs
- `error`: Retrieval error

**Example:**
```go
values, err := cache.GetBatch(ctx, []string{"q1", "q2", "q3"})
for key, value := range values {
    fmt.Printf("%s: %v\n", key, value)
}
```

**Notes:**
- Only returns found keys (missing keys not in map)
- Returns error only on operation failure, not missing keys

### GetBatchAsync

Asynchronously retrieves multiple values using backend async capabilities.

```go
func (sc *SemanticCache[K, V]) GetBatchAsync(
    ctx context.Context,
    keys []K,
) <-chan GetBatchResult[K, V]
```

**Parameters:**
- Same as `GetBatch`

**Returns:**
- `<-chan GetBatchResult[K, V]`: Channel delivering results

**Example:**
```go
resultCh := cache.GetBatchAsync(ctx, []string{"key1", "key2"})
result := <-resultCh
if result.Error == nil {
    for k, v := range result.Values {
        fmt.Printf("%s: %v\n", k, v)
    }
}
```

**Performance:**
- Redis backend: Uses pipelining for single network round-trip
- In-memory: Concurrent goroutine retrieval

### DeleteBatch

Deletes multiple entries.

```go
func (sc *SemanticCache[K, V]) DeleteBatch(
    ctx context.Context,
    keys []K,
) error
```

**Parameters:**
- `ctx`: Context
- `keys`: Keys to delete

**Returns:**
- `error`: First deletion error

**Example:**
```go
err := cache.DeleteBatch(ctx, []string{"key1", "key2", "key3"})
```

### DeleteBatchAsync

Asynchronously deletes multiple entries concurrently.

```go
func (sc *SemanticCache[K, V]) DeleteBatchAsync(
    ctx context.Context,
    keys []K,
) <-chan error
```

**Parameters:**
- Same as `DeleteBatch`

**Returns:**
- `<-chan error`: Channel delivering result

**Example:**
```go
errCh := cache.DeleteBatchAsync(ctx, keys)
err := <-errCh
```

**Notes:**
- Spawns goroutine per key for parallel deletion
- Fails fast on first error

---

## Configuration Options

All options are in the `options` package.

### Backend Options

#### WithLRUBackend

Creates an LRU (Least Recently Used) in-memory backend.

```go
func WithLRUBackend[K comparable, V any](capacity int) Option[K, V]
```

**Parameters:**
- `capacity`: Maximum number of entries

**Example:**
```go
options.WithLRUBackend[string, string](1000)
```

#### WithLFUBackend

Creates an LFU (Least Frequently Used) in-memory backend.

```go
func WithLFUBackend[K comparable, V any](capacity int) Option[K, V]
```

**Parameters:**
- `capacity`: Maximum number of entries

**Example:**
```go
options.WithLFUBackend[string, string](500)
```

#### WithFIFOBackend

Creates a FIFO (First In, First Out) in-memory backend.

```go
func WithFIFOBackend[K comparable, V any](capacity int) Option[K, V]
```

**Parameters:**
- `capacity`: Maximum number of entries

**Example:**
```go
options.WithFIFOBackend[string, string](100)
```

#### WithRedisBackend

Creates a Redis backend.

```go
func WithRedisBackend[K comparable, V any](addr string, db int) Option[K, V]
```

**Parameters:**
- `addr`: Redis server address (`host:port`)
- `db`: Database number

**Example:**
```go
options.WithRedisBackend[string, string]("localhost:6379", 0)
```

**Notes:**
- Requires Redis with JSON support (RedisJSON module)
- Supports vector search (FT.SEARCH)

#### WithCustomBackend

Uses a custom backend implementation.

```go
func WithCustomBackend[K comparable, V any](
    backend types.CacheBackend[K, V],
) Option[K, V]
```

**Parameters:**
- `backend`: Custom backend implementing `CacheBackend` interface

**Example:**
```go
options.WithCustomBackend[string, string](myBackend)
```

### Provider Options

#### WithOpenAIProvider

Creates an OpenAI embedding provider.

```go
func WithOpenAIProvider[K comparable, V any](
    apiKey string,
    model ...string,
) Option[K, V]
```

**Parameters:**
- `apiKey`: OpenAI API key (or empty to use `OPENAI_API_KEY` env var)
- `model`: Optional model name (default: `text-embedding-3-small`)

**Example:**
```go
// Default model
options.WithOpenAIProvider[string, string]("sk-...")

// Specific model
options.WithOpenAIProvider[string, string]("sk-...", "text-embedding-3-large")

// From environment
options.WithOpenAIProvider[string, string]("")
```

**Available Models:**
- `text-embedding-3-small` (1536 dims, default)
- `text-embedding-3-large` (3072 dims)
- `text-embedding-ada-002` (1536 dims, legacy)

#### WithCustomProvider

Uses a custom embedding provider.

```go
func WithCustomProvider[K comparable, V any](
    provider types.EmbeddingProvider,
) Option[K, V]
```

**Parameters:**
- `provider`: Custom provider implementing `EmbeddingProvider` interface

**Example:**
```go
options.WithCustomProvider[string, string](myProvider)
```

### Similarity Options

#### WithSimilarityComparator

Sets the similarity function for semantic search.

```go
func WithSimilarityComparator[K comparable, V any](
    comparator similarity.SimilarityFunc,
) Option[K, V]
```

**Parameters:**
- `comparator`: Similarity function

**Example:**
```go
options.WithSimilarityComparator[string, string](similarity.CosineSimilarity)
```

**Built-in Functions:**
- `similarity.CosineSimilarity` (default) - Range: [-1, 1]
- `similarity.EuclideanSimilarity` - Range: [0, 1]
- `similarity.DotProductSimilarity` - Range: unbounded
- `similarity.ManhattanSimilarity` - Range: [0, 1]
- `similarity.PearsonCorrelationSimilarity` - Range: [-1, 1]

---

## Interfaces

### CacheBackend

Backend storage interface.

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

    // Async methods
    SetAsync(ctx context.Context, key K, entry Entry[V]) <-chan error
    GetAsync(ctx context.Context, key K) <-chan AsyncGetResult[V]
    DeleteAsync(ctx context.Context, key K) <-chan error
    GetBatchAsync(ctx context.Context, keys []K) <-chan AsyncBatchResult[K, V]
}
```

### EmbeddingProvider

Embedding generation interface.

```go
type EmbeddingProvider interface {
    EmbedText(text string) ([]float32, error)
    Close()
}
```

### SimilarityFunc

Similarity calculation function type.

```go
type SimilarityFunc func(a, b []float32) float32
```

**Parameters:**
- `a`, `b`: Embedding vectors

**Returns:**
- `float32`: Similarity score (interpretation depends on function)

**Notes:**
- Should handle empty vectors gracefully (return 0)
- Should handle mismatched lengths (return 0)
- Higher scores = more similar (except inverted distances)

---

## Error Handling

### Common Errors

**Configuration Errors:**
```go
cache, err := semanticcache.New[string, string](
    // Missing required options
)
// err: "backend is required - use WithLRUBackend, etc"
// err: "embedding provider is required - use WithOpenAIProvider, etc"
```

**Operation Errors:**
```go
err := cache.Set(ctx, "", "text", "value")
// err: "key cannot be zero value"

matches, err := cache.TopMatches(ctx, "query", -1)
// err: "n must be positive"
```

**Provider Errors:**
```go
// OpenAI API errors
// err: "failed to generate embedding: <OpenAI error>"
```

**Backend Errors:**
```go
// Redis connection errors
// err: "failed to connect to Redis: <connection error>"
```

### Error Checking Pattern

```go
// Construction
cache, err := semanticcache.New[string, string](opts...)
if err != nil {
    log.Fatalf("Failed to create cache: %v", err)
}
defer cache.Close()

// Operations
if err := cache.Set(ctx, key, text, value); err != nil {
    log.Printf("Set failed: %v", err)
    return err
}

// Semantic search
match, err := cache.Lookup(ctx, query, 0.8)
if err != nil {
    log.Printf("Lookup failed: %v", err)
    return err
}
if match == nil {
    log.Println("No match found")
    return
}
fmt.Printf("Found: %v (score: %.2f)\n", match.Value, match.Score)
```

---

## Performance Considerations

### Operation Complexity

**In-Memory Backends:**
- `Set`: O(1)
- `Get`: O(1)
- `Delete`: O(1) for LRU/LFU, O(n) for FIFO
- `Lookup`: O(n) - iterates all entries
- `TopMatches`: O(n log n) - sorts all similarities

**Redis Backend:**
- `Set`: O(1) + network
- `Get`: O(1) + network
- `Lookup`: O(n) - could use FT.SEARCH KNN for O(log n)
- `GetBatchAsync`: O(n) with pipelining (single round-trip)

### Best Practices

1. **Choose Right Backend:**
   - LRU: Time-based locality
   - LFU: Frequency-based access
   - FIFO: Simple caching
   - Redis: Shared cache, persistence

2. **Use Async for Concurrency:**
   - Non-blocking operations
   - Parallel processing
   - Better throughput

3. **Batch Operations:**
   - Reduce API calls (embeddings)
   - Network efficiency (Redis)
   - Better performance

4. **Context Timeouts:**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   ```

5. **Similarity Thresholds:**
   - 0.9+: High precision
   - 0.8-0.9: Balanced (recommended)
   - 0.7-0.8: High recall
