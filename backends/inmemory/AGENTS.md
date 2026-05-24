# inmemory -- Agent Instructions

## What this package does
Implements in-memory cache backends: `LRUBackend`, `LFUBackend`, `FIFOBackend`. All satisfy `types.Backend[K, V]`.

## Key patterns
- All backends use `sync.RWMutex` for thread safety.
- LRU wraps `hashicorp/golang-lru`.
- LFU and FIFO are hand-rolled.
- All backends store `types.Entry[V]` which holds both the value and embedding.

## Rules
- New backends must implement all 9 methods of `types.Backend[K, V]`.
- Add a compile-time check: `var _ types.Backend[string, string] = (*YourBackend[string, string])(nil)`
- Add test cases in `backend_test.go` using the `factories()` pattern.
- Add benchmarks in `bench_test.go`.

## Testing
```
go test ./backends/inmemory/
go test -bench=. ./backends/inmemory/
```
