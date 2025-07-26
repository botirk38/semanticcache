# Redis Backend Setup Guide

This guide shows how to set up SemanticCache with Redis for persistent, scalable semantic caching across multiple applications.

## Prerequisites

- Go 1.18+
- Redis server (local or remote)
- OpenAI API key

## Redis Installation

### Local Redis (for development)

**Docker (recommended):**
```bash
# Start Redis with Docker
docker run -d --name redis-cache -p 6379:6379 redis:7-alpine

# Or with persistence
docker run -d --name redis-cache \
  -p 6379:6379 \
  -v redis-data:/data \
  redis:7-alpine redis-server --appendonly yes
```

**macOS:**
```bash
brew install redis
brew services start redis
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis-server
```

### Cloud Redis

For production, consider managed Redis services:
- **AWS ElastiCache**
- **Google Cloud Memorystore** 
- **Azure Cache for Redis**
- **Redis Cloud**

## Setup Code

The Redis backend now supports two configuration methods:

### Option 1: Simple Address Format

```go
config := types.BackendConfig{
    ConnectionString: "localhost:6379",
    Database:         0,
    Password:         "optional-password",
}
```

### Option 2: Full Redis URL

```go
// Basic URL
config := types.BackendConfig{
    ConnectionString: "redis://localhost:6379",
}

// With authentication and database
config := types.BackendConfig{
    ConnectionString: "redis://username:password@localhost:6379/5",
}

// With TLS (rediss://)
config := types.BackendConfig{
    ConnectionString: "rediss://username:password@redis.example.com:6380/2",
}
```

**Override behavior:** Individual config fields (`Username`, `Password`, `Database`) will override URL values if both are provided.

### Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/botirk38/semanticcache"
    "github.com/botirk38/semanticcache/backends"
    "github.com/botirk38/semanticcache/providers/openai"
    "github.com/botirk38/semanticcache/types"
)

func main() {
    // 1. Configure Redis backend (choose one format)
    config := types.BackendConfig{
        // Option A: Simple format
        ConnectionString: "localhost:6379",
        Password:         "your-redis-password",
        Database:         0,
        
        // Option B: Full URL (comment out Option A)
        // ConnectionString: "redis://username:password@localhost:6379/0",
    }

    // 2. Create backend factory
    factory := &backends.BackendFactory[string, string]{}
    backend, err := factory.NewBackend(types.BackendRedis, config)
    if err != nil {
        log.Fatal("Failed to create Redis backend:", err)
    }
    defer backend.Close()

    // 3. Create OpenAI provider
    provider, err := openai.NewOpenAIProvider(openai.OpenAIConfig{
        APIKey: "your-openai-api-key",
        Model:  "text-embedding-3-small",
    })
    if err != nil {
        log.Fatal("Failed to create OpenAI provider:", err)
    }
    defer provider.Close()

    // 4. Create semantic cache
    cache, err := semanticcache.NewSemanticCache(backend, provider, nil)
    if err != nil {
        log.Fatal("Failed to create cache:", err)
    }
    defer cache.Close()

    // 5. Use the cache
    ctx := context.Background()
    
    // Store some data
    cache.Set("q1", "What is the capital of France?", "Paris")
    cache.Set("q2", "What is the capital of Germany?", "Berlin")
    
    // Semantic lookup
    value, found, err := cache.Lookup("French capital city?", 0.8)
    if err != nil {
        log.Fatal("Lookup failed:", err)
    }
    if found {
        fmt.Printf("Found: %s\n", value) // "Paris"
    }
}
```

### More Configuration Examples

**Redis with Authentication (Simple Format):**
```go
config := types.BackendConfig{
    ConnectionString: "localhost:6379",
    Username:         "your-username",    // Redis 6+ ACL
    Password:         "your-password",
    Database:         0,
}
```

**Redis with Authentication (URL Format):**
```go
config := types.BackendConfig{
    ConnectionString: "redis://username:password@localhost:6379/0",
}
```

**Redis with TLS (URL Format):**
```go
config := types.BackendConfig{
    ConnectionString: "rediss://username:password@redis.example.com:6380/0",
}
```

**Redis Cluster:**
```go
config := types.BackendConfig{
    ConnectionString: "redis-cluster-node1:6379,redis-cluster-node2:6379,redis-cluster-node3:6379",
    Password:         "your-password",
}
```

**Override Example (URL + Individual Fields):**
```go
config := types.BackendConfig{
    ConnectionString: "redis://:urlpassword@localhost:6379/1",
    Username:         "override-user",     // Overrides URL username
    Password:         "override-pass",     // Overrides URL password  
    Database:         5,                   // Overrides URL database
}
```

## Configuration Options

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `ConnectionString` | `string` | Redis server address or full URL | Required |
| `Database` | `int` | Redis database number (overrides URL) | 0 |
| `Username` | `string` | Redis username (overrides URL) | "" |
| `Password` | `string` | Redis password (overrides URL) | "" |

### Supported URL Formats

- `redis://localhost:6379` - Basic connection
- `redis://username:password@localhost:6379` - With authentication  
- `redis://username:password@localhost:6379/5` - With database selection
- `rediss://username:password@host:6380/0` - With TLS encryption
- `localhost:6379` - Simple address format (legacy)

