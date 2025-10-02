# Performance Guide & Best Practices

Comprehensive guide to optimizing SemanticCache performance and following best practices.

## Table of Contents

- [Performance Characteristics](#performance-characteristics)
- [Backend Selection](#backend-selection)
- [Optimization Strategies](#optimization-strategies)
- [Async vs Sync Operations](#async-vs-sync-operations)
- [Best Practices](#best-practices)
- [Common Pitfalls](#common-pitfalls)
- [Benchmarking](#benchmarking)

---

## Performance Characteristics

### Operation Complexity

#### In-Memory Backends (LRU, LFU, FIFO)

| Operation | Time Complexity | Space Complexity | Notes |
|-----------|----------------|------------------|-------|
| `Set` | O(1) | O(1) | Hash map insert |
| `Get` | O(1) | O(1) | Hash map lookup |
| `Delete` | O(1) - O(n) | O(1) | O(n) for FIFO queue removal |
| `Contains` | O(1) | O(1) | Hash map check |
| `Lookup` | O(n) | O(1) | Linear scan of all entries |
| `TopMatches` | O(n log n) | O(n) | Compute all scores, then sort |
| `GetEmbedding` | O(1) | O(1) | Separate index lookup |
| `Keys` | O(n) | O(n) | Return all keys |
| `Flush` | O(n) | O(1) | Clear all entries |

**Key Insight:** Semantic search (`Lookup`, `TopMatches`) is O(n) - scales linearly with cache size.

#### Redis Backend

| Operation | Time Complexity | Notes |
|-----------|----------------|-------|
| `Set` | O(1) + network | JSON.SET |
| `Get` | O(1) + network | JSON.GET |
| `Delete` | O(1) + network | DEL command |
| `Lookup` | O(n) + network | **Could be O(log n) with FT.SEARCH** |
| `GetBatchAsync` | O(n) + 1 RTT | Pipelining (single round-trip) |
| `Flush` | O(n) + network | SCAN + DEL |

**Key Insight:** Network latency dominates. Use batch operations to minimize round-trips.

### Memory Usage

**Per Entry Overhead:**

```
In-Memory (LRU):
  Entry struct:        16 bytes (pointer to embedding, value)
  Embedding:           dimensions × 4 bytes (float32)
  Value:               size of V
  Map overhead:        ~48 bytes per entry (Go map internals)
  Index map:           8 bytes (pointer) per key
  Total:               ~72 bytes + embedding_size + value_size

Example (text-embedding-3-small, string value):
  Embedding:           1536 × 4 = 6,144 bytes
  Value (string):      ~64 bytes (average)
  Overhead:            ~72 bytes
  Total per entry:     ~6,280 bytes

  10,000 entries:      ~60 MB
  100,000 entries:     ~600 MB
  1,000,000 entries:   ~6 GB
```

**Redis:**
```
Per entry:
  JSON document:       ~embedding_size + value_size + metadata
  Vector index:        Additional ~10-20% overhead (HNSW)

Network:
  Bandwidth:           ~6 KB per Set operation (1536-dim embedding)
```

---

## Backend Selection

### When to Use LRU (Least Recently Used)

**Best For:**
- Time-based access patterns
- Recent queries more important
- Working set fits in memory
- Single-application caching

**Example Use Cases:**
- User session data
- Recent search queries
- Active conversation history

**Configuration:**
```go
cache, err := semanticcache.New[string, Response](
    options.WithLRUBackend(10000),  // Keep 10k most recent
    options.WithOpenAIProvider("key"),
)
```

**Performance:**
- Set: O(1)
- Get: O(1), updates recency
- Best when: Access patterns show temporal locality

### When to Use LFU (Least Frequently Used)

**Best For:**
- Frequency-based access patterns
- Popular items stay cached
- Long-running applications
- Skewed access distribution

**Example Use Cases:**
- FAQ systems (popular questions)
- Product catalog (hot items)
- API response caching (common endpoints)

**Configuration:**
```go
cache, err := semanticcache.New[string, string](
    options.WithLFUBackend(5000),  // Keep 5k most frequent
    options.WithOpenAIProvider("key"),
)
```

**Performance:**
- Set: O(1), tracks frequency
- Get: O(1), increments frequency
- Best when: 80/20 rule applies (20% of items = 80% of accesses)

### When to Use FIFO (First In, First Out)

**Best For:**
- Simple eviction policy
- Data has inherent ordering
- Predictable access patterns
- Low-overhead caching

**Example Use Cases:**
- Event stream caching
- Log data
- Time-series data

**Configuration:**
```go
cache, err := semanticcache.New[string, LogEntry](
    options.WithFIFOBackend(1000),
    options.WithOpenAIProvider("key"),
)
```

**Performance:**
- Set: O(1)
- Get: O(1)
- Delete: O(n) - queue removal
- Best when: Simple eviction is acceptable

### When to Use Redis

**Best For:**
- Shared cache across multiple instances
- Persistence required
- Large datasets (> memory)
- Distributed systems

**Example Use Cases:**
- Multi-server deployments
- Microservices sharing cache
- Production systems requiring durability
- Caches > 100GB

**Configuration:**
```go
cache, err := semanticcache.New[string, MyStruct](
    options.WithRedisBackend("localhost:6379", 0),
    options.WithOpenAIProvider("key"),
)
```

**Performance:**
- Network latency: 1-5ms (local), 10-50ms (cloud)
- Throughput: Limited by network, not computation
- Best when: Sharing cache is more important than latency

**Optimization:**
- Use `GetBatchAsync` for parallel retrieval
- Enable Redis pipelining (automatic)
- Consider Redis Cluster for scale

---

## Optimization Strategies

### 1. Choose Right Similarity Function

**Performance Comparison (1536-dim vectors):**

| Function | Complexity | Speed | Use Case |
|----------|-----------|-------|----------|
| Cosine | O(n) | ~1 μs | General purpose (default) |
| Euclidean | O(n) | ~0.8 μs | Absolute position matters |
| Dot Product | O(n) | ~0.5 μs | Pre-normalized vectors |
| Manhattan | O(n) | ~0.7 μs | High dimensions, outliers |
| Pearson | O(n) | ~2 μs | Pattern matching |

**Recommendation:**
- **Default:** Use Cosine (good balance)
- **Speed:** Use Dot Product if vectors pre-normalized
- **Accuracy:** Use Pearson for correlation-based matching

```go
// Fastest (if embeddings are normalized)
options.WithSimilarityComparator(similarity.DotProductSimilarity)

// Most accurate (general case)
options.WithSimilarityComparator(similarity.CosineSimilarity)

// Pattern matching
options.WithSimilarityComparator(similarity.PearsonCorrelationSimilarity)
```

### 2. Optimize Threshold Selection

**Threshold impacts both performance and accuracy:**

```go
// High threshold = fewer comparisons (early exit possible)
match, _ := cache.Lookup(ctx, query, 0.95)  // Very strict

// Medium threshold = balanced
match, _ := cache.Lookup(ctx, query, 0.85)  // Recommended

// Low threshold = more matches (slower)
match, _ := cache.Lookup(ctx, query, 0.70)  // Permissive
```

**Best Practice:**
1. Start with 0.85 threshold
2. Measure false positive/negative rates
3. Adjust based on use case:
   - High precision needed? Increase to 0.9+
   - High recall needed? Decrease to 0.7-0.8

### 3. Use Batch Operations

**Instead of:**
```go
// Slow: N API calls, N cache operations
for _, item := range items {
    cache.Set(ctx, item.Key, item.Text, item.Value)
}
```

**Do this:**
```go
// Fast: Single batch operation
batchItems := make([]semanticcache.BatchItem[string, Data], len(items))
for i, item := range items {
    batchItems[i] = semanticcache.BatchItem[string, Data]{
        Key: item.Key,
        InputText: item.Text,
        Value: item.Value,
    }
}
cache.SetBatch(ctx, batchItems)
```

**Performance Gain:**
- In-memory: ~2-3x faster (less overhead)
- Redis: ~10x faster (pipelining)
- OpenAI API: Could batch embeddings (future optimization)

### 4. Leverage Async Operations

**For Independent Operations:**
```go
// Sequential (slow): 3× latency
cache.Set(ctx, "key1", "text1", "value1")
cache.Set(ctx, "key2", "text2", "value2")
cache.Set(ctx, "key3", "text3", "value3")

// Concurrent (fast): max(latencies)
errCh1 := cache.SetAsync(ctx, "key1", "text1", "value1")
errCh2 := cache.SetAsync(ctx, "key2", "text2", "value2")
errCh3 := cache.SetAsync(ctx, "key3", "text3", "value3")

// Wait for all
for _, ch := range []<-chan error{errCh1, errCh2, errCh3} {
    if err := <-ch; err != nil {
        log.Printf("Error: %v", err)
    }
}
```

**When to Use Async:**
- ✅ Multiple independent operations
- ✅ Don't need immediate result
- ✅ Can parallelize work
- ❌ Operations depend on each other
- ❌ Need synchronous flow

### 5. Optimize Embedding API Calls

**OpenAI API Limits:**
- Rate limits: Tier-based (500-10,000 RPM)
- Token limits: 8,191 tokens per request
- Cost: Per token charged

**Strategies:**

**a) Cache Query Embeddings:**
```go
type CachedEmbeddingProvider struct {
    provider types.EmbeddingProvider
    cache    map[string][]float32  // Query → embedding
    mu       sync.RWMutex
}

func (p *CachedEmbeddingProvider) EmbedText(text string) ([]float32, error) {
    p.mu.RLock()
    if emb, ok := p.cache[text]; ok {
        p.mu.RUnlock()
        return emb, nil
    }
    p.mu.RUnlock()

    emb, err := p.provider.EmbedText(text)
    if err != nil {
        return nil, err
    }

    p.mu.Lock()
    p.cache[text] = emb
    p.mu.Unlock()

    return emb, nil
}
```

**b) Batch Embeddings (Future):**
```go
// TODO: Library could detect batch operations and batch API calls
embeddings, err := provider.EmbedTexts([]string{"text1", "text2", ...})
```

**c) Use Smaller Models:**
```go
// Faster, cheaper, slightly less accurate
options.WithOpenAIProvider("key", "text-embedding-3-small")  // 1536 dims

// vs. larger model
options.WithOpenAIProvider("key", "text-embedding-3-large")  // 3072 dims
```

### 6. Limit Cache Size

**Problem:** Large caches slow down semantic search (O(n))

**Solution:** Set appropriate capacity

```go
// Bad: Unlimited growth
cache := New(options.WithLRUBackend(1000000))  // 1M entries = slow search

// Good: Bounded by use case
cache := New(options.WithLRUBackend(10000))  // 10k entries = fast search
```

**Rule of Thumb:**
- < 1,000 entries: Any backend works well
- 1,000 - 10,000: In-memory backends acceptable
- 10,000 - 100,000: Consider optimization or Redis
- > 100,000: Need vector indexing (future feature)

### 7. Use Context Timeouts

**Prevent Hanging Operations:**
```go
// Bad: No timeout
ctx := context.Background()
match, err := cache.Lookup(ctx, query, 0.8)  // Could hang forever

// Good: With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

match, err := cache.Lookup(ctx, query, 0.8)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Lookup timed out")
    }
}
```

**Recommended Timeouts:**
- `Set`: 10s (includes embedding API call)
- `Get`: 1s (cache lookup only)
- `Lookup`: 30s (depends on cache size)
- `TopMatches`: 60s (compute-heavy)

---

## Async vs Sync Operations

### When to Use Sync

**Use synchronous operations when:**

1. **Simple request-response flow:**
   ```go
   func handleRequest(w http.ResponseWriter, r *http.Request) {
       value, found, err := cache.Get(r.Context(), key)
       if found {
           json.NewEncoder(w).Encode(value)
       }
   }
   ```

2. **Operation must complete before proceeding:**
   ```go
   err := cache.Set(ctx, key, text, value)
   if err != nil {
       return err  // Can't continue
   }
   // Use value immediately
   processValue(value)
   ```

3. **Sequential dependencies:**
   ```go
   // Must happen in order
   cache.Set(ctx, "user", userQuery, userData)
   match, _ := cache.Lookup(ctx, "similar query", 0.8)
   ```

### When to Use Async

**Use asynchronous operations when:**

1. **Fire-and-forget:**
   ```go
   func handleUpdate(update Update) {
       // Don't block on cache update
       errCh := cache.SetAsync(ctx, update.Key, update.Text, update.Value)

       // Do other work
       processUpdate(update)

       // Check result later (optional)
       if err := <-errCh; err != nil {
           log.Printf("Cache update failed: %v", err)
       }
   }
   ```

2. **Parallel operations:**
   ```go
   // Search multiple queries concurrently
   results := make([]Match, len(queries))
   channels := make([]<-chan LookupResult, len(queries))

   for i, query := range queries {
       channels[i] = cache.LookupAsync(ctx, query, 0.8)
   }

   for i, ch := range channels {
       result := <-ch
       if result.Match != nil {
           results[i] = *result.Match
       }
   }
   ```

3. **Non-blocking background tasks:**
   ```go
   func warmCache(items []Item) {
       for _, item := range items {
           // Non-blocking warmup
           cache.SetAsync(context.Background(), item.Key, item.Text, item.Value)
       }
       // Continue immediately
   }
   ```

### Performance Comparison

**Scenario: Set 100 items**

| Approach | Latency | Throughput | Complexity |
|----------|---------|------------|------------|
| Sync Sequential | 100 × T | 1/T | Simple |
| Sync Batch | 1 × T_batch | Higher | Simple |
| Async Concurrent | max(T) | 100/T | Moderate |
| Async Batch | 1 × T_batch | Highest | Simple |

Where T = time per operation

**Recommendation:**
- **Best:** `SetBatchAsync` (combines batching + async)
- **Good:** `SetBatch` (batching alone)
- **Okay:** Individual `SetAsync` (parallel)
- **Slow:** Individual `Set` (sequential)

---

## Best Practices

### 1. Resource Management

**Always close cache when done:**
```go
func main() {
    cache, err := semanticcache.New[string, string](...)
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()  // ← Important!

    // Use cache...
}
```

**Why:** Closes embedding provider, releases backend resources (connections, goroutines)

### 2. Error Handling

**Check all errors:**
```go
// Bad
cache.Set(ctx, key, text, value)  // Ignoring error

// Good
if err := cache.Set(ctx, key, text, value); err != nil {
    log.Printf("Failed to cache: %v", err)
    // Handle appropriately (retry, fallback, etc.)
}
```

**Semantic search errors:**
```go
match, err := cache.Lookup(ctx, query, 0.8)
if err != nil {
    // Embedding or search error
    return err
}
if match == nil {
    // No match found (not an error)
    log.Println("No similar content found")
}
```

### 3. Context Usage

**Pass request context through:**
```go
func handleAPI(w http.ResponseWriter, r *http.Request) {
    // Use request context (includes cancellation)
    match, err := cache.Lookup(r.Context(), query, 0.8)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            // Client disconnected
            return
        }
        http.Error(w, err.Error(), 500)
        return
    }
    // ...
}
```

### 4. Key Design

**Use structured keys:**
```go
// Bad: Hard to manage
cache.Set(ctx, "user123query", text, value)

// Good: Structured
type CacheKey struct {
    UserID string
    QueryID string
}

func (k CacheKey) String() string {
    return fmt.Sprintf("%s:%s", k.UserID, k.QueryID)
}

cache.Set(ctx, CacheKey{"user123", "query456"}, text, value)
```

**Key namespacing:**
```go
// Prevent key collisions across use cases
cache.Set(ctx, "faq:"+questionID, question, answer)
cache.Set(ctx, "docs:"+docID, docText, docContent)
```

### 5. Monitoring

**Track cache performance:**
```go
type CacheMetrics struct {
    Hits   int64
    Misses int64
    Errors int64
}

var metrics CacheMetrics

func cacheGet(ctx context.Context, key string) (string, bool) {
    value, found, err := cache.Get(ctx, key)
    if err != nil {
        atomic.AddInt64(&metrics.Errors, 1)
        return "", false
    }
    if found {
        atomic.AddInt64(&metrics.Hits, 1)
    } else {
        atomic.AddInt64(&metrics.Misses, 1)
    }
    return value, found
}

// Expose metrics
func (m *CacheMetrics) HitRatio() float64 {
    total := m.Hits + m.Misses
    if total == 0 {
        return 0
    }
    return float64(m.Hits) / float64(total)
}
```

---

## Common Pitfalls

### 1. Ignoring Zero-Value Keys

**Problem:**
```go
var key string  // Zero value: ""
cache.Set(ctx, key, text, value)  // Error: key cannot be zero value
```

**Solution:**
```go
if key == "" {
    return errors.New("invalid key")
}
cache.Set(ctx, key, text, value)
```

### 2. Not Checking for Nil Matches

**Problem:**
```go
match, _ := cache.Lookup(ctx, query, 0.8)
fmt.Println(match.Value)  // Panic if match is nil!
```

**Solution:**
```go
match, err := cache.Lookup(ctx, query, 0.8)
if err != nil {
    return err
}
if match != nil {
    fmt.Println(match.Value)
} else {
    fmt.Println("No match found")
}
```

### 3. Goroutine Leaks with Async

**Problem:**
```go
for i := 0; i < 1000000; i++ {
    cache.SetAsync(ctx, key, text, value)  // Channel never read!
}
// Goroutines accumulate
```

**Solution:**
```go
for i := 0; i < 1000; i++ {
    errCh := cache.SetAsync(ctx, key, text, value)
    // Must read from channel
    if err := <-errCh; err != nil {
        log.Printf("Error: %v", err)
    }
}
```

**Or use sync version:**
```go
for i := 0; i < 1000000; i++ {
    cache.Set(ctx, key, text, value)  // No goroutines
}
```

### 4. Wrong Similarity Threshold

**Problem:**
```go
// Looking for exact matches but using low threshold
match, _ := cache.Lookup(ctx, "password reset", 0.5)  // Too permissive
```

**Solution:**
```go
// Adjust threshold to use case
match, _ := cache.Lookup(ctx, "password reset", 0.9)  // Strict
```

**Guidelines:**
- Exact match needed: 0.95+
- Close match: 0.85-0.95
- Similar topic: 0.75-0.85
- Related content: 0.65-0.75

### 5. Large Batch Operations

**Problem:**
```go
// Memory spike: All items in memory
items := make([]BatchItem, 1000000)  // 1M items!
cache.SetBatch(ctx, items)
```

**Solution:**
```go
// Chunk into smaller batches
const batchSize = 1000
for i := 0; i < len(items); i += batchSize {
    end := i + batchSize
    if end > len(items) {
        end = len(items)
    }
    cache.SetBatch(ctx, items[i:end])
}
```

---

## Benchmarking

### Running Benchmarks

**Create benchmark file:**
```go
// cache_bench_test.go
package semanticcache_test

import (
    "context"
    "testing"
    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
)

func BenchmarkSet(b *testing.B) {
    cache, _ := semanticcache.New[string, string](
        options.WithLRUBackend(10000),
        options.WithCustomProvider(&mockProvider{}),
    )
    defer cache.Close()

    ctx := context.Background()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        cache.Set(ctx, "key", "text", "value")
    }
}

func BenchmarkLookup(b *testing.B) {
    cache, _ := setupCache(1000)  // 1k entries
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Lookup(ctx, "query", 0.8)
    }
}
```

**Run benchmarks:**
```bash
# All benchmarks
go test -bench=. -benchmem

# Specific benchmark
go test -bench=BenchmarkLookup -benchtime=10s

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Interpreting Results

**Example output:**
```
BenchmarkSet-8              50000    25000 ns/op    1024 B/op    15 allocs/op
BenchmarkGet-8            1000000     1200 ns/op      64 B/op     2 allocs/op
BenchmarkLookup-8            1000  1200000 ns/op    4096 B/op   100 allocs/op
```

**Columns:**
- `50000`: Number of iterations
- `25000 ns/op`: 25 μs per operation
- `1024 B/op`: Bytes allocated per op
- `15 allocs/op`: Allocations per op

**Red Flags:**
- Lookup time grows linearly with cache size (expected, but note threshold)
- High allocation rate (potential GC pressure)
- Operations slower than network latency (for Redis)

### Performance Goals

**Target Performance (Approximate):**

| Operation | In-Memory | Redis | Notes |
|-----------|-----------|-------|-------|
| Set | < 50 μs | < 5 ms | Includes embedding |
| Get | < 1 μs | < 2 ms | Cache hit |
| Lookup (1k entries) | < 1 ms | < 10 ms | Linear scan |
| Lookup (10k entries) | < 10 ms | < 50 ms | Needs indexing |
| TopMatches (1k, n=10) | < 2 ms | < 15 ms | Sort overhead |

**Note:** Actual performance depends on:
- Hardware (CPU, network)
- Embedding dimensions
- Similarity function
- Value sizes
