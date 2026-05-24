# openai

OpenAI embedding provider. Uses the official `openai-go` SDK.

## Usage

```go
p, err := openai.NewOpenAIProvider(openai.OpenAIConfig{
    APIKey: "sk-...",
    Model:  "text-embedding-3-small",  // optional, this is the default
})
```

If `APIKey` is empty, falls back to the `OPENAI_API_KEY` environment variable.

## Configuration

| Field | Description |
|-------|-------------|
| `APIKey` | OpenAI API key |
| `Model` | Embedding model (default: `text-embedding-3-small`) |
| `BaseURL` | Custom API base URL |
| `OrgID` | OpenAI organization ID |

## Batch support

Implements `types.BatchEmbeddingProvider`. `EmbedBatch` sends up to 2048 texts in a single API call.
