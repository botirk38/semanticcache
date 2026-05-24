# providers -- Agent Instructions

## What this package does
Re-exports provider constructors from subpackages.

## Subpackages
- `openai/` -- OpenAI embedding API
- `local/` -- hash-based provider for testing

## Rules
- When adding a new provider subpackage, add a re-export here.
- Every provider must implement `types.EmbeddingProvider`.
- Optionally implement `types.BatchEmbeddingProvider`.
- Add tests using `httptest` for HTTP-based providers, or simple unit tests for local providers.
