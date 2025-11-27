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
			MaxTokens:    50, // Low limit to force chunking
			ChunkSize:    20,
			ChunkOverlap: 5,
			Strategy:     FixedSizeOverlap,
		}
		chunker, _ := NewFixedOverlapChunker(config)

		// Create text that exceeds MaxTokens
		text := strings.Repeat("This is a test sentence. ", 10) // Should exceed 50 tokens

		tokenCount, err := chunker.CountTokens(text)
		if err != nil {
			t.Fatalf("CountTokens() error = %v", err)
		}

		// Verify text exceeds MaxTokens
		if tokenCount <= config.MaxTokens {
			t.Fatalf("test setup error: text has %d tokens, should exceed MaxTokens %d", tokenCount, config.MaxTokens)
		}

		chunks, err := chunker.ChunkText(text)
		if err != nil {
			t.Fatalf("ChunkText() error = %v", err)
		}

		// Should have multiple chunks
		if len(chunks) <= 1 {
			t.Errorf("expected multiple chunks for %d tokens (>%d limit), got %d", tokenCount, config.MaxTokens, len(chunks))
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

func TestFixedOverlapChunker_GetMaxTokens(t *testing.T) {
	config := ChunkConfig{
		MaxTokens:    5000,
		ChunkSize:    512,
		ChunkOverlap: 50,
		Strategy:     FixedSizeOverlap,
	}
	chunker, err := NewFixedOverlapChunker(config)
	if err != nil {
		t.Fatalf("failed to create chunker: %v", err)
	}

	if chunker.GetMaxTokens() != 5000 {
		t.Errorf("GetMaxTokens() = %d, want 5000", chunker.GetMaxTokens())
	}
}

func TestFixedOverlapChunker_ChunkingOnlyWhenExceedsMaxTokens(t *testing.T) {
	t.Run("text within MaxTokens returns single chunk", func(t *testing.T) {
		config := ChunkConfig{
			MaxTokens:    100, // Low limit to force chunking decision
			ChunkSize:    50,
			ChunkOverlap: 10,
			Strategy:     FixedSizeOverlap,
		}
		chunker, _ := NewFixedOverlapChunker(config)

		// Create short text that fits within MaxTokens
		shortText := "This is a short text."
		tokenCount, err := chunker.CountTokens(shortText)
		if err != nil {
			t.Fatalf("CountTokens() error = %v", err)
		}

		// Verify text is within MaxTokens
		if tokenCount > config.MaxTokens {
			t.Fatalf("test setup error: short text has %d tokens, exceeds MaxTokens %d", tokenCount, config.MaxTokens)
		}

		chunks, err := chunker.ChunkText(shortText)
		if err != nil {
			t.Fatalf("ChunkText() error = %v", err)
		}

		// Should return exactly one chunk
		if len(chunks) != 1 {
			t.Errorf("expected 1 chunk for text within MaxTokens, got %d", len(chunks))
		}

		if chunks[0].Text != shortText {
			t.Errorf("chunk text should match original text")
		}
	})

	t.Run("text exceeding MaxTokens gets chunked", func(t *testing.T) {
		config := ChunkConfig{
			MaxTokens:    50, // Low limit to force chunking
			ChunkSize:    25,
			ChunkOverlap: 5,
			Strategy:     FixedSizeOverlap,
		}
		chunker, _ := NewFixedOverlapChunker(config)

		// Create long text that exceeds MaxTokens
		longText := strings.Repeat("This is a longer sentence that will exceed the token limit. ", 10)
		tokenCount, err := chunker.CountTokens(longText)
		if err != nil {
			t.Fatalf("CountTokens() error = %v", err)
		}

		// Verify text exceeds MaxTokens
		if tokenCount <= config.MaxTokens {
			t.Fatalf("test setup error: long text has %d tokens, should exceed MaxTokens %d", tokenCount, config.MaxTokens)
		}

		chunks, err := chunker.ChunkText(longText)
		if err != nil {
			t.Fatalf("ChunkText() error = %v", err)
		}

		// Should return multiple chunks
		if len(chunks) <= 1 {
			t.Errorf("expected multiple chunks for text exceeding MaxTokens (%d tokens > %d limit), got %d chunks",
				tokenCount, config.MaxTokens, len(chunks))
		}

		// Verify all chunks are properly formed
		for i, chunk := range chunks {
			if chunk.Index != i {
				t.Errorf("chunk %d has wrong index: got %d, want %d", i, chunk.Index, i)
			}
			if chunk.Text == "" {
				t.Errorf("chunk %d has empty text", i)
			}
		}
	})

	t.Run("chunking respects MaxTokens threshold", func(t *testing.T) {
		config := ChunkConfig{
			MaxTokens:    100,
			ChunkSize:    50,
			ChunkOverlap: 10,
			Strategy:     FixedSizeOverlap,
		}
		chunker, _ := NewFixedOverlapChunker(config)

		// Test various text lengths around the MaxTokens threshold
		testCases := []struct {
			name        string
			text        string
			expectChunk bool
		}{
			{"well under limit", "Short text.", false},
			{"at limit", strings.Repeat("word ", 20), false},   // Approximately at limit
			{"over limit", strings.Repeat("word ", 200), true}, // Definitely over limit
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tokenCount, err := chunker.CountTokens(tc.text)
				if err != nil {
					t.Fatalf("CountTokens() error = %v", err)
				}

				chunks, err := chunker.ChunkText(tc.text)
				if err != nil {
					t.Fatalf("ChunkText() error = %v", err)
				}

				if tc.expectChunk {
					if len(chunks) <= 1 {
						t.Errorf("expected multiple chunks for %d tokens (>%d limit), got %d chunks",
							tokenCount, config.MaxTokens, len(chunks))
					}
				} else {
					if len(chunks) != 1 {
						t.Errorf("expected single chunk for %d tokens (<=%d limit), got %d chunks",
							tokenCount, config.MaxTokens, len(chunks))
					}
				}
			})
		}
	})
}
