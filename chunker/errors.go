package chunker

import "errors"

// Common chunker errors
var (
	// ErrInvalidChunkSize indicates chunk size is invalid (<=0)
	ErrInvalidChunkSize = errors.New("chunk size must be positive")

	// ErrChunkSizeExceedsMax indicates chunk size exceeds max tokens
	ErrChunkSizeExceedsMax = errors.New("chunk size cannot exceed max tokens")

	// ErrInvalidOverlap indicates overlap value is invalid (<0)
	ErrInvalidOverlap = errors.New("overlap must be non-negative")

	// ErrOverlapTooLarge indicates overlap is >= chunk size
	ErrOverlapTooLarge = errors.New("overlap must be less than chunk size")

	// ErrInvalidMaxTokens indicates max tokens is invalid (<=0)
	ErrInvalidMaxTokens = errors.New("max tokens must be positive")

	// ErrEmptyText indicates text to chunk is empty
	ErrEmptyText = errors.New("cannot chunk empty text")

	// ErrTokenizerFailed indicates tokenization failed
	ErrTokenizerFailed = errors.New("tokenization failed")
)
