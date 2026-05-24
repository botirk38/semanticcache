// Package cohere provides an embedding provider backed by the Cohere Embed API.
//
// See https://docs.cohere.com/reference/embed for API details.
package cohere

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
	defaultBaseURL = "https://api.cohere.com/v2"
	defaultModel   = "embed-english-v3.0"
	defaultTokens  = 512
)

// Provider implements EmbeddingProvider using the Cohere Embed API.
type Provider struct {
	apiKey    string
	model     string
	baseURL   string
	inputType string
	client    *http.Client
}

// Option configures a Cohere Provider.
type Option func(*Provider)

// WithModel overrides the default embedding model.
func WithModel(model string) Option {
	return func(p *Provider) { p.model = model }
}

// WithBaseURL overrides the API base URL.
func WithBaseURL(url string) Option {
	return func(p *Provider) { p.baseURL = url }
}

// WithInputType sets the input type (e.g. "search_document", "search_query").
func WithInputType(t string) Option {
	return func(p *Provider) { p.inputType = t }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(p *Provider) { p.client = c }
}

// New creates a Cohere embedding provider. The API key is read from
// the apiKey parameter or the COHERE_API_KEY environment variable.
func New(apiKey string, opts ...Option) (*Provider, error) {
	if apiKey == "" {
		apiKey = os.Getenv("COHERE_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("cohere: API key is required (pass directly or set COHERE_API_KEY)")
	}
	p := &Provider{
		apiKey:    apiKey,
		model:     defaultModel,
		baseURL:   defaultBaseURL,
		inputType: "search_document",
		client:    http.DefaultClient,
	}
	for _, o := range opts {
		o(p)
	}
	return p, nil
}

type embedRequest struct {
	Model         string   `json:"model"`
	Texts         []string `json:"texts"`
	InputType     string   `json:"input_type"`
	EmbeddingType string   `json:"embedding_types"`
}

type embedResponse struct {
	Embeddings struct {
		Float [][]float64 `json:"float"`
	} `json:"embeddings"`
	Message string `json:"message"`
}

func (p *Provider) embed(texts []string) ([][]float64, error) {
	body, err := json.Marshal(embedRequest{
		Model:         p.model,
		Texts:         texts,
		InputType:     p.inputType,
		EmbeddingType: "float",
	})
	if err != nil {
		return nil, fmt.Errorf("cohere: marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, p.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cohere: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cohere: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cohere: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cohere: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cohere: decode response: %w", err)
	}

	if len(result.Embeddings.Float) != len(texts) {
		return nil, fmt.Errorf("cohere: expected %d embeddings, got %d", len(texts), len(result.Embeddings.Float))
	}

	return result.Embeddings.Float, nil
}

// EmbedText embeds a single text.
func (p *Provider) EmbedText(text string) ([]float64, error) {
	results, err := p.embed([]string{text})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

// EmbedBatch embeds multiple texts in a single API call.
func (p *Provider) EmbedBatch(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, errors.New("cohere: no texts provided")
	}
	return p.embed(texts)
}

// Close is a no-op.
func (p *Provider) Close() {}

// GetMaxTokens returns the token limit for the configured model.
func (p *Provider) GetMaxTokens() int { return defaultTokens }
