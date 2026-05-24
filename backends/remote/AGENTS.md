# remote -- Agent Instructions

## What this package does
Implements the Redis backend (`RedisBackend[K, V]`). Stores entries as JSON documents using Redis JSONSet/JSONGet.

## Key patterns
- Connection string parsing supports `host:port`, `redis://`, and `rediss://` URLs.
- Configuration via `RedisOption` functional options.
- Key format: `{prefix}{key}` (default prefix: `semanticcache:`).
- `Keys()` uses SCAN to iterate without blocking.
- Constructor pings Redis to verify connectivity.

## Rules
- Requires RedisJSON module or Redis 7.2+.
- Do not add vector search logic here -- the cache layer handles similarity search.
- Tests for Redis require a running Redis instance, so they are not run in CI by default.

## Testing
Requires a local Redis: `go test ./backends/remote/ -v`
