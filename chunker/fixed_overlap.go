package chunker

import (
	"fmt"

	"github.com/tiktoken-go/tokenizer"
)

// FixedOverlapChunker implements the Chunker interface using a fixed-size
// chunking strategy with overlap between chunks.
type FixedOverlapChunker struct {
	config   ChunkConfig
	encoding tokenizer.Codec
}

// NewFixedOverlapChunker creates a new FixedOverlapChunker with the given configuration.
// It uses tiktoken's cl100k_base encoding (used by OpenAI's text-embedding-3-small).
func NewFixedOverlapChunker(config ChunkConfig) (*FixedOverlapChunker, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid chunk config: %w", err)
	}

	// Get the cl100k_base encoding (used by OpenAI embeddings)
	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	return &FixedOverlapChunker{
		config:   config,
		encoding: enc,
	}, nil
}

// CountTokens counts the number of tokens in the given text.
func (c *FixedOverlapChunker) CountTokens(text string) (int, error) {
	if text == "" {
		return 0, nil
	}

	ids, _, err := c.encoding.Encode(text)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrTokenizerFailed, err)
	}

	return len(ids), nil
}

// ChunkText splits the text into overlapping chunks based on token count.
func (c *FixedOverlapChunker) ChunkText(text string) ([]Chunk, error) {
	if text == "" {
		return nil, ErrEmptyText
	}

	// Tokenize the entire text
	tokens, _, err := c.encoding.Encode(text)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenizerFailed, err)
	}

	totalTokens := len(tokens)

	// If text fits within chunk size, return single chunk
	if totalTokens <= c.config.ChunkSize {
		return []Chunk{
			{
				Text:       text,
				StartToken: 0,
				EndToken:   totalTokens,
				Index:      0,
			},
		}, nil
	}

	// Calculate number of chunks needed
	stride := c.config.ChunkSize - c.config.ChunkOverlap
	if stride <= 0 {
		stride = c.config.ChunkSize // Fallback if overlap is misconfigured
	}

	var chunks []Chunk
	chunkIndex := 0

	for start := 0; start < totalTokens; start += stride {
		end := start + c.config.ChunkSize
		if end > totalTokens {
			end = totalTokens
		}

		// Decode the token slice back to text
		chunkTokens := tokens[start:end]
		chunkText, err := c.encoding.Decode(chunkTokens)
		if err != nil {
			return nil, fmt.Errorf("failed to decode chunk %d: %w", chunkIndex, err)
		}

		chunks = append(chunks, Chunk{
			Text:       chunkText,
			StartToken: start,
			EndToken:   end,
			Index:      chunkIndex,
		})

		chunkIndex++

		// If we've reached the end, stop
		if end >= totalTokens {
			break
		}
	}

	return chunks, nil
}
