package inmemory

import (
	"context"
	"testing"
	"time"

	"github.com/botirk38/semanticcache/types"
)

func entry(emb []float64, val string) types.Entry[string] {
	return types.Entry[string]{Embedding: emb, Value: val}
}

func TestLRUBackendTTL(t *testing.T) {
	ttl := 50 * time.Millisecond
	b, err := NewLRUBackend[string, string](types.BackendConfig{Capacity: 10, TTL: ttl})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_ = b.Set(ctx, "k", entry([]float64{1}, "v"))

	// Before expiry
	_, found, _ := b.Get(ctx, "k")
	if !found {
		t.Fatal("expected key before TTL")
	}

	time.Sleep(ttl + 20*time.Millisecond)

	// After expiry
	_, found, _ = b.Get(ctx, "k")
	if found {
		t.Fatal("expected key to be expired")
	}
}

func TestLFUBackendTTL(t *testing.T) {
	ttl := 50 * time.Millisecond
	b, _ := NewLFUBackend[string, string](types.BackendConfig{Capacity: 10, TTL: ttl})
	ctx := context.Background()

	_ = b.Set(ctx, "k", entry([]float64{1}, "v"))

	_, found, _ := b.Get(ctx, "k")
	if !found {
		t.Fatal("expected key before TTL")
	}

	time.Sleep(ttl + 20*time.Millisecond)

	_, found, _ = b.Get(ctx, "k")
	if found {
		t.Fatal("expected key to be expired")
	}

	// Also check Contains
	_ = b.Set(ctx, "k2", entry([]float64{2}, "v2"))
	time.Sleep(ttl + 20*time.Millisecond)
	exists, _ := b.Contains(ctx, "k2")
	if exists {
		t.Fatal("expected Contains to return false after TTL")
	}
}

func TestFIFOBackendTTL(t *testing.T) {
	ttl := 50 * time.Millisecond
	b, _ := NewFIFOBackend[string, string](types.BackendConfig{Capacity: 10, TTL: ttl})
	ctx := context.Background()

	_ = b.Set(ctx, "k", entry([]float64{1}, "v"))

	_, found, _ := b.Get(ctx, "k")
	if !found {
		t.Fatal("expected key before TTL")
	}

	time.Sleep(ttl + 20*time.Millisecond)

	_, found, _ = b.Get(ctx, "k")
	if found {
		t.Fatal("expected key to be expired")
	}

	// Check Keys() filters expired
	_ = b.Set(ctx, "a", entry([]float64{1}, "va"))
	_ = b.Set(ctx, "b", entry([]float64{2}, "vb"))
	time.Sleep(ttl + 20*time.Millisecond)
	keys, _ := b.Keys(ctx)
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys after TTL, got %d", len(keys))
	}
}

func TestNoTTL(t *testing.T) {
	// Verify zero TTL means no expiration
	b, _ := NewLFUBackend[string, string](types.BackendConfig{Capacity: 10})
	ctx := context.Background()

	_ = b.Set(ctx, "k", entry([]float64{1}, "v"))
	time.Sleep(10 * time.Millisecond)
	_, found, _ := b.Get(ctx, "k")
	if !found {
		t.Fatal("with zero TTL, entries should not expire")
	}
}
