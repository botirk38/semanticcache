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

// RedisOption configures a RedisBackend.
type RedisOption func(*redisConfig)

type redisConfig struct {
	username   string
	password   string
	db         int
	prefix     string
	indexName  string
	dimensions int
	tlsConfig  *tls.Config
}

// WithUsername sets the Redis username.
func WithUsername(username string) RedisOption {
	return func(c *redisConfig) { c.username = username }
}

// WithPassword sets the Redis password.
func WithPassword(password string) RedisOption {
	return func(c *redisConfig) { c.password = password }
}

// WithDB sets the Redis database number.
func WithDB(db int) RedisOption {
	return func(c *redisConfig) { c.db = db }
}

// WithPrefix sets the key prefix for all entries.
func WithPrefix(prefix string) RedisOption {
	return func(c *redisConfig) { c.prefix = prefix }
}

// WithIndexName sets the Redis search index name.
func WithIndexName(name string) RedisOption {
	return func(c *redisConfig) { c.indexName = name }
}

// WithDimensions sets the embedding vector dimensions.
func WithDimensions(dim int) RedisOption {
	return func(c *redisConfig) { c.dimensions = dim }
}

// WithTLS sets the TLS configuration for the Redis connection.
func WithTLS(cfg *tls.Config) RedisOption {
	return func(c *redisConfig) { c.tlsConfig = cfg }
}

// RedisBackend implements Backend and VectorSearcher using Redis.
type RedisBackend[K comparable, V any] struct {
	client     *redis.Client
	prefix     string
	indexName  string
	dimensions int
}

type redisDocument[V any] struct {
	Key       string    `json:"key"`
	Value     V         `json:"value"`
	Embedding []float64 `json:"embedding"`
	Timestamp int64     `json:"timestamp"`
}

func parseRedisURL(connectionString string) (*redis.Options, error) {
	if strings.HasPrefix(connectionString, "redis://") || strings.HasPrefix(connectionString, "rediss://") {
		parsedURL, err := url.Parse(connectionString)
		if err != nil {
			return nil, fmt.Errorf("invalid Redis URL: %w", err)
		}

		opts := &redis.Options{Addr: parsedURL.Host}

		if parsedURL.Scheme == "rediss" {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}

		if parsedURL.User != nil {
			opts.Username = parsedURL.User.Username()
			if pw, ok := parsedURL.User.Password(); ok {
				opts.Password = pw
			}
		}

		if parsedURL.Path != "" && parsedURL.Path != "/" {
			dbStr := strings.TrimPrefix(parsedURL.Path, "/")
			if db, err := strconv.Atoi(dbStr); err == nil {
				opts.DB = db
			}
		}
		return opts, nil
	}
	return &redis.Options{Addr: connectionString}, nil
}

// NewRedisBackend creates a new Redis backend. addr can be "host:port" or a
// redis:// / rediss:// URL.
func NewRedisBackend[K comparable, V any](addr string, opts ...RedisOption) (*RedisBackend[K, V], error) {
	cfg := &redisConfig{
		prefix:     "semanticcache:",
		dimensions: 1536,
	}
	for _, o := range opts {
		o(cfg)
	}
	if cfg.indexName == "" {
		cfg.indexName = cfg.prefix + "idx"
	}

	redisOpts, err := parseRedisURL(addr)
	if err != nil {
		return nil, err
	}
	if cfg.username != "" {
		redisOpts.Username = cfg.username
	}
	if cfg.password != "" {
		redisOpts.Password = cfg.password
	}
	if cfg.db != 0 {
		redisOpts.DB = cfg.db
	}
	if cfg.tlsConfig != nil {
		redisOpts.TLSConfig = cfg.tlsConfig
	}

	client := redis.NewClient(redisOpts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	b := &RedisBackend[K, V]{
		client:     client,
		prefix:     cfg.prefix,
		indexName:  cfg.indexName,
		dimensions: cfg.dimensions,
	}
	b.initializeIndex()
	return b, nil
}

func (b *RedisBackend[K, V]) initializeIndex() {
	ctx := context.Background()

	// Drop existing index (ignore errors).
	b.client.FTDropIndex(ctx, b.indexName)

	_, _ = b.client.FTCreate(ctx, b.indexName, &redis.FTCreateOptions{
		OnJSON: true,
		Prefix: []any{b.prefix},
	},
		&redis.FieldSchema{FieldName: "$.key", As: "key", FieldType: redis.SearchFieldTypeText},
		&redis.FieldSchema{FieldName: "$.timestamp", As: "timestamp", FieldType: redis.SearchFieldTypeNumeric},
		&redis.FieldSchema{
			FieldName: "$.embedding", As: "embedding",
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				HNSWOptions: &redis.FTHNSWOptions{
					Type: "FLOAT64", Dim: b.dimensions, DistanceMetric: "COSINE",
				},
			},
		},
	).Result()
}

func (b *RedisBackend[K, V]) keyString(key K) string {
	return fmt.Sprintf("%s%v", b.prefix, key)
}

func floatsToBytes(fs []float64) []byte {
	buf := make([]byte, len(fs)*8)
	for i, f := range fs {
		binary.LittleEndian.PutUint64(buf[i*8:(i+1)*8], math.Float64bits(f))
	}
	return buf
}

