package inmemory

import (
	"context"
	"testing"

	"github.com/botirk38/semanticcache/types"
)

// backendFactory creates a fresh Backend for testing.
type backendFactory func(t *testing.T) types.Backend[string, string]

func factories() map[string]backendFactory {
	return map[string]backendFactory{
		"LRU": func(t *testing.T) types.Backend[string, string] {
			t.Helper()
			b, err := NewLRUBackend[string, string](100)
			if err != nil {
				t.Fatalf("NewLRUBackend: %v", err)
			}
			return b
		},
		"LFU": func(t *testing.T) types.Backend[string, string] {
			t.Helper()
			b, _ := NewLFUBackend[string, string](100)
			return b
		},
		"FIFO": func(t *testing.T) types.Backend[string, string] {
			t.Helper()
			b, _ := NewFIFOBackend[string, string](100)
			return b
		},
	}
}

func TestBackend_SetGetDelete(t *testing.T) {
	for name, factory := range factories() {
		t.Run(name, func(t *testing.T) {
			b := factory(t)
			ctx := context.Background()

			if err := b.Set(ctx, "k1", []float64{1, 0}, "v1"); err != nil {
				t.Fatalf("Set: %v", err)
			}

			v, ok, err := b.Get(ctx, "k1")
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if !ok || v != "v1" {
				t.Fatalf("expected v1, got %q (ok=%v)", v, ok)
			}

			if err := b.Delete(ctx, "k1"); err != nil {
				t.Fatalf("Delete: %v", err)
			}

			_, ok, err = b.Get(ctx, "k1")
			if err != nil {
				t.Fatalf("Get after delete: %v", err)
			}
			if ok {
				t.Fatal("expected not found after delete")
			}
		})
	}
}

func TestBackend_Contains(t *testing.T) {
	for name, factory := range factories() {
		t.Run(name, func(t *testing.T) {
			b := factory(t)
			ctx := context.Background()

			ok, _ := b.Contains(ctx, "nope")
			if ok {
				t.Fatal("expected not found")
			}

			_ = b.Set(ctx, "k", []float64{1}, "v")
			ok, _ = b.Contains(ctx, "k")
			if !ok {
				t.Fatal("expected found")
			}
		})
	}
}

func TestBackend_KeysAndGetEmbedding(t *testing.T) {
	for name, factory := range factories() {
		t.Run(name, func(t *testing.T) {
			b := factory(t)
			ctx := context.Background()

			_ = b.Set(ctx, "a", []float64{1, 2}, "va")
			_ = b.Set(ctx, "b", []float64{3, 4}, "vb")

			keys, err := b.Keys(ctx)
			if err != nil {
				t.Fatalf("Keys: %v", err)
			}
			if len(keys) != 2 {
				t.Fatalf("expected 2 keys, got %d", len(keys))
			}

			emb, ok, err := b.GetEmbedding(ctx, "a")
			if err != nil || !ok {
				t.Fatalf("GetEmbedding: err=%v ok=%v", err, ok)
			}
			if len(emb) != 2 || emb[0] != 1 || emb[1] != 2 {
				t.Fatalf("unexpected embedding: %v", emb)
			}

			_, ok, _ = b.GetEmbedding(ctx, "missing")
			if ok {
				t.Fatal("expected not found for missing key")
			}
		})
	}
}

func TestBackend_FlushAndLen(t *testing.T) {
	for name, factory := range factories() {
		t.Run(name, func(t *testing.T) {
			b := factory(t)
			ctx := context.Background()

			_ = b.Set(ctx, "a", nil, "va")
			_ = b.Set(ctx, "b", nil, "vb")

			n, _ := b.Len(ctx)
			if n != 2 {
				t.Fatalf("expected 2, got %d", n)
			}

			if err := b.Flush(ctx); err != nil {
				t.Fatalf("Flush: %v", err)
			}

			n, _ = b.Len(ctx)
			if n != 0 {
				t.Fatalf("expected 0 after flush, got %d", n)
			}
		})
	}
}

func TestBackend_Eviction(t *testing.T) {
	ctx := context.Background()

	t.Run("LRU", func(t *testing.T) {
		b, _ := NewLRUBackend[string, string](2)
		_ = b.Set(ctx, "a", nil, "1")
		_ = b.Set(ctx, "b", nil, "2")
		_ = b.Set(ctx, "c", nil, "3")

		n, _ := b.Len(ctx)
		if n != 2 {
			t.Fatalf("expected 2 after eviction, got %d", n)
		}
	})

	t.Run("LFU", func(t *testing.T) {
		b, _ := NewLFUBackend[string, string](2)
		_ = b.Set(ctx, "a", nil, "1")
		_ = b.Set(ctx, "b", nil, "2")
		// Access "a" so it has higher frequency
		_, _, _ = b.Get(ctx, "a")
		_ = b.Set(ctx, "c", nil, "3")

		n, _ := b.Len(ctx)
		if n != 2 {
			t.Fatalf("expected 2 after eviction, got %d", n)
		}
		// "a" should survive, "b" should be evicted
		ok, _ := b.Contains(ctx, "a")
		if !ok {
			t.Fatal("expected 'a' to survive (higher frequency)")
		}
	})

	t.Run("FIFO", func(t *testing.T) {
		b, _ := NewFIFOBackend[string, string](2)
		_ = b.Set(ctx, "a", nil, "1")
		_ = b.Set(ctx, "b", nil, "2")
		_ = b.Set(ctx, "c", nil, "3")

		n, _ := b.Len(ctx)
		if n != 2 {
			t.Fatalf("expected 2 after eviction, got %d", n)
		}
		// oldest ("a") should be evicted
		ok, _ := b.Contains(ctx, "a")
		if ok {
			t.Fatal("expected 'a' to be evicted (FIFO)")
		}
	})
}

func TestBackend_Overwrite(t *testing.T) {
	for name, factory := range factories() {
		t.Run(name, func(t *testing.T) {
			b := factory(t)
			ctx := context.Background()

			_ = b.Set(ctx, "k", []float64{1}, "old")
			_ = b.Set(ctx, "k", []float64{2}, "new")

			v, ok, _ := b.Get(ctx, "k")
			if !ok || v != "new" {
				t.Fatalf("expected new, got %q", v)
			}

			emb, ok, _ := b.GetEmbedding(ctx, "k")
			if !ok || emb[0] != 2 {
				t.Fatalf("expected updated embedding")
			}

			n, _ := b.Len(ctx)
			if n != 1 {
				t.Fatalf("overwrite should not increase count, got %d", n)
			}
		})
	}
}

func TestBackend_Close(t *testing.T) {
	for name, factory := range factories() {
		t.Run(name, func(t *testing.T) {
			b := factory(t)
			if err := b.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
		})
	}
}

// Compile-time interface compliance checks.
var (
	_ types.Backend[string, string] = (*LRUBackend[string, string])(nil)
	_ types.Backend[string, string] = (*LFUBackend[string, string])(nil)
	_ types.Backend[string, string] = (*FIFOBackend[string, string])(nil)
)
