// Package voyage provides an embedding provider backed by the Voyage AI API.
//
// See https://docs.voyageai.com/reference/embeddings-api for details.
package voyage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	defaultBaseURL = "https://api.voyageai.com/v1"
	defaultModel   = "voyage-3"
	defaultTokens  = 32000
)

// Provider implements EmbeddingProvider using the Voyage AI API.
type Provider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// Option configures a Voyage Provider.
type Option func(*Provider)

// WithModel overrides the default model.
func WithModel(model string) Option { return func(p *Provider) { p.model = model } }

// WithBaseURL overrides the API base URL.
func WithBaseURL(url string) Option { return func(p *Provider) { p.baseURL = url } }

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(p *Provider) { p.client = c } }

// New creates a Voyage AI embedding provider.
func New(apiKey string, opts ...Option) (*Provider, error) {
	if apiKey == "" {
		apiKey = os.Getenv("VOYAGE_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("voyage: API key is required (pass directly or set VOYAGE_API_KEY)")
	}
	p := &Provider{apiKey: apiKey, model: defaultModel, baseURL: defaultBaseURL, client: http.DefaultClient}
	for _, o := range opts {
		o(p)
	}
	return p, nil
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func (p *Provider) embed(texts []string) ([][]float64, error) {
	body, err := json.Marshal(embedRequest{Model: p.model, Input: texts})
	if err != nil {
		return nil, fmt.Errorf("voyage: marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, p.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("voyage: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("voyage: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("voyage: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("voyage: decode response: %w", err)
	}

	embeddings := make([][]float64, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = d.Embedding
	}
	return embeddings, nil
}

// EmbedText embeds a single text.
func (p *Provider) EmbedText(text string) ([]float64, error) {
	results, err := p.embed([]string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("voyage: no embedding returned")
	}
	return results[0], nil
}

// EmbedBatch embeds multiple texts in a single API call.
func (p *Provider) EmbedBatch(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, errors.New("voyage: no texts provided")
	}
	return p.embed(texts)
}

// Close is a no-op.
func (p *Provider) Close() {}

// GetMaxTokens returns the token limit.
func (p *Provider) GetMaxTokens() int { return defaultTokens }
