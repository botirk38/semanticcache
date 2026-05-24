// Package local provides a hash-based embedding provider for testing and
// development. It produces deterministic embeddings from text using FNV
// hashing, so no external API key is required.
//
// The embeddings are NOT semantically meaningful — two sentences with
// similar meaning will NOT produce similar vectors. Use this provider only
// for unit tests, benchmarks, and local development where you need a
// provider that satisfies the EmbeddingProvider interface without network
// calls.
package local

import (
	"context"
	"hash/fnv"
	"math"
)

// Provider generates deterministic embeddings by hashing input text.
type Provider struct {
	dimensions int
}

// New creates a Provider that produces vectors of the given dimension.
func New(dimensions int) *Provider {
	if dimensions <= 0 {
		dimensions = 128
	}
	return &Provider{dimensions: dimensions}
}

// EmbedText hashes text into a deterministic float64 vector and normalises it
// to unit length so cosine similarity is well-defined.
func (p *Provider) EmbedText(_ context.Context, text string) ([]float64, error) {
	vec := make([]float64, p.dimensions)
	h := fnv.New64a()
	for i := range vec {
		h.Reset()
		_, _ = h.Write([]byte(text))
		_, _ = h.Write([]byte{byte(i), byte(i >> 8)})
		vec[i] = float64(h.Sum64())/float64(math.MaxUint64)*2 - 1
	}
	// L2-normalise
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec, nil
}

// EmbedBatch embeds multiple texts by calling EmbedText for each one.
func (p *Provider) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	out := make([][]float64, len(texts))
	for i, t := range texts {
		v, err := p.EmbedText(ctx, t)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// Close is a no-op.
func (p *Provider) Close() error { return nil }
