// Command basic demonstrates the core semanticcache API using the local
// hash-based provider (no API key required).
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/botirk38/semanticcache"
	"github.com/botirk38/semanticcache/options"
)

func main() {
	cache, err := semanticcache.New[string, string](
		options.WithLRUBackend[string, string](100),
		options.WithLocalProvider[string, string](128),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = cache.Close() }()

	ctx := context.Background()

	// Store some entries
	entries := []struct {
		key, text, value string
	}{
		{"q1", "What is Go?", "Go is a statically typed, compiled language."},
		{"q2", "How does caching work?", "Caching stores results for faster retrieval."},
		{"q3", "What is a goroutine?", "A goroutine is a lightweight thread managed by Go."},
	}
	for _, e := range entries {
		if err := cache.Set(ctx, e.key, e.text, e.value); err != nil {
			log.Fatal(err)
		}
	}

	n, _ := cache.Len(ctx)
	fmt.Printf("Stored %d entries\n", n)

	// Exact key lookup
	val, found, _ := cache.Get(ctx, "q1")
	fmt.Printf("Get q1: found=%v value=%q\n", found, val)

	// Semantic lookup (note: local provider is hash-based, so similarity
	// scores are not semantically meaningful -- this just demonstrates the API)
	match, _ := cache.Lookup(ctx, "Tell me about Go", 0.0)
	if match != nil {
		fmt.Printf("Lookup: value=%q score=%.4f\n", match.Value, match.Score)
	}

	// Top matches
	matches, _ := cache.TopMatches(ctx, "Tell me about Go", 3)
	fmt.Println("TopMatches:")
	for i, m := range matches {
		fmt.Printf("  %d. score=%.4f value=%q\n", i+1, m.Score, m.Value)
	}
}
