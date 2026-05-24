# local

Hash-based embedding provider for testing and development. Produces deterministic vectors from text using FNV hashing. No API key or network access required.

The embeddings are NOT semantically meaningful. Two sentences with similar meaning will NOT produce similar vectors. Use this only for:

- Unit tests
- Benchmarks
- Local development where you need a provider that satisfies the interface

## Usage

```go
p := local.New(128)  // 128-dimensional vectors

vec, err := p.EmbedText(ctx, "hello world")
// vec is deterministic: same input always produces the same output
// vec is L2-normalised to unit length
```

Pass `0` or a negative number to get the default of 128 dimensions.