## Environment Variables

For security, use environment variables:

**Simple Format:**
```go
import "os"

config := types.BackendConfig{
    ConnectionString: os.Getenv("REDIS_HOST"),         // "localhost:6379"
    Password:         os.Getenv("REDIS_PASSWORD"),
    Database:         0,
}
```

**URL Format:**
```go
import "os"

config := types.BackendConfig{
    ConnectionString: os.Getenv("REDIS_URL"),          // "redis://user:pass@host:6379/0"
}
```

**.env file examples:**
```bash
# Simple format
REDIS_HOST=localhost:6379
REDIS_PASSWORD=your-secret-password

# URL format  
REDIS_URL=redis://username:password@localhost:6379/0

# OpenAI
OPENAI_API_KEY=your-openai-key
```

## Testing Redis Connection

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/botirk38/semanticcache/backends"
    "github.com/botirk38/semanticcache/types"
)

func testRedisConnection() {
    config := types.BackendConfig{
        ConnectionString: "localhost:6379",
        Database:         0,
    }

    factory := &backends.BackendFactory[string, string]{}
    backend, err := factory.NewBackend(types.BackendRedis, config)
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
    defer backend.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Test basic operations
    testEntry := types.Entry[string]{
        Value:     "test-value",
        Embedding: []float32{0.1, 0.2, 0.3},
    }

    // Set
    err = backend.Set(ctx, "test-key", testEntry)
    if err != nil {
        log.Fatal("Failed to set value:", err)
    }

    // Get
    retrieved, found, err := backend.Get(ctx, "test-key")
    if err != nil {
        log.Fatal("Failed to get value:", err)
    }
    if !found {
        log.Fatal("Value not found")
    }

    fmt.Printf("✅ Redis connection successful!\n")
    fmt.Printf("Retrieved value: %s\n", retrieved.Value)

    // Cleanup
    backend.Delete(ctx, "test-key")
}
```

## Performance Considerations

### Connection Pooling

Redis backend automatically handles connection pooling. For high-throughput applications:

```go
// Redis handles connection pooling internally
// Default pool size is sufficient for most use cases
```

### Memory Usage

Monitor Redis memory usage:

```bash
# Check Redis memory info
redis-cli info memory

# Set maxmemory policy (recommended for cache)
redis-cli config set maxmemory-policy allkeys-lru
```

### Persistence

Choose Redis persistence based on your needs:

- **RDB**: Periodic snapshots (default)
- **AOF**: Log every write operation
- **Both**: Maximum durability
- **None**: Pure cache (fastest)

## Troubleshooting

### Connection Issues

```bash
# Test Redis connectivity
redis-cli ping
# Expected: PONG

# Check if Redis is running
redis-cli info server
```

### Common Errors

**Error: "connection refused"**
```bash
# Check if Redis is running
sudo systemctl status redis-server
# or
brew services list | grep redis
```

**Error: "NOAUTH Authentication required"**
```go
// Add password to config
config.Password = "your-redis-password"
```

**Error: "WRONGTYPE Operation against a key holding the wrong kind of value"**
```bash
# Clear conflicting keys
redis-cli flushdb
```

## Production Deployment

### Security Checklist

- [ ] Use authentication (`requirepass` or ACL)
- [ ] Enable TLS encryption
- [ ] Configure firewall rules
- [ ] Use environment variables for secrets
- [ ] Set up monitoring and alerts

### Monitoring

Monitor these Redis metrics:
- Memory usage
- Connection count
- Hit/miss ratio
- Network I/O
- CPU usage

### Backup Strategy

```bash
# Manual backup
redis-cli bgsave

# Automated backup script
#!/bin/bash
redis-cli bgsave
cp /var/lib/redis/dump.rdb /backup/redis-$(date +%Y%m%d-%H%M%S).rdb
```

## Example: Multi-Application Cache

Share semantic cache across multiple applications:

**App 1 (Writer):**
```go
// Store FAQ data
cache.Set("faq1", "How do I reset my password?", "Go to Settings > Security > Reset Password")
cache.Set("faq2", "How do I contact support?", "Email us at support@company.com")
```

**App 2 (Reader):**
```go
// Query from different app
value, found, _ := cache.Lookup("password reset help", 0.8)
if found {
    fmt.Println("Answer:", value) // "Go to Settings > Security > Reset Password"
}
```

This enables semantic search across your entire application ecosystem!