package openai

import (
	"context"
	"errors"
	"os"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

const (
	// DefaultOpenAIModel is the default embedding model.
	DefaultOpenAIModel = openai.EmbeddingModelTextEmbedding3Small
)

// OpenAIConfig provides configuration for the OpenAI embedding provider.
type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	OrgID   string
	Model   string
}

// OpenAIProvider uses OpenAI's API to embed text.
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAIProvider creates a new OpenAI embedding provider.
func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, errors.New("OpenAI API key is required")
		}
	}

	model := config.Model
	if model == "" {
		model = DefaultOpenAIModel
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if config.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(config.BaseURL))
	}
	if config.OrgID != "" {
		opts = append(opts, option.WithOrganization(config.OrgID))
	}

	client := openai.NewClient(opts...)
	return &OpenAIProvider{client: &client, model: model}, nil
}

// EmbedText computes the embedding vector for a single piece of text.
func (p *OpenAIProvider) EmbedText(ctx context.Context, text string) ([]float64, error) {
	resp, err := p.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: p.model,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{text},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("no embedding returned by OpenAI")
	}
	return resp.Data[0].Embedding, nil
}

// EmbedBatch embeds multiple texts in a single API call.
func (p *OpenAIProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, errors.New("no texts provided for batch embedding")
	}
	if len(texts) > 2048 {
		return nil, errors.New("batch size exceeds OpenAI limit of 2048 texts")
	}

	resp, err := p.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: p.model,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: texts,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) != len(texts) {
		return nil, errors.New("number of embeddings returned does not match number of texts")
	}

	embeddings := make([][]float64, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}
	return embeddings, nil
}

// Close releases resources held by the provider.
func (p *OpenAIProvider) Close() error { return nil }
