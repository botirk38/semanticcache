package semanticcache

import (
	"context"
	"fmt"
	"testing"

	"github.com/botirk38/semanticcache/options"
	"github.com/botirk38/semanticcache/similarity"
)

// benchProvider returns a deterministic embedding from text length.
type benchProvider struct{}

func (benchProvider) EmbedText(text string) ([]float64, error) {
	dim := 128
	emb := make([]float64, dim)
	seed := float64(len(text))
	for i := range emb {
		emb[i] = seed / float64(i+1)
	}
	return emb, nil
}
func (benchProvider) EmbedBatch(texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))
	for i, t := range texts {
		e, _ := benchProvider{}.EmbedText(t)
		results[i] = e
	}
	return results, nil
}
func (benchProvider) Close()             {}
func (benchProvider) GetMaxTokens() int  { return 8192 }
func (benchProvider) GetDimensions() int { return 128 }

func newBenchCache(b *testing.B, capacity int) *SemanticCache[string, string] {
	b.Helper()
	cache, err := New[string, string](
		options.WithLRUBackend[string, string](capacity),
		options.WithCustomProvider[string, string](benchProvider{}),
		options.WithSimilarityComparator[string, string](similarity.CosineSimilarity),
		options.WithChunking[string, string](false),
	)
	if err != nil {
		b.Fatal(err)
	}
	return cache
}

func BenchmarkCache_Set(b *testing.B) {
	cache := newBenchCache(b, b.N+1)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Set(ctx, fmt.Sprintf("key-%d", i), fmt.Sprintf("text-%d", i), fmt.Sprintf("val-%d", i))
	}
}

func BenchmarkCache_Get(b *testing.B) {
	cache := newBenchCache(b, 10000)
	ctx := context.Background()
	for i := 0; i < 10000; i++ {
		_ = cache.Set(ctx, fmt.Sprintf("k%d", i), fmt.Sprintf("text-%d", i), "v")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, fmt.Sprintf("k%d", i%10000))
	}
}

func BenchmarkCache_Lookup(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			cache := newBenchCache(b, size)
			ctx := context.Background()
			for i := 0; i < size; i++ {
				_ = cache.Set(ctx, fmt.Sprintf("k%d", i), fmt.Sprintf("text-%d", i), "v")
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cache.Lookup(ctx, "some query", 0.5)
			}
		})
	}
}

func BenchmarkCache_TopMatches(b *testing.B) {
	cache := newBenchCache(b, 1000)
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		_ = cache.Set(ctx, fmt.Sprintf("k%d", i), fmt.Sprintf("text-%d", i), "v")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.TopMatches(ctx, "some query", 10)
	}
}

func BenchmarkSimilarity_Lookup_Parallel(b *testing.B) {
	cache := newBenchCache(b, 1000)
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		_ = cache.Set(ctx, fmt.Sprintf("k%d", i), fmt.Sprintf("text-%d", i), "v")
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Lookup(ctx, "parallel query", 0.5)
		}
	})
}
