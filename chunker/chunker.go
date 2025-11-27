package chunker

// Chunker defines the interface for text chunking strategies.
// Different implementations can provide various chunking approaches
// (fixed-size with overlap, semantic boundaries, sentence-based, etc.)
type Chunker interface {
	// ChunkText splits text into chunks based on the chunker's strategy
	// and token limits configured in the chunker.
	ChunkText(text string) ([]Chunk, error)

	// CountTokens counts the number of tokens in the given text.
	// This delegates to the underlying tokenizer.
	CountTokens(text string) (int, error)
}

// ChunkConfig holds configuration for text chunking behavior.
type ChunkConfig struct {
	// MaxTokens is the threshold that triggers chunking.
	// Text exceeding this limit will be split into chunks.
	// Default: 8191 (OpenAI text-embedding-3-small limit)
	MaxTokens int

	// ChunkSize is the target number of tokens per chunk.
	// Default: 512 tokens
	ChunkSize int

	// ChunkOverlap is the number of tokens to overlap between chunks.
	// This preserves context at chunk boundaries.
	// Default: 50 tokens
	ChunkOverlap int

	// Strategy specifies the chunking algorithm to use.
	// Default: FixedSizeOverlap
	Strategy ChunkStrategy
}

// ChunkStrategy represents the chunking algorithm type.
type ChunkStrategy string

const (
	// FixedSizeOverlap splits text into fixed-size chunks with overlap.
	FixedSizeOverlap ChunkStrategy = "fixed_overlap"

	// Future strategies:
	// SemanticBoundary ChunkStrategy = "semantic"
	// SentenceBased ChunkStrategy = "sentence"
	// ParagraphBased ChunkStrategy = "paragraph"
)

// Chunk represents a single chunk of text with its metadata.
type Chunk struct {
	// Text is the actual text content of this chunk
	Text string

	// StartToken is the starting token index in the original text
	StartToken int

	// EndToken is the ending token index in the original text
	EndToken int

	// Index is the chunk's position in the sequence (0-based)
	Index int
}

// DefaultChunkConfig returns the default chunking configuration.
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		MaxTokens:    8191, // OpenAI text-embedding-3-small limit
		ChunkSize:    512,  // Good balance of context and granularity
		ChunkOverlap: 50,   // Preserves context at boundaries
		Strategy:     FixedSizeOverlap,
	}
}

// Validate checks if the chunk configuration is valid.
func (c ChunkConfig) Validate() error {
	// Validate MaxTokens first
	if c.MaxTokens <= 0 {
		return ErrInvalidMaxTokens
	}

	// Validate ChunkSize
	if c.ChunkSize <= 0 {
		return ErrInvalidChunkSize
	}
	if c.ChunkSize > c.MaxTokens {
		return ErrChunkSizeExceedsMax
	}

	// Validate Overlap
	if c.ChunkOverlap < 0 {
		return ErrInvalidOverlap
	}
	if c.ChunkOverlap >= c.ChunkSize {
		return ErrOverlapTooLarge
	}

	return nil
}
