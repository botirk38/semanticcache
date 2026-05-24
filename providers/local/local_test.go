package local

import (
	"context"
	"math"
	"testing"
)

func TestProvider_EmbedText(t *testing.T) {
	p := New(64)
	ctx := context.Background()

	t.Run("deterministic", func(t *testing.T) {
		a, _ := p.EmbedText(ctx, "hello")
		b, _ := p.EmbedText(ctx, "hello")
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("embeddings differ at index %d", i)
			}
		}
	})

	t.Run("correct dimension", func(t *testing.T) {
		v, _ := p.EmbedText(ctx, "test")
		if len(v) != 64 {
			t.Fatalf("expected 64 dimensions, got %d", len(v))
		}
	})

	t.Run("unit length", func(t *testing.T) {
		v, _ := p.EmbedText(ctx, "normalise check")
		var norm float64
		for _, x := range v {
			norm += x * x
		}
		norm = math.Sqrt(norm)
		if math.Abs(norm-1.0) > 1e-10 {
			t.Fatalf("expected unit norm, got %f", norm)
		}
	})

	t.Run("different texts differ", func(t *testing.T) {
		a, _ := p.EmbedText(ctx, "foo")
		b, _ := p.EmbedText(ctx, "bar")
		same := true
		for i := range a {
			if a[i] != b[i] {
				same = false
				break
			}
		}
		if same {
			t.Fatal("different texts produced identical embeddings")
		}
	})
}

func TestProvider_EmbedBatch(t *testing.T) {
	p := New(32)
	ctx := context.Background()

	vecs, err := p.EmbedBatch(ctx, []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("EmbedBatch failed: %v", err)
	}
	if len(vecs) != 3 {
		t.Fatalf("expected 3 vectors, got %d", len(vecs))
	}
	for i, v := range vecs {
		if len(v) != 32 {
			t.Fatalf("vector %d has %d dims, expected 32", i, len(v))
		}
	}
}

func TestProvider_DefaultDimensions(t *testing.T) {
	p := New(0)
	v, _ := p.EmbedText(context.Background(), "x")
	if len(v) != 128 {
		t.Fatalf("expected default 128 dimensions, got %d", len(v))
	}
}

func TestProvider_Close(t *testing.T) {
	p := New(16)
	if err := p.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}
