package semanticcache

import "errors"

// Sentinel errors returned by Cache methods.
var (
	// ErrClosed is returned when an operation is attempted on a closed cache.
	ErrClosed = errors.New("semanticcache: cache is closed")

	// ErrZeroKey is returned when a zero-value key is used.
	ErrZeroKey = errors.New("semanticcache: key cannot be zero value")

	// ErrInvalidN is returned when n <= 0 is passed to TopMatches.
	ErrInvalidN = errors.New("semanticcache: n must be positive")
)
