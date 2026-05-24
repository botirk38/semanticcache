package inmemory

import (
	"context"
	"fmt"
	"testing"

	"github.com/botirk38/semanticcache/types"
)

func benchEntry(dim int) types.Entry[string] {
	emb := make([]float64, dim)
	for i := range emb {
		emb[i] = float64(i) / float64(dim)
	}
	return types.Entry[string]{Embedding: emb, Value: "val"}
}

// --- LRU benchmarks ---

func BenchmarkLRU_Set(b *testing.B) {
	backend, _ := NewLRUBackend[string, string](types.BackendConfig{Capacity: b.N + 1})
	ctx := context.Background()
	e := benchEntry(1536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), e)
	}
}

func BenchmarkLRU_Get(b *testing.B) {
	backend, _ := NewLRUBackend[string, string](types.BackendConfig{Capacity: 10000})
	ctx := context.Background()
	e := benchEntry(1536)
	for i := 0; i < 10000; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.Get(ctx, fmt.Sprintf("k%d", i%10000))
	}
}

func BenchmarkLRU_GetEmbedding(b *testing.B) {
	backend, _ := NewLRUBackend[string, string](types.BackendConfig{Capacity: 10000})
	ctx := context.Background()
	e := benchEntry(1536)
	for i := 0; i < 10000; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.GetEmbedding(ctx, fmt.Sprintf("k%d", i%10000))
	}
}

// --- LFU benchmarks ---

func BenchmarkLFU_Set(b *testing.B) {
	backend, _ := NewLFUBackend[string, string](types.BackendConfig{Capacity: b.N + 1})
	ctx := context.Background()
	e := benchEntry(1536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), e)
	}
}

func BenchmarkLFU_Get(b *testing.B) {
	backend, _ := NewLFUBackend[string, string](types.BackendConfig{Capacity: 10000})
	ctx := context.Background()
	e := benchEntry(1536)
	for i := 0; i < 10000; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.Get(ctx, fmt.Sprintf("k%d", i%10000))
	}
}

// --- FIFO benchmarks ---

func BenchmarkFIFO_Set(b *testing.B) {
	backend, _ := NewFIFOBackend[string, string](types.BackendConfig{Capacity: b.N + 1})
	ctx := context.Background()
	e := benchEntry(1536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), e)
	}
}

func BenchmarkFIFO_Get(b *testing.B) {
	backend, _ := NewFIFOBackend[string, string](types.BackendConfig{Capacity: 10000})
	ctx := context.Background()
	e := benchEntry(1536)
	for i := 0; i < 10000; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), e)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.Get(ctx, fmt.Sprintf("k%d", i%10000))
	}
}
