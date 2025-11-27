package chunker

import (
	"testing"
)

func TestDefaultChunkConfig(t *testing.T) {
	config := DefaultChunkConfig()

	if config.MaxTokens != 8191 {
		t.Errorf("expected MaxTokens=8191, got %d", config.MaxTokens)
	}
	if config.ChunkSize != 512 {
		t.Errorf("expected ChunkSize=512, got %d", config.ChunkSize)
	}
	if config.ChunkOverlap != 50 {
		t.Errorf("expected ChunkOverlap=50, got %d", config.ChunkOverlap)
	}
	if config.Strategy != FixedSizeOverlap {
		t.Errorf("expected Strategy=FixedSizeOverlap, got %s", config.Strategy)
	}
}

func TestChunkConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ChunkConfig
		wantErr error
	}{
		{
			name: "valid config",
			config: ChunkConfig{
				MaxTokens:    8191,
				ChunkSize:    512,
				ChunkOverlap: 50,
				Strategy:     FixedSizeOverlap,
			},
			wantErr: nil,
		},
		{
			name: "chunk size zero",
			config: ChunkConfig{
				MaxTokens:    8191,
				ChunkSize:    0,
				ChunkOverlap: 50,
			},
			wantErr: ErrInvalidChunkSize,
		},
		{
			name: "chunk size negative",
			config: ChunkConfig{
				MaxTokens:    8191,
				ChunkSize:    -1,
				ChunkOverlap: 50,
			},
			wantErr: ErrInvalidChunkSize,
		},
		{
			name: "chunk size exceeds max",
			config: ChunkConfig{
				MaxTokens:    512,
				ChunkSize:    1024,
				ChunkOverlap: 50,
			},
			wantErr: ErrChunkSizeExceedsMax,
		},
		{
			name: "overlap negative",
			config: ChunkConfig{
				MaxTokens:    8191,
				ChunkSize:    512,
				ChunkOverlap: -1,
			},
			wantErr: ErrInvalidOverlap,
		},
		{
			name: "overlap equals chunk size",
			config: ChunkConfig{
				MaxTokens:    8191,
				ChunkSize:    512,
				ChunkOverlap: 512,
			},
			wantErr: ErrOverlapTooLarge,
		},
		{
			name: "overlap exceeds chunk size",
			config: ChunkConfig{
				MaxTokens:    8191,
				ChunkSize:    512,
				ChunkOverlap: 600,
			},
			wantErr: ErrOverlapTooLarge,
		},
		{
			name: "max tokens zero",
			config: ChunkConfig{
				MaxTokens:    0,
				ChunkSize:    512,
				ChunkOverlap: 50,
			},
			wantErr: ErrInvalidMaxTokens,
		},
		{
			name: "max tokens negative",
			config: ChunkConfig{
				MaxTokens:    -1,
				ChunkSize:    512,
				ChunkOverlap: 50,
			},
			wantErr: ErrInvalidMaxTokens,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
