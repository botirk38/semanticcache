# remote

Remote cache backends. Currently provides a Redis backend.

## RedisBackend

Stores entries as JSON documents in Redis using `JSONSet`/`JSONGet` (requires RedisJSON module or Redis 7.2+).

```go
b, err := remote.NewRedisBackend[string, string]("localhost:6379",
    remote.WithPassword("secret"),
    remote.WithDB(2),
    remote.WithPrefix("myapp:"),
)
```

### Connection

Accepts `host:port`, `redis://`, or `rediss://` (TLS) URLs. The constructor pings Redis to verify connectivity.

### Options

| Option | Description |
|--------|-------------|
| `WithUsername(u)` | Redis ACL username |
| `WithPassword(p)` | Redis password |
| `WithDB(n)` | Database number (default 0) |
| `WithPrefix(p)` | Key prefix (default `semanticcache:`) |
| `WithTLS(cfg)` | Custom TLS configuration |

### Key layout

Each entry is stored as a JSON document at `{prefix}{key}` with fields: `key`, `value`, `embedding`.
