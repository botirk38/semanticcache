package observability

import (
	"testing"
	"time"
)

func TestNopLoggerDoesNotPanic(t *testing.T) {
	var l NopLogger
	l.Debug("msg", "k", "v")
	l.Info("msg")
	l.Warn("msg")
	l.Error("msg")
}

func TestNopMetricsDoesNotPanic(t *testing.T) {
	var m NopMetrics
	m.CacheHit()
	m.CacheMiss()
	m.EmbedLatency(time.Millisecond)
	m.BackendLatency("Set", time.Millisecond)
}
