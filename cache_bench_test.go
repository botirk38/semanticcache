package semanticcache

import (
	"context"
	"fmt"
	"testing"

	"github.com/botirk38/semanticcache/providers/local"
	"github.com/botirk38/semanticcache/similarity"
)

func benchCache(b *testing.B) *Cache[string, string] {
	b.Helper()
	prov := local.New(128)
	backend := newMockBackend[string, string]()
	c, err := NewSemanticCache[string, string](backend, prov, similarity.CosineSimilarity)
	if err != nil {
		b.Fatal(err)
	}
	return c
}

func BenchmarkCache_Set(b *testing.B) {
	c := benchCache(b)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Set(ctx, fmt.Sprintf("k%d", i%1000), fmt.Sprintf("text %d", i), "val")
	}
}

func BenchmarkCache_Get(b *testing.B) {
	c := benchCache(b)
	ctx := context.Background()
	for i := 0; i < 500; i++ {
		_ = c.Set(ctx, fmt.Sprintf("k%d", i), fmt.Sprintf("text %d", i), "val")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = c.Get(ctx, fmt.Sprintf("k%d", i%500))
	}
}

func BenchmarkCache_Lookup(b *testing.B) {
	c := benchCache(b)
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		_ = c.Set(ctx, fmt.Sprintf("k%d", i), fmt.Sprintf("text %d", i), "val")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Lookup(ctx, "query", 0.5)
	}
}

func BenchmarkCache_TopMatches(b *testing.B) {
	c := benchCache(b)
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		_ = c.Set(ctx, fmt.Sprintf("k%d", i), fmt.Sprintf("text %d", i), "val")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.TopMatches(ctx, "query", 5)
	}
}
