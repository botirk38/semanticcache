# semanticcache

A Go library for **semantic caching** with LRU eviction, supporting vector-based similarity search with pluggable embedding backends (local or cloud).
Useful for AI, retrieval-augmented generation, chatbot memory, and any workload where you want to cache by semantic similarity—not just string keys.

---

## Project Overview

`semanticcache` is a pluggable semantic cache for Go.
It lets you cache arbitrary values indexed by **embeddings** (vector representations of text, etc.), using the [Least Recently Used (LRU)](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_%28LRU%29) cache eviction policy for scalability.
Supports both **local embedding models** (via [kelindar/search](https://github.com/kelindar/search)) and **OpenAI Embedding API**.

---

## Key Features

- **Semantic Cache:** Store and retrieve values by semantic similarity, not just by string key.
- **LRU Eviction:** Keeps cache memory usage bounded, evicts least recently used items.
- **Pluggable Embedding Providers:** Use local models (fast, private) or OpenAI (no local install).
- **Fast Similarity Search:** Lookup or rank by vector similarity (defaults to cosine).
- **Customizable Capacity:** Choose your cache size for your workload.
- **Custom Similarity:** Use your own similarity function if needed.

---

## Installation

```sh
go get github.com/botirk38/semanticcache
```

---

## Embedding Provider Requirements

- For **local embedding**, you need the `libllama_go.so` shared library built and available (see [kelindar/search docs](https://github.com/kelindar/search#compile-library)).
- For **OpenAI provider**, set the `OPENAI_API_KEY` environment variable.

---

## Usage

### 1. Choose your provider

#### **Local Embedding Provider**

```go
import (
    "github.com/botirk38/semanticcache/semanticcache"
)

provider, err := semanticcache.NewLocalProvider("", 0) // (modelPath, gpuLayers)
if err != nil {
    panic(err)
}
defer provider.Close()
```

#### **OpenAI Embedding Provider**

```go
provider, err := semanticcache.NewOpenAIProvider("", "") // (apiKey, model)
if err != nil {
    panic(err)
}
defer provider.Close()
```

---

### 2. Create your cache

```go
cache, err := semanticcache.NewSemanticCache(1000, provider, nil) // (capacity, provider, comparator)
if err != nil {
    panic(err)
}
defer cache.Close()
```

- Pass your own comparator for custom similarity, or use `nil` for cosine similarity.

---

### 3. Add entries to the cache

```go
// The Set method takes a unique key, the text to embed, and your value.
err := cache.Set("unique-key", "What is the capital of France?", "Paris")
if err != nil {
    panic(err)
}
```

---

### 4. Retrieve by semantic similarity

#### Exact key

```go
value, ok := cache.Get("unique-key")
if ok {
    // do something with value
}
```

#### Semantic lookup

```go
value, ok, err := cache.Lookup("Which city is France's capital?", 0.8)
if err != nil {
    panic(err)
}
if ok {
    // value is the closest matching cached value above threshold
}
```

#### Top N matches

```go
matches, err := cache.TopMatches("french capital", 3)
if err != nil {
    panic(err)
}
for _, m := range matches {
    // m.Value and m.Score
}
```

---

## API Overview

**Types:**

- `SemanticCache`: The main cache struct.
- `Entry`: Holds the embedding and value.
- `Match`: A value and similarity score for ranked results.
- `EmbeddingProvider`: Interface for embedding backends.

**Key Methods:**

- `Set(key string, text string, value any) error`
- `Get(key string) (any, bool)`
- `Lookup(text string, threshold float32) (any, bool, error)`
- `TopMatches(text string, n int) ([]Match, error)`
- `Flush() error`
- `Len() int`
- `Close()`

---

## Architecture

- **SemanticCache**: Handles LRU cache, in-memory embedding index, and similarity logic.
- **Providers**: Pluggable, implement `EmbeddingProvider` interface (`EmbedText(text string) ([]float32, error)`).
- **Comparator**: Any function of type `func(a, b []float32) float32` (cosine by default).

---

## Testing

Run:

```sh
go test -v ./...
```

To run tests that require local embedding, ensure you have `libllama_go.so` installed and available.

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

- [github.com/hashicorp/golang-lru/v2](https://github.com/hashicorp/golang-lru) – LRU cache
- [github.com/kelindar/search](https://github.com/kelindar/search) – (optional) local vector search/embedding
- [github.com/openai/openai-go](https://github.com/openai/openai-go) – (optional) OpenAI embedding provider

---

## Compatibility

- Requires Go 1.18 or newer.
- Supports Linux, macOS, Windows (embedding provider support may vary).

---

**Questions?** Open an issue or discussion on GitHub!
**Need an install script or provider for Cohere/Gemini? PRs welcome!**
