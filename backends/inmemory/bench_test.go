package inmemory

import (
	"context"
	"fmt"
	"testing"

	"github.com/botirk38/semanticcache/types"
)

func benchSet(b *testing.B, backend types.Backend[string, string]) {
	ctx := context.Background()
	emb := make([]float64, 128)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i%1000), emb, "v")
	}
}

func benchGet(b *testing.B, backend types.Backend[string, string]) {
	ctx := context.Background()
	emb := make([]float64, 128)
	for i := 0; i < 1000; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), emb, "v")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = backend.Get(ctx, fmt.Sprintf("k%d", i%1000))
	}
}

func benchKeys(b *testing.B, backend types.Backend[string, string]) {
	ctx := context.Background()
	emb := make([]float64, 128)
	for i := 0; i < 1000; i++ {
		_ = backend.Set(ctx, fmt.Sprintf("k%d", i), emb, "v")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = backend.Keys(ctx)
	}
}

func BenchmarkLRU_Set(b *testing.B) {
	backend, _ := NewLRUBackend[string, string](1000)
	benchSet(b, backend)
}

func BenchmarkLRU_Get(b *testing.B) {
	backend, _ := NewLRUBackend[string, string](1000)
	benchGet(b, backend)
}

func BenchmarkLRU_Keys(b *testing.B) {
	backend, _ := NewLRUBackend[string, string](1000)
	benchKeys(b, backend)
}

func BenchmarkLFU_Set(b *testing.B) {
	backend, _ := NewLFUBackend[string, string](1000)
	benchSet(b, backend)
}

func BenchmarkLFU_Get(b *testing.B) {
	backend, _ := NewLFUBackend[string, string](1000)
	benchGet(b, backend)
}

func BenchmarkLFU_Keys(b *testing.B) {
	backend, _ := NewLFUBackend[string, string](1000)
	benchKeys(b, backend)
}

func BenchmarkFIFO_Set(b *testing.B) {
	backend, _ := NewFIFOBackend[string, string](1000)
	benchSet(b, backend)
}

func BenchmarkFIFO_Get(b *testing.B) {
	backend, _ := NewFIFOBackend[string, string](1000)
	benchGet(b, backend)
}

func BenchmarkFIFO_Keys(b *testing.B) {
	backend, _ := NewFIFOBackend[string, string](1000)
	benchKeys(b, backend)
}
