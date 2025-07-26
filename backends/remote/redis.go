package remote

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/botirk38/semanticcache/types"
	"github.com/redis/go-redis/v9"
)

// RedisBackend implements CacheBackend using Redis with vector search
type RedisBackend[K comparable, V any] struct {
	client     *redis.Client
	prefix     string
	indexName  string
	dimensions int
}

// redisDocument represents a cached entry stored in Redis
type redisDocument[V any] struct {
	Key       string    `json:"key"`
	Value     V         `json:"value"`
	Embedding []float64 `json:"embedding"`
	Timestamp int64     `json:"timestamp"`
}

// parseRedisURL parses a Redis URL and returns redis.Options
func parseRedisURL(connectionString string) (*redis.Options, error) {
	// Handle redis:// or rediss:// URLs
	if strings.HasPrefix(connectionString, "redis://") || strings.HasPrefix(connectionString, "rediss://") {
		parsedURL, err := url.Parse(connectionString)
		if err != nil {
			return nil, fmt.Errorf("invalid Redis URL: %w", err)
		}

		opts := &redis.Options{
			Addr: parsedURL.Host,
		}

		// Handle TLS
		if parsedURL.Scheme == "rediss" {
			opts.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}

		// Extract username and password
		if parsedURL.User != nil {
			opts.Username = parsedURL.User.Username()
			if password, ok := parsedURL.User.Password(); ok {
				opts.Password = password
			}
		}

		// Extract database number from path
		if parsedURL.Path != "" && parsedURL.Path != "/" {
			dbStr := strings.TrimPrefix(parsedURL.Path, "/")
			if db, err := strconv.Atoi(dbStr); err == nil {
				opts.DB = db
			}
		}

		return opts, nil
	}

	// For simple address format (host:port), return minimal options
	return &redis.Options{
		Addr: connectionString,
	}, nil
}

// NewRedisBackend creates a new Redis backend
func NewRedisBackend[K comparable, V any](config types.BackendConfig) (*RedisBackend[K, V], error) {
	// Parse connection string (supports both URLs and simple addresses)
	opts, err := parseRedisURL(config.ConnectionString)
	if err != nil {
		return nil, err
	}

	// Override with explicit config values if provided
	if config.Username != "" {
		opts.Username = config.Username
	}
	if config.Password != "" {
		opts.Password = config.Password
	}
	if config.Database != 0 {
		opts.DB = config.Database
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	prefix := "semanticcache:"
	if prefixOpt, ok := config.Options["prefix"]; ok {
		if p, ok := prefixOpt.(string); ok {
			prefix = p
		}
	}

	indexName := prefix + "idx"
	if indexOpt, ok := config.Options["index_name"]; ok {
		if idx, ok := indexOpt.(string); ok {
			indexName = idx
		}
	}

	dimensions := 1536 // Default OpenAI embedding dimensions
	if dimOpt, ok := config.Options["dimensions"]; ok {
		if d, ok := dimOpt.(int); ok {
			dimensions = d
		}
	}

	backend := &RedisBackend[K, V]{
		client:     client,
		prefix:     prefix,
		indexName:  indexName,
		dimensions: dimensions,
	}

	// Initialize vector search index
	backend.initializeIndex()

	return backend, nil
}

// initializeIndex creates the Redis vector search index if it doesn't exist
func (b *RedisBackend[K, V]) initializeIndex() {
	ctx := context.Background()

	// Drop existing index if it exists (ignore errors)
	b.client.FTDropIndex(ctx, b.indexName)

	// Create new index with vector field
	_, err := b.client.FTCreate(ctx, b.indexName, &redis.FTCreateOptions{
		OnJSON: true,
		Prefix: []any{b.prefix},
	},
		&redis.FieldSchema{
			FieldName: "$.key",
			As:        "key",
			FieldType: redis.SearchFieldTypeText,
		},
		&redis.FieldSchema{
			FieldName: "$.timestamp",
			As:        "timestamp",
			FieldType: redis.SearchFieldTypeNumeric,
		},
		&redis.FieldSchema{
			FieldName: "$.embedding",
			As:        "embedding",
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				HNSWOptions: &redis.FTHNSWOptions{
					Type:           "FLOAT64",
					Dim:            b.dimensions,
					DistanceMetric: "COSINE",
				},
			},
		},
	).Result()
	if err != nil {
		// Index might already exist, which is fine - ignore error
		_ = err
	}
}

// keyString converts a key to a Redis key string
func (b *RedisBackend[K, V]) keyString(key K) string {
	return fmt.Sprintf("%s%v", b.prefix, key)
}

// floatsToBytes converts a float64 slice to bytes for Redis storage
func floatsToBytes(fs []float64) []byte {
	buf := make([]byte, len(fs)*8)
	for i, f := range fs {
		binary.LittleEndian.PutUint64(buf[i*8:(i+1)*8], math.Float64bits(f))
	}
	return buf
}

// float32ToFloat64 converts float32 slice to float64 slice
func float32ToFloat64(fs []float32) []float64 {
	result := make([]float64, len(fs))
	for i, f := range fs {
		result[i] = float64(f)
	}
	return result
}

// Set stores an entry in Redis using JSON.SET
func (b *RedisBackend[K, V]) Set(ctx context.Context, key K, entry types.Entry[V]) error {
	redisKey := b.keyString(key)

	doc := redisDocument[V]{
		Key:       fmt.Sprintf("%v", key),
		Value:     entry.Value,
		Embedding: float32ToFloat64(entry.Embedding),
		Timestamp: time.Now().Unix(),
	}

	// Store as JSON document
	_, err := b.client.JSONSet(ctx, redisKey, "$", doc).Result()
	if err != nil {
		return fmt.Errorf("failed to set entry in Redis: %w", err)
	}

	return nil
}

