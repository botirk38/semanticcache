// Package observability provides pluggable Logger and Metrics interfaces
// for the semanticcache library. Users can supply their own implementations
// (e.g. wrapping slog, zap, prometheus, etc.) via functional options.
package observability

import "time"

// Logger is a minimal structured-logging interface.
// Implementations should be safe for concurrent use.
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

// Metrics is a minimal metrics interface for cache observability.
// Implementations should be safe for concurrent use.
type Metrics interface {
	// CacheHit records a cache hit event.
	CacheHit()
	// CacheMiss records a cache miss event.
	CacheMiss()
	// EmbedLatency records the duration of an embedding call.
	EmbedLatency(d time.Duration)
	// BackendLatency records the duration of a backend operation.
	BackendLatency(op string, d time.Duration)
}

// NopLogger is a no-op logger that discards all output.
type NopLogger struct{}

func (NopLogger) Debug(string, ...any) {}
func (NopLogger) Info(string, ...any)  {}
func (NopLogger) Warn(string, ...any)  {}
func (NopLogger) Error(string, ...any) {}

// NopMetrics is a no-op metrics implementation.
type NopMetrics struct{}

func (NopMetrics) CacheHit()                            {}
func (NopMetrics) CacheMiss()                           {}
func (NopMetrics) EmbedLatency(time.Duration)           {}
func (NopMetrics) BackendLatency(string, time.Duration) {}
