# chunker

Text chunking utilities for splitting long text into smaller pieces before embedding.

## Interface

```go
type Chunker interface {
    ChunkText(text string) ([]Chunk, error)
    CountTokens(text string) (int, error)
    GetMaxTokens() int
}
```

## Strategies

- `FixedSizeOverlap` -- splits text into fixed-size token chunks with configurable overlap

## Configuration

```go
cfg := chunker.DefaultChunkConfig()
// cfg.MaxTokens    = 8191   (OpenAI text-embedding-3-small limit)
// cfg.ChunkSize    = 512
// cfg.ChunkOverlap = 50
// cfg.Strategy     = FixedSizeOverlap
```

## Errors

Defined in `chunker/errors.go`: `ErrInvalidChunkSize`, `ErrChunkSizeExceedsMax`, `ErrInvalidOverlap`, `ErrOverlapTooLarge`, `ErrInvalidMaxTokens`, `ErrEmptyText`, `ErrTokenizerFailed`.
