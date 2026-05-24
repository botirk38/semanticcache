// Package main demonstrates basic usage of the semanticcache library
// with the OpenAI embedding provider and an LRU backend.
//
// Run: OPENAI_API_KEY=sk-... go run ./examples/basic
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	semanticcache "github.com/botirk38/semanticcache"
	"github.com/botirk38/semanticcache/options"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	cache, err := semanticcache.New[string, string](
		options.WithLRUBackend[string, string](100),
		options.WithOpenAIProvider[string, string](apiKey),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = cache.Close() }()

	ctx := context.Background()

	// Store some entries
	entries := map[string]string{
		"greeting":  "Hello, how are you?",
		"farewell":  "Goodbye, see you later!",
		"gratitude": "Thank you very much!",
	}
	for key, text := range entries {
		if err := cache.Set(ctx, key, text, text); err != nil {
			log.Fatalf("Set %q: %v", key, err)
		}
		fmt.Printf("Stored: %q\n", key)
	}

	// Semantic lookup
	query := "Hi there, how's it going?"
	match, err := cache.Lookup(ctx, query, 0.7)
	if err != nil {
		log.Fatal(err)
	}
	if match != nil {
		fmt.Printf("\nQuery: %q\nMatch: %q (score: %.3f)\n", query, match.Value, match.Score)
	} else {
		fmt.Printf("\nQuery: %q\nNo match above threshold\n", query)
	}

	// Top matches
	matches, err := cache.TopMatches(ctx, query, 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nTop %d matches for %q:\n", len(matches), query)
	for i, m := range matches {
		fmt.Printf("  %d. %q (score: %.3f)\n", i+1, m.Value, m.Score)
	}
}
