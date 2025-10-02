# Architecture & Design

This document describes the architecture, design patterns, and technical decisions behind SemanticCache.

## Table of Contents

- [Overview](#overview)
- [Architecture Layers](#architecture-layers)
- [Design Patterns](#design-patterns)
- [Component Details](#component-details)
- [Data Flow](#data-flow)
- [Concurrency Model](#concurrency-model)
- [Extension Points](#extension-points)

---

## Overview

SemanticCache is a layered architecture for semantic caching with vector embeddings. The design prioritizes:

1. **Modularity**: Pluggable components (backends, providers, similarity)
2. **Type Safety**: Full generic support with compile-time checking
3. **Performance**: Async operations, batch processing, optimized algorithms
4. **Extensibility**: Well-defined interfaces for custom implementations
5. **Developer Experience**: Fluent API, functional options, clear errors

### High-Level Architecture

```
┌──────────────────────────────────────────────────────┐
│              Application Code                         │
└────────────────────┬─────────────────────────────────┘
                     │
                     ↓
┌──────────────────────────────────────────────────────┐
│         SemanticCache[K, V] (Main API)               │
│  • Set/Get/Delete                                     │
│  • Lookup/TopMatches (Semantic Search)                │
│  • Batch Operations                                   │
│  • Async Variants                                     │
└──┬─────────────┬──────────────┬──────────────────────┘
   │             │              │
   ↓             ↓              ↓
┌────────┐  ┌─────────────┐  ┌──────────────┐
│Backend │  │  Provider   │  │ Similarity   │
│Storage │  │ (Embeddings)│  │  Functions   │
└────────┘  └─────────────┘  └──────────────┘
   │             │              │
   ↓             ↓              ↓
┌────────┐  ┌─────────────┐  ┌──────────────┐
│• LRU   │  │• OpenAI     │  │• Cosine      │
│• LFU   │  │• Custom     │  │• Euclidean   │
│• FIFO  │  │             │  │• Dot Product │
│• Redis │  │             │  │• Manhattan   │
│• Custom│  │             │  │• Pearson     │
└────────┘  └─────────────┘  └──────────────┘
```

---

## Architecture Layers

### Layer 1: Public API

**Location:** `cache.go`
**Responsibility:** High-level semantic caching operations

**Key Types:**
- `SemanticCache[K, V]` - Main cache type
- `Match[V]` - Search result
- `BatchItem[K, V]` - Batch operation item
- Result types for async operations

**Operations:**
- CRUD: `Set`, `Get`, `Delete`, `Contains`
- Semantic Search: `Lookup`, `TopMatches`
- Batch: `SetBatch`, `GetBatch`, `DeleteBatch`
- Async variants of all operations
- Management: `Flush`, `Len`, `Close`

### Layer 2: Configuration

**Location:** `options/`
**Responsibility:** Fluent configuration API

**Pattern:** Functional Options
```go
type Option[K comparable, V any] func(*Config[K, V]) error

type Config[K comparable, V any] struct {
    Backend    types.CacheBackend[K, V]
    Provider   types.EmbeddingProvider
    Comparator similarity.SimilarityFunc
}
```

**Benefits:**
- Self-documenting API
- Type-safe configuration
- Easy to extend without breaking changes
- Validation at construction time

### Layer 3: Abstractions

**Location:** `types/`
**Responsibility:** Interfaces and shared types

**Core Interfaces:**

```go
// Storage abstraction
type CacheBackend[K comparable, V any] interface {
    // Sync operations
    Set(ctx, key, entry) error
    Get(ctx, key) (Entry[V], bool, error)
    Delete(ctx, key) error
    // ... more methods

    // Async operations
    SetAsync(ctx, key, entry) <-chan error
    GetAsync(ctx, key) <-chan AsyncGetResult[V]
    // ... more async methods
}

// Embedding generation abstraction
type EmbeddingProvider interface {
    EmbedText(text string) ([]float32, error)
    Close()
}
```

### Layer 4: Implementations

**Backends** (`backends/`)
- In-memory: LRU, LFU, FIFO
- Remote: Redis
- Custom: User-defined

**Providers** (`providers/`)
- OpenAI: Official SDK integration
- Custom: User-defined

**Similarity** (`similarity/`)
- Cosine, Euclidean, Dot Product, Manhattan, Pearson
- Custom: User-defined functions

---

## Design Patterns

### 1. Functional Options Pattern

**Problem:** How to provide flexible, extensible configuration?

**Solution:** Options as functions that modify config

```go
// Option definition
type Option[K comparable, V any] func(*Config[K, V]) error

// Option constructors
func WithLRUBackend[K, V](cap int) Option[K, V] {
    return func(c *Config[K, V]) error {
        backend, err := inmemory.NewLRUBackend[K, V](
            types.BackendConfig{Capacity: cap},
        )
        c.Backend = backend
        return err
    }
}

// Usage
cache, err := New[string, string](
    WithLRUBackend(1000),
    WithOpenAIProvider("key"),
    WithSimilarityComparator(similarity.CosineSimilarity),
)
```

**Benefits:**
- Variable number of options
- Options can fail gracefully
- Easy to add new options
- Options compose naturally

### 2. Interface Segregation

**Problem:** How to make components swappable?

**Solution:** Small, focused interfaces

```go
// Not this (fat interface)
type Backend interface {
    Set(...)
    Get(...)
    Delete(...)
    SetLRU(...)     // LRU-specific
    SetTTL(...)     // Redis-specific
    // ... 20 methods
}

// But this (focused interface)
type CacheBackend[K, V] interface {
    Set(ctx, key, entry) error
    Get(ctx, key) (Entry[V], bool, error)
    // ... essential methods only
}
```

**Benefits:**
- Easy to implement custom backends
- Clear separation of concerns
- Minimal required methods

### 3. Strategy Pattern

**Problem:** How to make similarity algorithms pluggable?

**Solution:** Function type as strategy

```go
type SimilarityFunc func(a, b []float32) float32

// Strategies
func CosineSimilarity(a, b []float32) float32 { /*...*/ }
func EuclideanSimilarity(a, b []float32) float32 { /*...*/ }

// Usage
cache := New(
    // ...
    WithSimilarityComparator(CosineSimilarity), // Strategy injection
)
```

**Benefits:**
- Strategies are stateless (pure functions)
- Easy to test individually
- Easy to add custom strategies
- No class hierarchy needed

### 4. Generics for Type Safety

**Problem:** How to support arbitrary key/value types?

**Solution:** Go 1.18+ generics

```go
type SemanticCache[K comparable, V any] struct {
    backend    types.CacheBackend[K, V]
    provider   types.EmbeddingProvider
    comparator similarity.SimilarityFunc
}

// Type-safe usage
cache1 := New[string, MyStruct](...)   // string keys, struct values
cache2 := New[int, string](...)        // int keys, string values
cache3 := New[UUID, []byte](...)       // UUID keys, byte slice values
```

**Benefits:**
- Compile-time type checking
- No type assertions needed
- Better IDE support
- No runtime overhead

### 5. Factory Pattern

**Problem:** How to create backends by type?

**Solution:** Factory with type-based dispatch

```go
type BackendFactory[K comparable, V any] struct{}

func (f *BackendFactory[K, V]) NewBackend(
    backendType types.BackendType,
    config types.BackendConfig,
) (types.CacheBackend[K, V], error) {
    switch backendType {
    case types.BackendLRU:
        return inmemory.NewLRUBackend[K, V](config)
    case types.BackendRedis:
        return remote.NewRedisBackend[K, V](config)
    // ...
    }
}
```

**Benefits:**
- Centralized creation logic
- Easy to add new backend types
- Type-safe construction

### 6. Result Type Pattern (Async)

**Problem:** How to return multiple values from channels?

**Solution:** Result structs encapsulating outcomes

```go
// Instead of multiple channels
func GetAsync(...) (<-chan V, <-chan bool, <-chan error) {
    // Complex, error-prone
}

// Use result type
type GetResult[V any] struct {
    Value V
    Found bool
    Error error
}

func GetAsync(...) <-chan GetResult[V] {
    resultCh := make(chan GetResult[V], 1)
    go func() {
        value, found, err := sc.Get(...)
        resultCh <- GetResult[V]{value, found, err}
        close(resultCh)
    }()
    return resultCh
}
```

**Benefits:**
- Single channel to manage
- Atomic result delivery
- Clear error handling

---

## Component Details

### SemanticCache Core

**Responsibilities:**
1. Coordinate between backend, provider, and similarity function
2. Implement semantic search logic
3. Provide sync and async APIs
4. Handle errors and validation

**Key Algorithm (Lookup):**
```go
func (sc *SemanticCache[K, V]) Lookup(
    ctx context.Context,
    inputText string,
    threshold float32,
) (*Match[V], error) {
    // 1. Generate query embedding
    queryEmbedding, err := sc.provider.EmbedText(inputText)

    // 2. Get all keys
    keys, err := sc.backend.Keys(ctx)

    // 3. Iterate and compute similarities
    var bestMatch *Match[V]
    var bestScore = threshold

    for _, key := range keys {
        // Get embedding for key
        embedding, found, err := sc.backend.GetEmbedding(ctx, key)

        // Compute similarity
        score := sc.comparator(queryEmbedding, embedding)

        // Track best match
        if score >= bestScore {
            entry, found, err := sc.backend.Get(ctx, key)
            if found {
                bestMatch = &Match[V]{
                    Value: entry.Value,
                    Score: score,
                }
                bestScore = score // Raise threshold
            }
        }
    }

    return bestMatch, nil
}
```

**Optimization Opportunity:** Use vector index instead of linear scan

### Backend: LRU

**Implementation:** `backends/inmemory/lru.go`

**Data Structures:**
```go
type LRUBackend[K, V] struct {
    mu    *sync.RWMutex                    // Thread safety
    cache *lru.Cache[K, types.Entry[V]]    // LRU cache (hashicorp)
    index map[K][]float32                  // Fast embedding lookup
}
```

**Key Design:**
- Uses `hashicorp/golang-lru/v2` for proven LRU implementation
- Separate `index` for O(1) embedding retrieval
- RWMutex for concurrent access (read-heavy workload)
- Index cleanup on `Keys()` to remove stale entries

**Thread Safety:**
```go
func (b *LRUBackend[K, V]) Set(...) {
    b.mu.Lock()         // Write lock
    defer b.mu.Unlock()

    b.cache.Add(key, entry)
    b.index[key] = entry.Embedding
}

func (b *LRUBackend[K, V]) Get(...) {
    b.mu.RLock()        // Read lock (allows concurrent reads)
    defer b.mu.RUnlock()

    entry, ok := b.cache.Get(key)
    return entry, ok, nil
}
```

### Backend: Redis

**Implementation:** `backends/remote/redis.go`

**Data Structures:**
```go
type RedisBackend[K, V] struct {
    client     *redis.Client  // go-redis client
    prefix     string          // Key namespace
    indexName  string          // FT.SEARCH index
    dimensions int             // Embedding size
}

// Storage format
type redisDocument[V any] struct {
    Key       string
    Value     V
    Embedding []float64  // Redis JSON uses float64
    Timestamp int64
}
```

**Key Features:**
1. **JSON Storage:** Uses RedisJSON module for structured data
2. **Vector Index:** Creates FT.SEARCH index with HNSW algorithm
3. **URL Parsing:** Supports redis://, rediss://, and simple formats
4. **Pipelining:** Batch operations use Redis pipelining

**Async Batch Optimization:**
```go
func (b *RedisBackend[K, V]) GetBatchAsync(ctx, keys) <-chan Result {
    resultCh := make(chan Result, 1)
    go func() {
        pipe := b.client.Pipeline()

        // Queue all commands
        cmds := make(map[K]*redis.JSONCmd)
        for _, key := range keys {
            cmds[key] = pipe.JSONGet(ctx, key, "$")
        }

        // Single network round-trip
        pipe.Exec(ctx)

        // Process results
        entries := make(map[K]Entry[V])
        for key, cmd := range cmds {
            result, err := cmd.Result()
            // Parse and add to entries
        }

        resultCh <- Result{entries, nil}
    }()
    return resultCh
}
```

### Provider: OpenAI

**Implementation:** `providers/openai/openai.go`

**Configuration:**
```go
type OpenAIConfig struct {
    APIKey  string  // API key (or empty for env var)
    BaseURL string  // Custom endpoint (proxy, etc.)
    OrgID   string  // Organization ID
    Model   string  // Embedding model
}
```

**Key Features:**
1. **SDK Integration:** Uses official OpenAI Go SDK
2. **Env Fallback:** Reads `OPENAI_API_KEY` if not provided
3. **Format Conversion:** Converts float64 → float32

**Implementation:**
```go
func (p *OpenAIProvider) EmbedText(text string) ([]float32, error) {
    // Call OpenAI API
    resp, err := p.client.Embeddings.New(context.Background(),
        openai.EmbeddingNewParams{
            Input: openai.F[openai.EmbeddingNewParamsInputUnion](
                openai.EmbeddingNewParamsInputUnionArrayOfStrings([]string{text}),
            ),
            Model: openai.F(p.model),
        },
    )

    // Extract embedding
    embedding64 := resp.Data[0].Embedding

    // Convert float64 → float32
    embedding32 := make([]float32, len(embedding64))
    for i, v := range embedding64 {
        embedding32[i] = float32(v)
    }

    return embedding32, nil
}
```

### Similarity Functions

**Implementation:** Individual files in `similarity/`

**Design Principles:**
1. **Pure Functions:** No state, deterministic
2. **Defensive:** Handle edge cases (empty, mismatched length)
3. **Optimized:** Single-pass algorithms

**Example: Cosine Similarity**
```go
func CosineSimilarity(a, b []float32) float32 {
    // Handle edge cases
    if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
        return 0
    }

    // Single pass: dot product and norms
    var dot, normA, normB float32
    for i := range a {
        dot += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }

    // Avoid division by zero
    if normA == 0 || normB == 0 {
        return 0
    }

    // Cosine = dot / (||a|| * ||b||)
    return dot / (sqrt(normA) * sqrt(normB))
}
```

**Score Ranges:**
- Cosine: [-1, 1] where 1 = identical direction
- Euclidean: [0, 1] where 1 = identical position (normalized)
- Dot Product: Unbounded
- Manhattan: [0, 1] where 1 = identical (normalized)
- Pearson: [-1, 1] where 1 = perfect correlation

---

## Data Flow

### 1. Cache Creation Flow

```
User Code
    │
    ↓
semanticcache.New(opts...)
    │
    ↓
Config.Apply(opts...)
    ├── WithLRUBackend(1000)
    │   └── Creates LRU backend
    ├── WithOpenAIProvider("key")
    │   └── Creates OpenAI provider
    └── WithSimilarityComparator(CosineSimilarity)
        └── Sets similarity func
    │
    ↓
Config.Validate()
    ├── Check backend != nil
    ├── Check provider != nil
    └── Default comparator if nil
    │
    ↓
NewSemanticCache(backend, provider, comparator)
    │
    ↓
Return SemanticCache[K, V]
```

### 2. Set Operation Flow

```
cache.Set(ctx, key, "input text", value)
    │
    ↓
Validate key != zero value
    │
    ↓
provider.EmbedText("input text")
    │
    ↓ (API call)
OpenAI API
    │
    ↓
embedding []float32
    │
    ↓
backend.Set(ctx, key, Entry{embedding, value})
    │
    ↓
Storage (in-memory or Redis)
    │
    ↓
Return error (if any)
```

### 3. Semantic Lookup Flow

```
cache.Lookup(ctx, "query text", 0.8)
    │
    ↓
provider.EmbedText("query text")
    │
    ↓ (API call)
queryEmbedding []float32
    │
    ↓
backend.Keys(ctx)
    │
    ↓
for each key:
    │
    ├── backend.GetEmbedding(ctx, key)
    │   │
    │   ↓
    │   storedEmbedding []float32
    │   │
    │   ↓
    ├── comparator(queryEmbedding, storedEmbedding)
    │   │
    │   ↓
    │   score float32
    │   │
    │   ↓
    └── if score >= threshold:
            track as best match
    │
    ↓
Return best Match[V] or nil
```

### 4. Async Operation Flow

```
cache.SetAsync(ctx, key, text, value)
    │
    ↓
Create buffered channel (size 1)
    │
    ↓
Spawn goroutine:
    ├── embedding := provider.EmbedText(text)
    ├── backend.SetAsync(ctx, key, Entry{embedding, value})
    │   │
    │   ↓
    │   Spawn backend goroutine:
    │       └── backend.Set(ctx, key, entry)
    │   │
    │   ↓
    │   Wait for backend result
    │   │
    │   ↓
    └── Send result to channel
        └── Close channel
    │
    ↓
Return channel (immediately)
    │
    ↓
User receives from channel when ready
```

---

## Concurrency Model

### Thread Safety Guarantees

**All components are safe for concurrent use:**

1. **SemanticCache:** Delegates to thread-safe backends
2. **In-Memory Backends:** Use `sync.RWMutex`
3. **Redis Backend:** Redis client is thread-safe
4. **Providers:** OpenAI client is thread-safe
5. **Similarity Functions:** Pure, stateless

### In-Memory Locking Strategy

```go
type LRUBackend[K, V] struct {
    mu    *sync.RWMutex  // Shared lock
    cache *lru.Cache     // Protected by mu
    index map[K][]float32 // Protected by mu
}

// Write operations: exclusive lock
func (b *LRUBackend[K, V]) Set(...) {
    b.mu.Lock()         // Block all access
    defer b.mu.Unlock()
    // Mutate state
}

// Read operations: shared lock
func (b *LRUBackend[K, V]) Get(...) {
    b.mu.RLock()        // Allow concurrent reads
    defer b.mu.RUnlock()
    // Read state
}
```

**Benefits:**
- Read-heavy workloads scale (multiple concurrent readers)
- Writes are serialized (consistency)
- No deadlocks (defer unlock pattern)

### Async Concurrency

**Pattern:** Goroutine per operation with buffered channel

```go
func (sc *SemanticCache[K, V]) SetAsync(...) <-chan error {
    errCh := make(chan error, 1) // Buffered!
    go func() {
        defer close(errCh)

        // Do work (may block)
        err := sc.Set(...)

        // Send result (never blocks due to buffer)
        errCh <- err
    }()
    return errCh // Return immediately
}
```

**Key Properties:**
1. **Non-blocking:** Returns channel immediately
2. **No goroutine leak:** Goroutine exits after sending
3. **User controls wait:** User decides when to receive
4. **Buffered channel:** Send never blocks (goroutine can exit)

### Batch Async Concurrency

**Pattern:** Goroutine per item with result aggregation

```go
func (sc *SemanticCache[K, V]) SetBatchAsync(ctx, items) <-chan error {
    errCh := make(chan error, 1)
    go func() {
        defer close(errCh)

        resultCh := make(chan setResult, len(items))

        // Spawn goroutine per item
        for _, item := range items {
            go func(it BatchItem[K, V]) {
                embedding, err := sc.provider.EmbedText(it.InputText)
                if err != nil {
                    resultCh <- setResult{err: err}
                    return
                }

                backendErrCh := sc.backend.SetAsync(ctx, it.Key, ...)
                resultCh <- setResult{err: <-backendErrCh}
            }(item)
        }

        // Wait for all (fail-fast)
        for range items {
            result := <-resultCh
            if result.err != nil {
                errCh <- result.err
                return
            }
        }

        errCh <- nil
    }()
    return errCh
}
```

**Benefits:**
- Parallel embedding generation
- Concurrent backend operations
- Fail-fast on first error
- User still gets single channel

---

## Extension Points

### 1. Custom Backend

**Interface to Implement:**
```go
type CacheBackend[K comparable, V any] interface {
    Set(ctx, key, entry) error
    Get(ctx, key) (Entry[V], bool, error)
    Delete(ctx, key) error
    Contains(ctx, key) (bool, error)
    Flush(ctx) error
    Len(ctx) (int, error)
    Keys(ctx) ([]K, error)
    GetEmbedding(ctx, key) ([]float32, bool, error)
    Close() error

    SetAsync(ctx, key, entry) <-chan error
    GetAsync(ctx, key) <-chan AsyncGetResult[V]
    DeleteAsync(ctx, key) <-chan error
    GetBatchAsync(ctx, keys) <-chan AsyncBatchResult[K, V]
}
```

**Example: PostgreSQL with pgvector**
```go
type PgVectorBackend[K comparable, V any] struct {
    db *sql.DB
}

func (b *PgVectorBackend[K, V]) Set(ctx, key, entry) error {
    _, err := b.db.ExecContext(ctx,
        "INSERT INTO cache (key, value, embedding) VALUES ($1, $2, $3)",
        key, entry.Value, entry.Embedding,
    )
    return err
}

// Implement other methods...
```

**Usage:**
```go
cache := semanticcache.New[string, MyData](
    options.WithCustomBackend(pgBackend),
    options.WithOpenAIProvider("key"),
)
```

### 2. Custom Embedding Provider

**Interface to Implement:**
```go
type EmbeddingProvider interface {
    EmbedText(text string) ([]float32, error)
    Close()
}
```

**Example: HuggingFace**
```go
type HuggingFaceProvider struct {
    apiKey string
    model  string
}

func (p *HuggingFaceProvider) EmbedText(text string) ([]float32, error) {
    // Call HuggingFace API
    resp, err := http.Post(
        "https://api-inference.huggingface.co/pipeline/feature-extraction/"+p.model,
        "application/json",
        bytes.NewBuffer(json.Marshal(map[string]string{"inputs": text})),
    )

    // Parse response
    var embedding []float32
    json.NewDecoder(resp.Body).Decode(&embedding)

    return embedding, nil
}

func (p *HuggingFaceProvider) Close() {}
```

**Usage:**
```go
provider := &HuggingFaceProvider{
    apiKey: os.Getenv("HF_API_KEY"),
    model:  "sentence-transformers/all-MiniLM-L6-v2",
}

cache := semanticcache.New[string, string](
    options.WithLRUBackend(1000),
    options.WithCustomProvider(provider),
)
```

### 3. Custom Similarity Function

**Type to Implement:**
```go
type SimilarityFunc func(a, b []float32) float32
```

**Example: Weighted Cosine**
```go
func WeightedCosineSimilarity(weights []float32) similarity.SimilarityFunc {
    return func(a, b []float32) float32 {
        if len(a) != len(b) || len(a) != len(weights) {
            return 0
        }

        var dot, normA, normB float32
        for i := range a {
            wa := a[i] * weights[i]
            wb := b[i] * weights[i]
            dot += wa * wb
            normA += wa * wa
            normB += wb * wb
        }

        if normA == 0 || normB == 0 {
            return 0
        }

        return dot / (sqrt(normA) * sqrt(normB))
    }
}
```

**Usage:**
```go
// Emphasize first 100 dimensions
weights := make([]float32, 1536)
for i := range weights {
    if i < 100 {
        weights[i] = 2.0
    } else {
        weights[i] = 1.0
    }
}

cache := semanticcache.New[string, string](
    options.WithLRUBackend(1000),
    options.WithOpenAIProvider("key"),
    options.WithSimilarityComparator(WeightedCosineSimilarity(weights)),
)
```

---

## Design Decisions

### Why Generics?

**Decision:** Use Go 1.18+ generics for key/value types

**Rationale:**
- Type safety without code duplication
- Better developer experience (no type assertions)
- Compile-time error detection
- Modern Go practice

**Alternative Considered:** Interface-based (`any` types)
- Would require type assertions everywhere
- Runtime errors instead of compile-time
- Poor IDE support

### Why Functional Options?

**Decision:** Use functional options pattern for configuration

**Rationale:**
- Variable number of options
- Options can fail and return errors
- Easy to extend (add new options)
- Self-documenting API

**Alternative Considered:** Config struct
```go
// Not as flexible
cache := New(Config{
    Backend: lruBackend,
    Provider: openaiProvider,
    // What if backend creation fails?
})
```

### Why Channel-Based Async?

**Decision:** Return channels from async operations

**Rationale:**
- Idiomatic Go concurrency
- User controls when to wait
- Composable with select statements
- No callback hell

**Alternative Considered:** Callback functions
```go
// Not idiomatic Go
cache.SetAsync(ctx, key, text, value, func(err error) {
    // Callback hell
})
```

### Why Separate Index Map?

**Decision:** Maintain separate embedding index in in-memory backends

**Rationale:**
- O(1) embedding lookup (common operation)
- Avoids deserializing full entry for similarity computation
- Trade memory for speed

**Cost:** Additional memory (~8 bytes per entry + embedding size)

### Why Not Use Redis Vector Index?

**Decision:** Redis backend creates FT.SEARCH index but doesn't use it

**Rationale:**
- Current implementation: Simple, works for all backends
- Future: Can optimize with KNN queries
- Keeps API consistent across backends

**TODO:** Optimize Redis Lookup to use FT.SEARCH KNN

---

## Future Architecture Improvements

### 1. Vector Indexing

**Goal:** Sub-linear semantic search

**Approach:**
- In-memory: HNSW index (hnswlib or custom)
- Redis: Utilize FT.SEARCH KNN queries

**Impact:**
- Lookup: O(log n) instead of O(n)
- Scales to millions of entries

### 2. Streaming Results

**Goal:** Progressive result delivery

**Approach:**
```go
func (sc *SemanticCache[K, V]) TopMatchesStream(
    ctx context.Context,
    inputText string,
) <-chan Match[V] {
    matchCh := make(chan Match[V])
    go func() {
        defer close(matchCh)
        // Send matches as found
        for match := range findMatches(...) {
            select {
            case matchCh <- match:
            case <-ctx.Done():
                return
            }
        }
    }()
    return matchCh
}
```

**Impact:**
- Lower latency for first result
- Better UX for long searches

### 3. Pluggable Serialization

**Goal:** Support different serialization formats

**Approach:**
```go
type Serializer[V any] interface {
    Serialize(V) ([]byte, error)
    Deserialize([]byte) (V, error)
}

// Options
WithJSONSerializer[K, V]()
WithProtobufSerializer[K, V]()
WithMsgpackSerializer[K, V]()
```

**Impact:**
- Reduce storage size
- Faster serialization
- Support binary values

### 4. Observability Hooks

**Goal:** Enable metrics and tracing

**Approach:**
```go
type Observer interface {
    OnSet(key, duration, error)
    OnGet(key, found, duration, error)
    OnLookup(query, matches, duration, error)
}

options.WithObserver(prometheusObserver)
```

**Impact:**
- Production monitoring
- Performance analysis
- Debugging support