// Get retrieves an entry from Redis using JSON.GET
func (b *RedisBackend[K, V]) Get(ctx context.Context, key K) (types.Entry[V], bool, error) {
	redisKey := b.keyString(key)

	result, err := b.client.JSONGet(ctx, redisKey, "$").Result()
	if err == redis.Nil {
		return types.Entry[V]{}, false, nil
	}
	if err != nil {
		return types.Entry[V]{}, false, fmt.Errorf("failed to get entry from Redis: %w", err)
	}

	var docs []redisDocument[V]
	if err := json.Unmarshal([]byte(result), &docs); err != nil {
		return types.Entry[V]{}, false, fmt.Errorf("failed to unmarshal entry: %w", err)
	}

	if len(docs) == 0 {
		return types.Entry[V]{}, false, nil
	}

	doc := docs[0]

	// Convert float64 back to float32
	embedding := make([]float32, len(doc.Embedding))
	for i, f := range doc.Embedding {
		embedding[i] = float32(f)
	}

	entry := types.Entry[V]{
		Embedding: embedding,
		Value:     doc.Value,
	}

	return entry, true, nil
}

// Delete removes an entry from Redis
func (b *RedisBackend[K, V]) Delete(ctx context.Context, key K) error {
	redisKey := b.keyString(key)

	err := b.client.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete entry from Redis: %w", err)
	}

	return nil
}

// Contains checks if a key exists in Redis
func (b *RedisBackend[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	redisKey := b.keyString(key)

	exists, err := b.client.Exists(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence in Redis: %w", err)
	}

	return exists > 0, nil
}

// Flush clears all entries with the configured prefix from Redis
func (b *RedisBackend[K, V]) Flush(ctx context.Context) error {
	pattern := b.prefix + "*"
	var keys []string
	var cursor uint64

	// Use SCAN to get all keys with our prefix
	for {
		result, nextCursor, err := b.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys from Redis: %w", err)
		}

		keys = append(keys, result...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		if err := b.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to flush Redis: %w", err)
		}
	}

	return nil
}

// Len returns the number of entries in Redis with our prefix
func (b *RedisBackend[K, V]) Len(ctx context.Context) (int, error) {
	pattern := b.prefix + "*"
	var count int
	var cursor uint64

	for {
		result, nextCursor, err := b.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to count keys in Redis: %w", err)
		}

		count += len(result)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

// Keys returns all keys in Redis with our prefix using SCAN
func (b *RedisBackend[K, V]) Keys(ctx context.Context) ([]K, error) {
	pattern := b.prefix + "*"
	var redisKeys []string
	var cursor uint64

	for {
		result, nextCursor, err := b.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get keys from Redis: %w", err)
		}

		redisKeys = append(redisKeys, result...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	keys := make([]K, 0, len(redisKeys))
	prefixLen := len(b.prefix)

	for _, redisKey := range redisKeys {
		if len(redisKey) > prefixLen {
			keyStr := redisKey[prefixLen:]
			var key K
			// Try to convert string back to key type
			if err := json.Unmarshal(fmt.Appendf(nil, "\"%s\"", keyStr), &key); err == nil {
				keys = append(keys, key)
			}
		}
	}

	return keys, nil
}

// GetEmbedding retrieves just the embedding for a key using JSON.GET
func (b *RedisBackend[K, V]) GetEmbedding(ctx context.Context, key K) ([]float32, bool, error) {
	redisKey := b.keyString(key)

	result, err := b.client.JSONGet(ctx, redisKey, "$.embedding").Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get embedding from Redis: %w", err)
	}

	var embeddings [][]float64
	if err := json.Unmarshal([]byte(result), &embeddings); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, false, nil
	}

	// Convert float64 back to float32
	embedding := make([]float32, len(embeddings[0]))
	for i, f := range embeddings[0] {
		embedding[i] = float32(f)
	}

	return embedding, true, nil
}

// VectorSearch performs vector similarity search using Redis FT.SEARCH
func (b *RedisBackend[K, V]) VectorSearch(ctx context.Context, queryEmbedding []float32, threshold float32, limit int) ([]K, error) {
	// Convert embedding to bytes for search
	embedding64 := float32ToFloat64(queryEmbedding)
	embeddingBytes := floatsToBytes(embedding64)

	// Perform vector search
	query := fmt.Sprintf("*=>[KNN %d @embedding $vec AS vector_distance]", limit)

	results, err := b.client.FTSearchWithArgs(ctx, b.indexName, query, &redis.FTSearchOptions{
		Return: []redis.FTSearchReturn{
			{FieldName: "vector_distance"},
			{FieldName: "key"},
		},
		DialectVersion: 2,
		Params: map[string]any{
			"vec": embeddingBytes,
		},
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("vector search error: %w", err)
	}

	var keys []K
	for _, doc := range results.Docs {
		// Get the vector distance (lower is better for cosine similarity)
		distanceStr, ok := doc.Fields["vector_distance"]
		if !ok {
			continue
		}

		distance, err := strconv.ParseFloat(distanceStr, 32)
		if err != nil {
			continue
		}

		// Convert distance to similarity (1 - distance for cosine)
		similarity := 1.0 - distance

		// Check if similarity meets threshold
		if float32(similarity) >= threshold {
			keyStr, ok := doc.Fields["key"]
			if !ok {
				continue
			}

			var key K
			if err := json.Unmarshal(fmt.Appendf(nil, "\"%s\"", keyStr), &key); err == nil {
				keys = append(keys, key)
			}
		}
	}

	return keys, nil
}

// Close closes the Redis connection
func (b *RedisBackend[K, V]) Close() error {
	return b.client.Close()
}