// Set stores a value with its embedding in Redis.
func (b *RedisBackend[K, V]) Set(ctx context.Context, key K, embedding []float64, value V) error {
	doc := redisDocument[V]{
		Key:       fmt.Sprintf("%v", key),
		Value:     value,
		Embedding: embedding,
		Timestamp: time.Now().Unix(),
	}
	_, err := b.client.JSONSet(ctx, b.keyString(key), "$", doc).Result()
	if err != nil {
		return fmt.Errorf("failed to set entry in Redis: %w", err)
	}
	return nil
}

// Get retrieves the value for a key.
func (b *RedisBackend[K, V]) Get(ctx context.Context, key K) (V, bool, error) {
	result, err := b.client.JSONGet(ctx, b.keyString(key), "$").Result()
	if err == redis.Nil {
		var zero V
		return zero, false, nil
	}
	if err != nil {
		var zero V
		return zero, false, fmt.Errorf("failed to get entry from Redis: %w", err)
	}

	var docs []redisDocument[V]
	if err := json.Unmarshal([]byte(result), &docs); err != nil {
		var zero V
		return zero, false, fmt.Errorf("failed to unmarshal entry: %w", err)
	}
	if len(docs) == 0 {
		var zero V
		return zero, false, nil
	}
	return docs[0].Value, true, nil
}

// Delete removes an entry by key.
func (b *RedisBackend[K, V]) Delete(ctx context.Context, key K) error {
	if err := b.client.Del(ctx, b.keyString(key)).Err(); err != nil {
		return fmt.Errorf("failed to delete entry from Redis: %w", err)
	}
	return nil
}

// Contains checks whether a key exists.
func (b *RedisBackend[K, V]) Contains(ctx context.Context, key K) (bool, error) {
	n, err := b.client.Exists(ctx, b.keyString(key)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence in Redis: %w", err)
	}
	return n > 0, nil
}

// Flush removes all entries with the configured prefix.
func (b *RedisBackend[K, V]) Flush(ctx context.Context) error {
	var cursor uint64
	for {
		result, next, err := b.client.Scan(ctx, cursor, b.prefix+"*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys from Redis: %w", err)
		}
		if len(result) > 0 {
			if err := b.client.Del(ctx, result...).Err(); err != nil {
				return fmt.Errorf("failed to flush Redis: %w", err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

// Len returns the number of entries with the configured prefix.
func (b *RedisBackend[K, V]) Len(ctx context.Context) (int, error) {
	var count int
	var cursor uint64
	for {
		result, next, err := b.client.Scan(ctx, cursor, b.prefix+"*", 100).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to count keys in Redis: %w", err)
		}
		count += len(result)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return count, nil
}

// Keys returns all keys stored under the configured prefix.
func (b *RedisBackend[K, V]) Keys(ctx context.Context) ([]K, error) {
	var keys []K
	var cursor uint64
	for {
		result, next, err := b.client.Scan(ctx, cursor, b.prefix+"*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys from Redis: %w", err)
		}
		for _, rk := range result {
			raw := strings.TrimPrefix(rk, b.prefix)
			var key K
			if err := json.Unmarshal(fmt.Appendf(nil, "\"%s\"", raw), &key); err == nil {
				keys = append(keys, key)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// GetEmbedding retrieves the embedding vector for a key.
func (b *RedisBackend[K, V]) GetEmbedding(ctx context.Context, key K) ([]float64, bool, error) {
	result, err := b.client.JSONGet(ctx, b.keyString(key), "$").Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get embedding from Redis: %w", err)
	}
	var docs []redisDocument[V]
	if err := json.Unmarshal([]byte(result), &docs); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}
	if len(docs) == 0 {
		return nil, false, nil
	}
	return docs[0].Embedding, true, nil
}

// Close closes the Redis connection.
func (b *RedisBackend[K, V]) Close() error {
	return b.client.Close()
}

// VectorSearch performs a server-side similarity search using Redis FT.SEARCH.
func (b *RedisBackend[K, V]) VectorSearch(ctx context.Context, queryEmbedding []float64, threshold float64, limit int) ([]types.VectorSearchResult[K, V], error) {
	embeddingBytes := floatsToBytes(queryEmbedding)

	query := fmt.Sprintf("*=>[KNN %d @embedding $vec AS vector_distance]", limit)
	results, err := b.client.FTSearchWithArgs(ctx, b.indexName, query, &redis.FTSearchOptions{
		Return: []redis.FTSearchReturn{
			{FieldName: "vector_distance"},
			{FieldName: "key"},
		},
		DialectVersion: 2,
		Params:         map[string]any{"vec": embeddingBytes},
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("vector search error: %w", err)
	}

	var hits []types.VectorSearchResult[K, V]
	for _, doc := range results.Docs {
		distStr, ok := doc.Fields["vector_distance"]
		if !ok {
			continue
		}
		distance, err := strconv.ParseFloat(distStr, 64)
		if err != nil {
			continue
		}
		score := 1.0 - distance
		if score < threshold {
			continue
		}

		keyStr, ok := doc.Fields["key"]
		if !ok {
			continue
		}
		var key K
		if err := json.Unmarshal(fmt.Appendf(nil, "\"%s\"", keyStr), &key); err != nil {
			continue
		}

		value, found, err := b.Get(ctx, key)
		if err != nil || !found {
			continue
		}

		hits = append(hits, types.VectorSearchResult[K, V]{
			Key:   key,
			Value: value,
			Score: score,
		})
	}
	return hits, nil
}
