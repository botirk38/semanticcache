// Package ollama provides an embedding provider backed by a local Ollama server.
//
// See https://github.com/ollama/ollama/blob/main/docs/api.md#generate-embeddings
package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultBaseURL = "http://localhost:11434"
	defaultModel   = "nomic-embed-text"
	defaultTokens  = 8192
)

// Provider implements EmbeddingProvider using a local Ollama server.
type Provider struct {
	model   string
	baseURL string
	client  *http.Client
}

// Option configures an Ollama Provider.
type Option func(*Provider)

// WithModel overrides the default model.
func WithModel(model string) Option { return func(p *Provider) { p.model = model } }

// WithBaseURL overrides the Ollama server URL.
func WithBaseURL(url string) Option { return func(p *Provider) { p.baseURL = url } }

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(p *Provider) { p.client = c } }

// New creates an Ollama embedding provider. No API key needed — runs locally.
func New(opts ...Option) *Provider {
	p := &Provider{model: defaultModel, baseURL: defaultBaseURL, client: http.DefaultClient}
	for _, o := range opts {
		o(p)
	}
	return p
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Error      string      `json:"error,omitempty"`
}

// EmbedText embeds a single text via the Ollama /api/embed endpoint.
func (p *Provider) EmbedText(text string) ([]float64, error) {
	body, err := json.Marshal(embedRequest{Model: p.model, Input: text})
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, p.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: request failed (is Ollama running?): %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("ollama: decode response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("ollama: %s", result.Error)
	}

	if len(result.Embeddings) == 0 {
		return nil, errors.New("ollama: no embedding returned")
	}

	return result.Embeddings[0], nil
}

// EmbedBatch embeds multiple texts sequentially.
func (p *Provider) EmbedBatch(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, errors.New("ollama: no texts provided")
	}
	results := make([][]float64, len(texts))
	for i, t := range texts {
		emb, err := p.EmbedText(t)
		if err != nil {
			return nil, fmt.Errorf("ollama: batch item %d: %w", i, err)
		}
		results[i] = emb
	}
	return results, nil
}

// Close is a no-op.
func (p *Provider) Close() {}

// GetMaxTokens returns the token limit.
func (p *Provider) GetMaxTokens() int { return defaultTokens }
