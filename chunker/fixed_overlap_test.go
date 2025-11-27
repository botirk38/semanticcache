package chunker

import (
	"strings"
	"testing"
)

func TestNewFixedOverlapChunker(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := DefaultChunkConfig()
		chunker, err := NewFixedOverlapChunker(config)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if chunker == nil {
			t.Fatal("expected chunker, got nil")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		config := ChunkConfig{
			MaxTokens:    512,
			ChunkSize:    1024, // Exceeds MaxTokens
			ChunkOverlap: 50,
		}
		_, err := NewFixedOverlapChunker(config)
		if err == nil {
			t.Fatal("expected error for invalid config, got nil")
		}
	})
}

func TestFixedOverlapChunker_CountTokens(t *testing.T) {
	chunker, err := NewFixedOverlapChunker(DefaultChunkConfig())
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	tests := []struct {
		name    string
		text    string
		wantMin int // Approximate minimum tokens
		wantMax int // Approximate maximum tokens
	}{
		{
			name:    "empty string",
			text:    "",
			wantMin: 0,
			wantMax: 0,
		},
		{
			name:    "short text",
			text:    "Hello, world!",
			wantMin: 2,
			wantMax: 5,
		},
		{
			name:    "longer text",
			text:    "This is a longer piece of text that should have more tokens.",
			wantMin: 10,
			wantMax: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := chunker.CountTokens(tt.text)
			if err != nil {
				t.Fatalf("CountTokens() error = %v", err)
			}
			if count < tt.wantMin || count > tt.wantMax {
				t.Errorf("CountTokens() = %d, want between %d and %d", count, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFixedOverlapChunker_ChunkText(t *testing.T) {
	t.Run("empty text", func(t *testing.T) {
		chunker, _ := NewFixedOverlapChunker(DefaultChunkConfig())
		_, err := chunker.ChunkText("")
		if err != ErrEmptyText {
			t.Errorf("expected ErrEmptyText, got %v", err)
		}
	})

	t.Run("text fits in single chunk", func(t *testing.T) {
		config := ChunkConfig{
			MaxTokens:    8191,
			ChunkSize:    512,
			ChunkOverlap: 50,
			Strategy:     FixedSizeOverlap,
		}
		chunker, _ := NewFixedOverlapChunker(config)

		text := "This is a short text that fits in one chunk."
		chunks, err := chunker.ChunkText(text)
		if err != nil {
			t.Fatalf("ChunkText() error = %v", err)
		}

		if len(chunks) != 1 {
			t.Errorf("expected 1 chunk, got %d", len(chunks))
		}
		if chunks[0].Text != text {
			t.Errorf("chunk text mismatch")
		}
		if chunks[0].Index != 0 {
			t.Errorf("expected index 0, got %d", chunks[0].Index)
		}
	})

	t.Run("text requires multiple chunks", func(t *testing.T) {
		config := ChunkConfig{
			MaxTokens:    8191,
			ChunkSize:    20, // Small chunk size to force multiple chunks
			ChunkOverlap: 5,
			Strategy:     FixedSizeOverlap,
		}
		chunker, _ := NewFixedOverlapChunker(config)

		// Create text that will definitely need chunking
		text := strings.Repeat("This is a test sentence. ", 50)

		chunks, err := chunker.ChunkText(text)
		if err != nil {
			t.Fatalf("ChunkText() error = %v", err)
		}

		// Should have multiple chunks
		if len(chunks) <= 1 {
			t.Errorf("expected multiple chunks, got %d", len(chunks))
		}

		// Verify chunk properties
		for i, chunk := range chunks {
			if chunk.Index != i {
				t.Errorf("chunk %d has wrong index: %d", i, chunk.Index)
			}
			if chunk.Text == "" {
				t.Errorf("chunk %d has empty text", i)
			}
			if chunk.EndToken <= chunk.StartToken {
				t.Errorf("chunk %d has invalid token range: %d-%d", i, chunk.StartToken, chunk.EndToken)
			}

			// Verify overlap (except for last chunk)
			if i > 0 && i < len(chunks)-1 {
				// There should be some overlap in token positions
				prevChunk := chunks[i-1]
				if chunk.StartToken >= prevChunk.EndToken {
					t.Errorf("no overlap between chunks %d and %d", i-1, i)
				}
			}
		}

		// Reconstruct text from chunks (without overlap) and verify coverage
		if chunks[0].StartToken != 0 {
			t.Errorf("first chunk should start at token 0, got %d", chunks[0].StartToken)
		}
	})

	t.Run("chunk boundaries", func(t *testing.T) {
		config := ChunkConfig{
			MaxTokens:    8191,
			ChunkSize:    10,
			ChunkOverlap: 3,
			Strategy:     FixedSizeOverlap,
		}
		chunker, _ := NewFixedOverlapChunker(config)

		text := "The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog."

		chunks, err := chunker.ChunkText(text)
		if err != nil {
			t.Fatalf("ChunkText() error = %v", err)
		}

		// Verify all chunks have proper indices
		for i, chunk := range chunks {
			if chunk.Index != i {
				t.Errorf("chunk index mismatch: expected %d, got %d", i, chunk.Index)
			}
		}
	})
}

func TestFixedOverlapChunker_Integration(t *testing.T) {
	config := ChunkConfig{
		MaxTokens:    100,
		ChunkSize:    30,
		ChunkOverlap: 10,
		Strategy:     FixedSizeOverlap,
	}
	chunker, err := NewFixedOverlapChunker(config)
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	// Create a long text
	text := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 20)

	// Count tokens
	tokenCount, err := chunker.CountTokens(text)
	if err != nil {
		t.Fatalf("CountTokens() error = %v", err)
	}

	// Chunk the text
	chunks, err := chunker.ChunkText(text)
	if err != nil {
		t.Fatalf("ChunkText() error = %v", err)
	}

	// If text exceeds chunk size, should have multiple chunks
	if tokenCount > config.ChunkSize {
		if len(chunks) <= 1 {
			t.Errorf("expected multiple chunks for %d tokens, got %d chunks", tokenCount, len(chunks))
		}
	}

	// Verify each chunk
	for i, chunk := range chunks {
		chunkTokens, err := chunker.CountTokens(chunk.Text)
		if err != nil {
			t.Fatalf("failed to count tokens in chunk %d: %v", i, err)
		}

		// Each chunk (except possibly the last) should be close to ChunkSize
		if i < len(chunks)-1 {
			if chunkTokens > config.ChunkSize {
				t.Errorf("chunk %d has %d tokens, exceeds ChunkSize %d", i, chunkTokens, config.ChunkSize)
			}
		}
	}
}
