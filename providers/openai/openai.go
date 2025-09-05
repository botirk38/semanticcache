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
func (p *OpenAIProvider) EmbedText(text string) ([]float32, error) {
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
	// OpenAI returns []float64; convert to []float32
	embeddingF64 := resp.Data[0].Embedding
	embeddingF32 := make([]float32, len(embeddingF64))
	for i, v := range embeddingF64 {
		embeddingF32[i] = float32(v)
	}
	return embeddingF32, nil
}

func (p *OpenAIProvider) Close() {}
