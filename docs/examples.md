# Examples & Tutorials

Practical examples demonstrating SemanticCache usage patterns.

## Table of Contents

- [Basic Examples](#basic-examples)
- [Advanced Use Cases](#advanced-use-cases)
- [Integration Examples](#integration-examples)
- [Custom Implementations](#custom-implementations)

---

## Basic Examples

### Example 1: Simple FAQ System

Cache frequently asked questions with semantic lookup.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
)

type FAQAnswer struct {
    Answer   string
    Category string
    URL      string
}

func main() {
    // Create cache
    cache, err := semanticcache.New[string, FAQAnswer](
        options.WithOpenAIProvider(""),  // Uses OPENAI_API_KEY env var
        options.WithLRUBackend(1000),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // Populate FAQ cache
    faqs := map[string]FAQAnswer{
        "reset_password": {
            Answer:   "Go to Settings > Security > Reset Password",
            Category: "Account",
            URL:      "/docs/account/reset-password",
        },
        "billing_info": {
            Answer:   "View your billing in Settings > Billing",
            Category: "Billing",
            URL:      "/docs/billing/view-invoices",
        },
        "api_limits": {
            Answer:   "API rate limits are 1000 requests per minute",
            Category: "API",
            URL:      "/docs/api/rate-limits",
        },
    }

    for id, faq := range faqs {
        err := cache.Set(ctx, id, faq.Answer, faq)
        if err != nil {
            log.Printf("Failed to cache FAQ %s: %v", id, err)
        }
    }

    // User queries
    queries := []string{
        "How do I change my password?",
        "Where can I see my invoices?",
        "What are the rate limits for the API?",
    }

    for _, query := range queries {
        match, err := cache.Lookup(ctx, query, 0.8)
        if err != nil {
            log.Printf("Lookup failed: %v", err)
            continue
        }

        if match != nil {
            fmt.Printf("Q: %s\n", query)
            fmt.Printf("A: %s\n", match.Value.Answer)
            fmt.Printf("   Category: %s | Docs: %s\n",
                match.Value.Category, match.Value.URL)
            fmt.Printf("   Confidence: %.2f\n\n", match.Score)
        } else {
            fmt.Printf("Q: %s\nA: No answer found\n\n", query)
        }
    }
}
```

**Output:**
```
Q: How do I change my password?
A: Go to Settings > Security > Reset Password
   Category: Account | Docs: /docs/account/reset-password
   Confidence: 0.87

Q: Where can I see my invoices?
A: View your billing in Settings > Billing
   Category: Billing | Docs: /docs/billing/view-invoices
   Confidence: 0.82

Q: What are the rate limits for the API?
A: API rate limits are 1000 requests per minute
   Category: API | Docs: /docs/api/rate-limits
   Confidence: 0.91
```

---

### Example 2: LLM Response Caching

Cache LLM responses to avoid redundant API calls.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
)

type LLMResponse struct {
    Response  string
    Model     string
    Timestamp time.Time
    TokensUsed int
}

// Simulated LLM API call (expensive)
func callLLM(prompt string) (LLMResponse, error) {
    // Simulate API latency
    time.Sleep(2 * time.Second)

    return LLMResponse{
        Response:  "Generated response for: " + prompt,
        Model:     "gpt-4",
        Timestamp: time.Now(),
        TokensUsed: 150,
    }, nil
}

func getCachedOrGenerateResponse(
    ctx context.Context,
    cache *semanticcache.SemanticCache[string, LLMResponse],
    prompt string,
) (LLMResponse, bool, error) {
    // Try semantic lookup first
    match, err := cache.Lookup(ctx, prompt, 0.85)
    if err != nil {
        return LLMResponse{}, false, err
    }

    if match != nil {
        fmt.Printf("✅ Cache hit (score: %.2f)\n", match.Score)
        return match.Value, true, nil
    }

    // Cache miss - generate response
    fmt.Println("❌ Cache miss - calling LLM...")
    response, err := callLLM(prompt)
    if err != nil {
        return LLMResponse{}, false, err
    }

    // Cache for future use
    cacheKey := fmt.Sprintf("prompt_%d", time.Now().Unix())
    if err := cache.Set(ctx, cacheKey, prompt, response); err != nil {
        log.Printf("Failed to cache response: %v", err)
    }

    return response, false, nil
}

func main() {
    cache, err := semanticcache.New[string, LLMResponse](
        options.WithOpenAIProvider(""),
        options.WithLRUBackend(100),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // First request (cache miss)
    start := time.Now()
    resp1, cached, err := getCachedOrGenerateResponse(ctx, cache,
        "What is the capital of France?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Response: %s\n", resp1.Response)
    fmt.Printf("Latency: %v\n", time.Since(start))
    fmt.Printf("Cached: %v\n\n", cached)

    // Similar request (cache hit)
    start = time.Now()
    resp2, cached, err := getCachedOrGenerateResponse(ctx, cache,
        "Tell me the capital city of France")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Response: %s\n", resp2.Response)
    fmt.Printf("Latency: %v\n", time.Since(start))
    fmt.Printf("Cached: %v\n", cached)
}
```

**Output:**
```
❌ Cache miss - calling LLM...
Response: Generated response for: What is the capital of France?
Latency: 2.05s
Cached: false

✅ Cache hit (score: 0.92)
Response: Generated response for: What is the capital of France?
Latency: 150ms
Cached: true
```

---

### Example 3: Product Search with Semantic Matching

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
)

type Product struct {
    ID          string
    Name        string
    Description string
    Price       float64
    Category    string
}

func main() {
    cache, err := semanticcache.New[string, Product](
        options.WithOpenAIProvider(""),
        options.WithLFUBackend(500),  // Keep frequently searched items
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // Index products
    products := []Product{
        {
            ID:          "laptop-001",
            Name:        "MacBook Pro 16\"",
            Description: "High-performance laptop with M3 chip, 32GB RAM, perfect for developers",
            Price:       2999.99,
            Category:    "Laptops",
        },
        {
            ID:          "phone-001",
            Name:        "iPhone 15 Pro",
            Description: "Latest smartphone with advanced camera and A17 chip",
            Price:       1199.99,
            Category:    "Phones",
        },
        {
            ID:          "tablet-001",
            Name:        "iPad Pro 12.9\"",
            Description: "Powerful tablet with M2 chip, great for creative work",
            Price:       1099.99,
            Category:    "Tablets",
        },
    }

    for _, product := range products {
        // Use description for semantic embedding
        err := cache.Set(ctx, product.ID, product.Description, product)
        if err != nil {
            log.Printf("Failed to index product %s: %v", product.ID, err)
        }
    }

    // Natural language product search
    searches := []string{
        "best laptop for programming",
        "phone with good camera",
        "device for digital art",
    }

    for _, search := range searches {
        fmt.Printf("Search: %s\n", search)

        matches, err := cache.TopMatches(ctx, search, 3)
        if err != nil {
            log.Printf("Search failed: %v", err)
            continue
        }

        if len(matches) == 0 {
            fmt.Println("No matches found\n")
            continue
        }

        for i, match := range matches {
            fmt.Printf("%d. %s - $%.2f (score: %.2f)\n",
                i+1, match.Value.Name, match.Value.Price, match.Score)
        }
        fmt.Println()
    }
}
```

---

## Advanced Use Cases

### Example 4: Async Batch Processing

Process large datasets efficiently with async operations.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
)

type Document struct {
    ID      string
    Content string
}

func main() {
    cache, err := semanticcache.New[string, string](
        options.WithOpenAIProvider(""),
        options.WithRedisBackend("localhost:6379", 0),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // Large dataset to process
    documents := generateDocuments(10000)

    // Process in batches asynchronously
    const batchSize = 100
    start := time.Now()

    for i := 0; i < len(documents); i += batchSize {
        end := i + batchSize
        if end > len(documents) {
            end = len(documents)
        }

        batch := documents[i:end]
        items := make([]semanticcache.BatchItem[string, string], len(batch))

        for j, doc := range batch {
            items[j] = semanticcache.BatchItem[string, string]{
                Key:       doc.ID,
                InputText: doc.Content,
                Value:     doc.Content,
            }
        }

        // Async batch operation
        errCh := cache.SetBatchAsync(ctx, items)

        // Do other work while batch processes...

        // Wait for batch completion
        if err := <-errCh; err != nil {
            log.Printf("Batch %d failed: %v", i/batchSize, err)
            continue
        }

        fmt.Printf("Processed batch %d/%d\n",
            (i/batchSize)+1, (len(documents)+batchSize-1)/batchSize)
    }

    duration := time.Since(start)
    fmt.Printf("\n✅ Processed %d documents in %v\n", len(documents), duration)
    fmt.Printf("   Throughput: %.0f docs/sec\n",
        float64(len(documents))/duration.Seconds())
}

func generateDocuments(count int) []Document {
    docs := make([]Document, count)
    for i := 0; i < count; i++ {
        docs[i] = Document{
            ID:      fmt.Sprintf("doc_%d", i),
            Content: fmt.Sprintf("Document content number %d with some text", i),
        }
    }
    return docs
}
```

---

### Example 5: Multi-Similarity Comparison

Compare different similarity algorithms for your use case.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
    "github.com/botirk38/semanticcache/similarity"
)

func testSimilarityFunction(
    name string,
    simFunc similarity.SimilarityFunc,
    query string,
) {
    cache, err := semanticcache.New[string, string](
        options.WithOpenAIProvider(""),
        options.WithLRUBackend(100),
        options.WithSimilarityComparator(simFunc),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cache.Close()

    ctx := context.Background()

    // Sample data
    cache.Set(ctx, "1", "The weather is sunny today", "Sunny weather")
    cache.Set(ctx, "2", "It's raining heavily outside", "Rainy weather")
    cache.Set(ctx, "3", "Beautiful blue skies", "Clear skies")

    // Query
    matches, err := cache.TopMatches(ctx, query, 3)
    if err != nil {
        log.Printf("%s failed: %v", name, err)
        return
    }

    fmt.Printf("%s:\n", name)
    for i, match := range matches {
        fmt.Printf("  %d. %s (score: %.3f)\n",
            i+1, match.Value, match.Score)
    }
    fmt.Println()
}

func main() {
    query := "Nice sunny day outside"

    similarities := map[string]similarity.SimilarityFunc{
        "Cosine":        similarity.CosineSimilarity,
        "Euclidean":     similarity.EuclideanSimilarity,
        "Dot Product":   similarity.DotProductSimilarity,
        "Manhattan":     similarity.ManhattanSimilarity,
        "Pearson":       similarity.PearsonCorrelationSimilarity,
    }

    for name, simFunc := range similarities {
        testSimilarityFunction(name, simFunc, query)
    }
}
```

**Output:**
```
Cosine:
  1. Sunny weather (score: 0.912)
  2. Clear skies (score: 0.786)
  3. Rainy weather (score: 0.543)

Euclidean:
  1. Sunny weather (score: 0.887)
  2. Clear skies (score: 0.801)
  3. Rainy weather (score: 0.634)

Dot Product:
  1. Sunny weather (score: 245.821)
  2. Clear skies (score: 198.345)
  3. Rainy weather (score: 134.567)

...
```

---

## Integration Examples

### Example 6: HTTP API Server

Integrate semantic cache into a web service.

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
)

type SearchRequest struct {
    Query     string  `json:"query"`
    Threshold float32 `json:"threshold"`
}

type SearchResponse struct {
    Results []Result `json:"results"`
}

type Result struct {
    Content string  `json:"content"`
    Score   float32 `json:"score"`
}

type Server struct {
    cache *semanticcache.SemanticCache[string, string]
}

func NewServer() (*Server, error) {
    cache, err := semanticcache.New[string, string](
        options.WithOpenAIProvider(""),
        options.WithRedisBackend("localhost:6379", 0),
    )
    if err != nil {
        return nil, err
    }

    return &Server{cache: cache}, nil
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
    var req SearchRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Use request context (includes cancellation)
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    // Semantic search
    matches, err := s.cache.TopMatches(ctx, req.Query, 10)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Filter by threshold
    results := make([]Result, 0)
    for _, match := range matches {
        if match.Score >= req.Threshold {
            results = append(results, Result{
                Content: match.Value,
                Score:   match.Score,
            })
        }
    }

    json.NewEncoder(w).Encode(SearchResponse{Results: results})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
    var data struct {
        ID      string `json:"id"`
        Content string `json:"content"`
    }

    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    // Non-blocking index
    errCh := s.cache.SetAsync(ctx, data.ID, data.Content, data.Content)

    // Return immediately
    w.WriteHeader(http.StatusAccepted)

    // Log result asynchronously
    go func() {
        if err := <-errCh; err != nil {
            log.Printf("Failed to index %s: %v", data.ID, err)
        }
    }()
}

func main() {
    server, err := NewServer()
    if err != nil {
        log.Fatal(err)
    }
    defer server.cache.Close()

    http.HandleFunc("/search", server.handleSearch)
    http.HandleFunc("/index", server.handleIndex)

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Usage:**
```bash
# Index content
curl -X POST http://localhost:8080/index \
  -H "Content-Type: application/json" \
  -d '{"id": "doc1", "content": "How to reset password"}'

# Search
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query": "password reset help", "threshold": 0.8}'
```

---

## Custom Implementations

### Example 7: Custom Embedding Provider (Ollama)

Use local models with Ollama.

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/options"
    "github.com/botirk38/semanticcache/types"
)

type OllamaProvider struct {
    baseURL string
    model   string
    client  *http.Client
}

type ollamaRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
}

type ollamaResponse struct {
    Embedding []float32 `json:"embedding"`
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
    return &OllamaProvider{
        baseURL: baseURL,
        model:   model,
        client:  &http.Client{},
    }
}

func (p *OllamaProvider) EmbedText(text string) ([]float32, error) {
    reqBody, _ := json.Marshal(ollamaRequest{
        Model:  p.model,
        Prompt: text,
    })

    resp, err := p.client.Post(
        p.baseURL+"/api/embeddings",
        "application/json",
        bytes.NewBuffer(reqBody),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result ollamaResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Embedding, nil
}

func (p *OllamaProvider) Close() {}

func main() {
    // Create custom provider
    provider := NewOllamaProvider("http://localhost:11434", "llama2")

    // Use with cache
    cache, err := semanticcache.New[string, string](
        options.WithCustomProvider(provider),
        options.WithLRUBackend(1000),
    )
    if err != nil {
        panic(err)
    }
    defer cache.Close()

    // Use cache normally
    ctx := context.Background()
    cache.Set(ctx, "key1", "Local embedding test", "value1")

    match, _ := cache.Lookup(ctx, "embedding test locally", 0.8)
    if match != nil {
        fmt.Printf("Found: %s\n", match.Value)
    }
}
```

---

### Example 8: Custom Backend (SQLite)

Implement a custom backend with SQLite.

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"

    "github.com/botirk38/semanticcache/types"
    _ "github.com/mattn/go-sqlite3"
)

type SQLiteBackend[K comparable, V any] struct {
    db *sql.DB
}

func NewSQLiteBackend[K comparable, V any](dbPath string) (*SQLiteBackend[K, V], error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // Create tables
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS cache (
            key TEXT PRIMARY KEY,
            value TEXT,
            embedding TEXT
        )
    `)
    if err != nil {
        return nil, err
    }

    return &SQLiteBackend[K, V]{db: db}, nil
}

func (b *SQLiteBackend[K, V]) Set(
    ctx context.Context,
    key K,
    entry types.Entry[V],
) error {
    keyStr := fmt.Sprintf("%v", key)
    valueJSON, _ := json.Marshal(entry.Value)
    embeddingJSON, _ := json.Marshal(entry.Embedding)

    _, err := b.db.ExecContext(ctx,
        "INSERT OR REPLACE INTO cache (key, value, embedding) VALUES (?, ?, ?)",
        keyStr, valueJSON, embeddingJSON,
    )
    return err
}

func (b *SQLiteBackend[K, V]) Get(
    ctx context.Context,
    key K,
) (types.Entry[V], bool, error) {
    keyStr := fmt.Sprintf("%v", key)

    var valueJSON, embeddingJSON string
    err := b.db.QueryRowContext(ctx,
        "SELECT value, embedding FROM cache WHERE key = ?",
        keyStr,
    ).Scan(&valueJSON, &embeddingJSON)

    if err == sql.ErrNoRows {
        return types.Entry[V]{}, false, nil
    }
    if err != nil {
        return types.Entry[V]{}, false, err
    }

    var value V
    var embedding []float32
    json.Unmarshal([]byte(valueJSON), &value)
    json.Unmarshal([]byte(embeddingJSON), &embedding)

    return types.Entry[V]{
        Value:     value,
        Embedding: embedding,
    }, true, nil
}

// Implement other CacheBackend methods...
// (Delete, Contains, Flush, Len, Keys, GetEmbedding, Close, async methods)

func (b *SQLiteBackend[K, V]) SetAsync(
    ctx context.Context,
    key K,
    entry types.Entry[V],
) <-chan error {
    errCh := make(chan error, 1)
    go func() {
        defer close(errCh)
        errCh <- b.Set(ctx, key, entry)
    }()
    return errCh
}

// ... implement other async methods similarly

func main() {
    backend, err := NewSQLiteBackend[string, string]("cache.db")
    if err != nil {
        panic(err)
    }

    cache, err := semanticcache.New[string, string](
        options.WithCustomBackend(backend),
        options.WithOpenAIProvider(""),
    )
    if err != nil {
        panic(err)
    }
    defer cache.Close()

    // Use cache with persistent SQLite backend
    // ...
}
```

---

### Example 9: Monitoring Wrapper

Add observability to cache operations.

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/types"
)

type MonitoredCache[K comparable, V any] struct {
    cache *semanticcache.SemanticCache[K, V]

    // Metrics
    hits   int64
    misses int64
    errors int64
}

func NewMonitoredCache[K comparable, V any](
    cache *semanticcache.SemanticCache[K, V],
) *MonitoredCache[K, V] {
    return &MonitoredCache[K, V]{cache: cache}
}

func (m *MonitoredCache[K, V]) Get(
    ctx context.Context,
    key K,
) (V, bool, error) {
    start := time.Now()
    value, found, err := m.cache.Get(ctx, key)
    duration := time.Since(start)

    if err != nil {
        m.errors++
        log.Printf("[CACHE] Get error: key=%v duration=%v err=%v",
            key, duration, err)
    } else if found {
        m.hits++
        log.Printf("[CACHE] Hit: key=%v duration=%v", key, duration)
    } else {
        m.misses++
        log.Printf("[CACHE] Miss: key=%v duration=%v", key, duration)
    }

    return value, found, err
}

func (m *MonitoredCache[K, V]) Lookup(
    ctx context.Context,
    inputText string,
    threshold float32,
) (*semanticcache.Match[V], error) {
    start := time.Now()
    match, err := m.cache.Lookup(ctx, inputText, threshold)
    duration := time.Since(start)

    log.Printf("[CACHE] Lookup: query=%q threshold=%.2f duration=%v found=%v",
        inputText, threshold, duration, match != nil)

    return match, err
}

func (m *MonitoredCache[K, V]) Stats() map[string]interface{} {
    total := m.hits + m.misses
    hitRatio := 0.0
    if total > 0 {
        hitRatio = float64(m.hits) / float64(total)
    }

    return map[string]interface{}{
        "hits":      m.hits,
        "misses":    m.misses,
        "errors":    m.errors,
        "total":     total,
        "hit_ratio": hitRatio,
    }
}

func main() {
    cache, _ := semanticcache.New[string, string](...)
    monitored := NewMonitoredCache(cache)

    // Use monitored cache
    monitored.Get(context.Background(), "key1")
    monitored.Lookup(context.Background(), "query", 0.8)

    // Print stats
    log.Printf("Cache stats: %+v", monitored.Stats())
}
```

**Output:**
```
[CACHE] Hit: key=key1 duration=234µs
[CACHE] Lookup: query="search term" threshold=0.80 duration=45ms found=true
[CACHE] Cache stats: map[errors:0 hit_ratio:1 hits:1 misses:0 total:1]
```

---

## Tips for Production

### 1. Connection Pooling
```go
// Redis: Tune pool size for your workload
config := types.BackendConfig{
    ConnectionString: "localhost:6379",
    Options: map[string]any{
        "pool_size": 100,
    },
}
```

### 2. Graceful Shutdown
```go
func main() {
    cache, _ := semanticcache.New[string, string](...)

    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Shutting down...")
        cache.Close()
        os.Exit(0)
    }()

    // Run application...
}
```

### 3. Rate Limiting
```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(100), 10)  // 100 req/sec, burst 10

func rateLimitedSet(ctx context.Context, ...) error {
    if err := limiter.Wait(ctx); err != nil {
        return err
    }
    return cache.Set(ctx, ...)
}
```
