# types -- Agent Instructions

## What this package does
Defines the core interfaces (`Backend[K, V]`, `EmbeddingProvider`, `BatchEmbeddingProvider`) and the `Entry[V]` type. No implementation code lives here.

## Rules
- Do not add implementation code to this package.
- Any change to `Backend` or `EmbeddingProvider` requires updating all implementations (inmemory/*, remote/redis, providers/*).
- `Backend` has 9 methods. Do not add methods without updating every backend.
- Keep interfaces minimal. Prefer optional interfaces (like `BatchEmbeddingProvider`) over bloating the core interface.

## Testing
No tests in this package -- it only contains interface definitions. Run downstream tests after changes: `go test ./...`
