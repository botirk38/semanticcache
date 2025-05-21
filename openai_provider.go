package semanticcache

import (
	"context"
	"errors"
	"os"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	DefaultOpenAIModel = openai.EmbeddingModelTextEmbedding3Small
)

// OpenAIProvider uses OpenAI's API to embed text.
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAIProvider creates an embedding provider for OpenAI.
// If apiKey is empty, it uses os.Getenv("OPENAI_API_KEY").
// If model is empty, it uses DefaultOpenAIModel.
func NewOpenAIProvider(apiKey, model string) (*OpenAIProvider, error) {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, errors.New("OpenAI API key is required")
		}
	}
	if model == "" {
		model = DefaultOpenAIModel
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &OpenAIProvider{client: &client, model: model}, nil
}

// EmbedText sends the embedding request to OpenAI.
func (p *OpenAIProvider) EmbedText(text string) ([]float32, error) {
	resp, err := p.client.Embeddings.New(context.Background(), openai.EmbeddingNewParams{
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
	// OpenAI returns []float64; convert to []float32
	embeddingF64 := resp.Data[0].Embedding
	embeddingF32 := make([]float32, len(embeddingF64))
	for i, v := range embeddingF64 {
		embeddingF32[i] = float32(v)
	}
	return embeddingF32, nil
}

func (p *OpenAIProvider) Close() {}
