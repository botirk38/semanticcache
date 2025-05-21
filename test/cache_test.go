package semanticcache_test

import (
	"testing"

	"github.com/botirk38/semanticcache/semanticcache"
)

func cosine(a, b []float32) float32 {
	var dot, na, nb float32
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	return dot / (sqrt(na) * sqrt(nb))
}

func sqrt(x float32) float32 {
	z := x
	for range 10 {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func TestLookup(t *testing.T) {
	cache, err := semanticcache.New(10, cosine)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	if err := cache.Set("foo", []float32{0.1, 0.2}, "bar"); err != nil {
		t.Fatalf("Failed to set cache entry: %v", err)
	}

	val, ok := cache.Lookup([]float32{0.1, 0.2}, 0.9)
	if !ok || val != "bar" {
		t.Errorf("Expected bar, got %v", val)
	}
}
