package local

import (
	"math"
	"testing"
)

func TestDeterministic(t *testing.T) {
	p := New()
	a, _ := p.EmbedText("hello world")
	b, _ := p.EmbedText("hello world")

	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("embedding not deterministic at dim %d: %f != %f", i, a[i], b[i])
		}
	}
}

func TestDifferentTexts(t *testing.T) {
	p := New()
	a, _ := p.EmbedText("hello")
	b, _ := p.EmbedText("world")

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
}

func TestUnitVector(t *testing.T) {
	p := New(WithDimensions(256))
	emb, _ := p.EmbedText("test unit norm")

	var norm float64
	for _, v := range emb {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if math.Abs(norm-1.0) > 1e-10 {
		t.Fatalf("expected unit vector, got norm=%f", norm)
	}
}

func TestCustomDimensions(t *testing.T) {
	p := New(WithDimensions(64))
	emb, _ := p.EmbedText("test")
	if len(emb) != 64 {
		t.Fatalf("expected 64 dims, got %d", len(emb))
	}
}

func TestMaxTokens(t *testing.T) {
	p := New(WithMaxTokens(4096))
	if p.GetMaxTokens() != 4096 {
		t.Fatalf("expected 4096, got %d", p.GetMaxTokens())
	}
}

func TestEmbedBatch(t *testing.T) {
	p := New()
	texts := []string{"a", "b", "c"}
	results, err := p.EmbedBatch(texts)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// Verify batch matches individual calls
	for i, text := range texts {
		single, _ := p.EmbedText(text)
		for j := range single {
			if single[j] != results[i][j] {
				t.Fatalf("batch result differs from single at text=%q dim=%d", text, j)
			}
		}
	}
}

func TestClose(t *testing.T) {
	p := New()
	p.Close() // should not panic
}
