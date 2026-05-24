# backends

Convenience re-exports for backend constructors.

This package provides top-level constructor functions that delegate to the concrete implementations in `inmemory/` and `remote/`. You can import backends directly from their subpackages if you prefer.

## Subpackages

- `inmemory/` -- in-memory backends (LRU, LFU, FIFO)
- `remote/` -- remote backends (Redis)
