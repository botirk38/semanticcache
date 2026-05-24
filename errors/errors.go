// Package errors provides sentinel errors and typed error values for the
// semantic cache library.
package errors

import "errors"

// Sentinel errors returned by the cache and its components.
var (
	// ErrClosed is returned when an operation is attempted on a closed cache.
	ErrClosed = errors.New("semanticcache: cache is closed")

	// ErrNilBackend is returned when a nil backend is provided.
	ErrNilBackend = errors.New("semanticcache: backend cannot be nil")

	// ErrNilProvider is returned when a nil embedding provider is provided.
	ErrNilProvider = errors.New("semanticcache: embedding provider cannot be nil")

	// ErrNilComparator is returned when a nil similarity function is provided.
	ErrNilComparator = errors.New("semanticcache: similarity comparator cannot be nil")

	// ErrZeroKey is returned when a zero-value key is used.
	ErrZeroKey = errors.New("semanticcache: key cannot be zero value")

	// ErrInvalidN is returned when n <= 0 is passed to TopMatches.
	ErrInvalidN = errors.New("semanticcache: n must be positive")
)

// EmbeddingError wraps an error from the embedding provider.
type EmbeddingError struct {
	Err error
}

func (e *EmbeddingError) Error() string {
	return "semanticcache: embedding provider: " + e.Err.Error()
}

func (e *EmbeddingError) Unwrap() error { return e.Err }

// BackendError wraps an error from the cache backend.
type BackendError struct {
	Op  string
	Err error
}

func (e *BackendError) Error() string {
	return "semanticcache: backend " + e.Op + ": " + e.Err.Error()
}

func (e *BackendError) Unwrap() error { return e.Err }
