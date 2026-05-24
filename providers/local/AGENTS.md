# local -- Agent Instructions

## What this package does
Provides a hash-based `EmbeddingProvider` for testing and development. Uses FNV hashing to produce deterministic, L2-normalised vectors from text. No API key or network required.

## Key patterns
- `New(dimensions)` creates a provider. 0 or negative defaults to 128.
- Vectors are deterministic: same text always produces the same vector.
- Vectors are L2-normalised so cosine similarity is well-defined.
- NOT semantically meaningful -- do not use for production search.

## Rules
- Keep this dependency-free (stdlib only).
- Implements both `EmbeddingProvider` and `BatchEmbeddingProvider`.

## Testing
```
go test ./providers/local/
```
