package openai

import (
	"context"
	"errors"
	"os"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

const (
	DefaultOpenAIModel = openai.EmbeddingModelTextEmbedding3Small
)

// OpenAI model-specific token limits
var openAIModelLimits = map[string]int{
	openai.EmbeddingModelTextEmbedding3Small: 8191,
	openai.EmbeddingModelTextEmbedding3Large: 8191,
	openai.EmbeddingModelTextEmbeddingAda002: 8191,
}

// OpenAIProvider uses OpenAI's API to embed text.
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// OpenAIConfig provides configuration options for OpenAI embedding provider
type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	OrgID   string
	Model   string
}

// NewOpenAIProvider creates an embedding provider for OpenAI.
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

// EmbedText sends the embedding request to OpenAI.
func (p *OpenAIProvider) EmbedText(text string) ([]float64, error) {
	resp, err := p.client.Embeddings.New(context.Background(), openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(p.model),
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
	// Return native float64 from OpenAI
	return resp.Data[0].Embedding, nil
}

// EmbedBatch sends a batch embedding request to OpenAI.
// This is more efficient than calling EmbedText multiple times as it
// makes a single API call for all texts. OpenAI supports up to 2048 texts per request.
func (p *OpenAIProvider) EmbedBatch(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, errors.New("no texts provided for batch embedding")
	}

	// OpenAI supports up to 2048 texts per request
	if len(texts) > 2048 {
		return nil, errors.New("batch size exceeds OpenAI limit of 2048 texts")
	}

	resp, err := p.client.Embeddings.New(context.Background(), openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(p.model),
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

	// Return native float64 embeddings from OpenAI
	embeddings := make([][]float64, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

func (p *OpenAIProvider) Close() {}

// GetMaxTokens returns the maximum number of tokens this OpenAI model can handle.
func (p *OpenAIProvider) GetMaxTokens() int {
	if limit, ok := openAIModelLimits[p.model]; ok {
		return limit
	}
	return 8191 // Safe default for unknown models
}
