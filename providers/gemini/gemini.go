// Package gemini provides an embedding provider backed by the Google Gemini
// (Generative Language) embedding API.
//
// See https://ai.google.dev/api/embeddings for API details.
package gemini

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
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	defaultModel   = "text-embedding-004"
	defaultTokens  = 2048
)

// Provider implements EmbeddingProvider using the Gemini embedding API.
type Provider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// Option configures a Gemini Provider.
type Option func(*Provider)

// WithModel overrides the default embedding model.
func WithModel(model string) Option {
	return func(p *Provider) { p.model = model }
}

// WithBaseURL overrides the API base URL.
func WithBaseURL(url string) Option {
	return func(p *Provider) { p.baseURL = url }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(p *Provider) { p.client = c }
}

// New creates a Gemini embedding provider.
func New(apiKey string, opts ...Option) (*Provider, error) {
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("gemini: API key is required (pass directly or set GEMINI_API_KEY)")
	}
	p := &Provider{
		apiKey:  apiKey,
		model:   defaultModel,
		baseURL: defaultBaseURL,
		client:  http.DefaultClient,
	}
	for _, o := range opts {
		o(p)
	}
	return p, nil
}

type embedContentRequest struct {
	Model   string       `json:"model"`
	Content contentParts `json:"content"`
}

type contentParts struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type embedContentResponse struct {
	Embedding struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// EmbedText embeds a single text using the Gemini embedding API.
func (p *Provider) EmbedText(text string) ([]float64, error) {
	url := fmt.Sprintf("%s/models/%s:embedContent?key=%s", p.baseURL, p.model, p.apiKey)

	body, err := json.Marshal(embedContentRequest{
		Model: "models/" + p.model,
		Content: contentParts{
			Parts: []part{{Text: text}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gemini: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result embedContentResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("gemini: decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("gemini: API error: %s", result.Error.Message)
	}

	return result.Embedding.Values, nil
}

// EmbedBatch embeds multiple texts by making individual API calls.
func (p *Provider) EmbedBatch(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, errors.New("gemini: no texts provided")
	}
	results := make([][]float64, len(texts))
	for i, t := range texts {
		emb, err := p.EmbedText(t)
		if err != nil {
			return nil, fmt.Errorf("gemini: batch item %d: %w", i, err)
		}
		results[i] = emb
	}
	return results, nil
}

// Close is a no-op.
func (p *Provider) Close() {}

// GetMaxTokens returns the token limit.
func (p *Provider) GetMaxTokens() int { return defaultTokens }
