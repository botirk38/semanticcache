# chunker -- Agent Instructions

## What this package does
Text chunking utilities for splitting long text into token-limited pieces before embedding.

## Key types
- `Chunker` interface: `ChunkText`, `CountTokens`, `GetMaxTokens`
- `ChunkConfig`: `MaxTokens`, `ChunkSize`, `ChunkOverlap`, `Strategy`
- `Chunk`: `Text`, `StartToken`, `EndToken`, `Index`

## Rules
- Errors are defined in `errors.go` within this package.
- `DefaultChunkConfig()` returns sensible defaults (8191 max tokens, 512 chunk size, 50 overlap).
- Validate config before use with `ChunkConfig.Validate()`.

## Testing
```
go test ./chunker/
```
