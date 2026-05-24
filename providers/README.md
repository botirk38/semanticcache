# providers

Embedding provider implementations. Each subpackage implements `types.EmbeddingProvider`.

## Subpackages

- `openai/` -- OpenAI embedding API
- `local/` -- deterministic hash-based provider for testing (no API key needed)

## Implementing a provider

Implement the `types.EmbeddingProvider` interface:

```go
type EmbeddingProvider interface {
    EmbedText(ctx context.Context, text string) ([]float64, error)
    Close() error
}
```

Optionally implement `types.BatchEmbeddingProvider` for batch support.
