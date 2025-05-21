# semanticcache

> An in-memory, similarity-aware LRU cache for semantically indexed data in Go.

`semanticcache` is a Go library that enables fast semantic search with pluggable embedding support. It stores vectors (embeddings) and associated responses, allowing efficient nearest-neighbor lookups using any similarity function.

---

## 🚀 Features

- 🔁 **LRU eviction policy** using [hashicorp/golang-lru](https://github.com/hashicorp/golang-lru)
- 🔍 **Semantic search support** via cosine or custom similarity
- ⚡ **Concurrent safe** with read/write locks
- 🔌 **Plug-and-play**: works with any embedding generator (OpenAI, Groq, etc.)
- 🧪 **Tested**, **minimal**, and **modular**

---

## 📦 Installation

```bash
go get github.com/botirk38/semanticcache
```

---

## ✨ Usage Example

```go
package main

import (
 "fmt"
 "log"

 "github.com/botirk38/semanticcache"
)

func cosineSimilarity(a, b []float32) float32 {
 var dot, normA, normB float32
 for i := range a {
  dot += a[i] * b[i]
  normA += a[i] * a[i]
  normB += b[i] * b[i]
 }
 if normA == 0 || normB == 0 {
  return 0
 }
 return dot / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float32) float32 {
 z := x
 for i := 0; i < 10; i++ {
  z -= (z*z - x) / (2 * z)
 }
 return z
}

func main() {
 cache, err := semanticcache.New(100, cosineSimilarity)
 if err != nil {
  log.Fatal(err)
 }

 embedding := []float32{0.1, 0.2, 0.3}
 cache.Set("hello", embedding, "world")

 query := []float32{0.1, 0.2, 0.3}
 resp, found := cache.Lookup(query, 0.95)
 if found {
  fmt.Println("Found:", resp)
 } else {
  fmt.Println("No match found.")
 }
}
```

---

## 🧠 API

### `New(capacity int, comparator Comparator) (*SemanticCache, error)`

Creates a new cache with the given size and similarity function.

### `Set(key string, embedding []float32, response any) error`

Inserts a key, embedding, and response.

### `Lookup(embedding []float32, threshold float32) (any, bool)`

Searches for the first item with similarity ≥ `threshold`.

### `TopMatches(embedding []float32, n int) []Match`

Returns top `n` most similar items.

### `Delete(key string)`, `Flush()`, `Get(key string)`, `Contains(key string)`, `Len()`

Standard cache operations.

---

## 🧪 Testing

Run unit tests:

```bash
go test ./...
```

---

## 📁 Project Structure

```
semanticcache/
├── semanticcache/        # Main library code
│   └── cache.go
├── examples/             # Optional usage demos
│   └── basic/main.go
├── test/                 # Tests
│   └── cache_test.go
```

---

## 📄 License

MIT © [Botir Khaltaev](https://github.com/botirk38)

```

```
