// Package local provides a deterministic, hash-based embedding provider
// that requires no external API. It is intended for testing, development,
// and CI environments where network access or API keys are not available.
//
// Embeddings are computed by hashing the input text with FNV-1a and
// spreading the hash across a fixed-dimension vector. Semantically
// different texts will produce different vectors, but there is no
// meaningful semantic similarity — use a real provider for production.
package local

import (
	"hash/fnv"
	"math"
)

// Provider is a deterministic, offline embedding provider.
type Provider struct {
	dimensions int
	maxTokens  int
}

// Option configures a local Provider.
type Option func(*Provider)

// WithDimensions sets the embedding vector length (default 128).
func WithDimensions(d int) Option {
	return func(p *Provider) { p.dimensions = d }
}

// WithMaxTokens sets the reported max token limit (default 8192).
func WithMaxTokens(n int) Option {
	return func(p *Provider) { p.maxTokens = n }
}

// New creates a local hash-based embedding provider.
func New(opts ...Option) *Provider {
	p := &Provider{dimensions: 128, maxTokens: 8192}
	for _, o := range opts {
		o(p)
	}
	return p
}

// EmbedText produces a deterministic embedding from the text hash.
func (p *Provider) EmbedText(text string) ([]float64, error) {
	h := fnv.New64a()
	h.Write([]byte(text))
	seed := h.Sum64()

	emb := make([]float64, p.dimensions)
	for i := range emb {
		// Mix seed with index to get per-dimension variation
		mixed := seed ^ uint64(i)*2654435761
		// Map to [-1, 1]
		emb[i] = math.Float64frombits((mixed&0x3FFFFFFFFFFFFFFF)|0x3FF0000000000000) - 1.5
	}

	// Normalize to unit vector
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

// EmbedBatch embeds multiple texts.
func (p *Provider) EmbedBatch(texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))
	for i, t := range texts {
		emb, err := p.EmbedText(t)
		if err != nil {
			return nil, err
		}
		results[i] = emb
	}
	return results, nil
}

// Close is a no-op for the local provider.
func (p *Provider) Close() {}

// GetMaxTokens returns the configured maximum token count.
func (p *Provider) GetMaxTokens() int { return p.maxTokens }

// GetDimensions returns the embedding vector length.
func (p *Provider) GetDimensions() int { return p.dimensions }
