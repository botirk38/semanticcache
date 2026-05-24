# openai -- Agent Instructions

## What this package does
Implements `types.EmbeddingProvider` and `types.BatchEmbeddingProvider` using the OpenAI embeddings API via the official `openai-go` SDK.

## Key patterns
- Falls back to `OPENAI_API_KEY` env var if APIKey is empty.
- Default model: `text-embedding-3-small`.
- `EmbedBatch` has a hard limit of 2048 texts per call (OpenAI limit).

## Rules
- Do not change the SDK import from `openai-go/v2`.
- Tests use constructor validation only (no live API calls). Do not add tests that require a real API key.

## Testing
```
go test ./providers/openai/
```
