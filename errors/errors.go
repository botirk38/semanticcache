// Package errors provides sentinel errors and typed error wrappers
// for the semanticcache library.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for configuration and validation.
var (
	ErrNilBackend       = errors.New("semanticcache: backend cannot be nil")
	ErrNilProvider      = errors.New("semanticcache: provider cannot be nil")
	ErrNilComparator    = errors.New("semanticcache: comparator cannot be nil")
	ErrZeroKey          = errors.New("semanticcache: key cannot be zero value")
	ErrCacheClosed      = errors.New("semanticcache: cache is closed")
	ErrInvalidThreshold = errors.New("semanticcache: threshold must be in [0, 1]")
	ErrInvalidN         = errors.New("semanticcache: n must be greater than 0")
	ErrEmptyBatch       = errors.New("semanticcache: batch cannot be empty")
)

// EmbeddingError wraps errors from embedding providers.
type EmbeddingError struct {
	Provider string
	Op       string
	Err      error
}

func (e *EmbeddingError) Error() string {
	return fmt.Sprintf("semanticcache: embedding %s failed (provider=%s): %v", e.Op, e.Provider, e.Err)
}

func (e *EmbeddingError) Unwrap() error { return e.Err }

// BackendError wraps errors from cache backends.
type BackendError struct {
	Backend string
	Op      string
	Err     error
}

func (e *BackendError) Error() string {
	return fmt.Sprintf("semanticcache: backend %s failed (backend=%s): %v", e.Op, e.Backend, e.Err)
}

func (e *BackendError) Unwrap() error { return e.Err }
