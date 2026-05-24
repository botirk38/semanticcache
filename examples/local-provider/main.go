// Package main demonstrates using a custom hash-based provider,
// which requires no API key and works fully offline.
//
// Run: go run ./examples/local-provider
package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"math"

	semanticcache "github.com/botirk38/semanticcache"
	"github.com/botirk38/semanticcache/options"
)

// hashProvider is a simple deterministic embedding provider for demo purposes.
type hashProvider struct {
	dim int
}

func (p *hashProvider) EmbedText(text string) ([]float64, error) {
	h := fnv.New64a()
	h.Write([]byte(text))
	seed := h.Sum64()
	emb := make([]float64, p.dim)
	for i := range emb {
		mixed := seed ^ uint64(i)*2654435761
		emb[i] = math.Float64frombits((mixed&0x3FFFFFFFFFFFFFFF)|0x3FF0000000000000) - 1.5
	}
	var norm float64
	for _, v := range emb {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range emb {
			emb[i] /= norm
		}
	}
	return emb, nil
}

func (p *hashProvider) Close()            {}
func (p *hashProvider) GetMaxTokens() int { return 8192 }

func main() {
	cache, err := semanticcache.New[string, string](
		options.WithLRUBackend[string, string](100),
		options.WithCustomProvider[string, string](&hashProvider{dim: 128}),
		options.WithChunking[string, string](false),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Store entries
	data := []struct {
		key, text string
	}{
		{"doc1", "The quick brown fox jumps over the lazy dog"},
		{"doc2", "A fast auburn fox leaps above a sleepy canine"},
		{"doc3", "Machine learning models process natural language"},
	}
	for _, d := range data {
		if err := cache.Set(ctx, d.key, d.text, d.text); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Stored: %s -> %q\n", d.key, d.text)
	}

	fmt.Printf("\nCache size: ")
	n, _ := cache.Len(ctx)
	fmt.Println(n)

	// Exact key lookup
	val, found, err := cache.Get(ctx, "doc1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Get(doc1): found=%v, value=%q\n", found, val)

	// Top matches (note: hash-based provider, so similarity
	// scores are not semantically meaningful — this is for demo/testing)
	matches, err := cache.TopMatches(ctx, "fox jumping", 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nTopMatches for \"fox jumping\":\n")
	for i, m := range matches {
		fmt.Printf("  %d. score=%.4f value=%q\n", i+1, m.Score, m.Value)
	}
}
