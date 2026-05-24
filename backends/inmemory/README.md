# inmemory

In-memory cache backends with different eviction strategies. All implement `types.Backend[K, V]`.

## Backends

### LRUBackend

Least Recently Used eviction. Wraps `hashicorp/golang-lru`.

```go
b, err := inmemory.NewLRUBackend[string, string](1000)
```

### LFUBackend

Least Frequently Used eviction. Entries accessed less often are evicted first.

```go
b, err := inmemory.NewLFUBackend[string, string](1000)
```

### FIFOBackend

First In, First Out eviction. Oldest entries are evicted first regardless of access pattern.

```go
b, err := inmemory.NewFIFOBackend[string, string](1000)
```

## Thread safety

All backends are safe for concurrent use. They use `sync.RWMutex` internally.

## Choosing a backend

- LRU: good default for most workloads with temporal locality
- LFU: when some entries are accessed much more often than others
- FIFO: simplest eviction, useful for streaming/queue patterns
