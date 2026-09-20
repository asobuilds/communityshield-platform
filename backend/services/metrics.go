package services

import (
	"sync"
	"sync/atomic"
	"time"
)

// processStart is the wall-clock time the process booted. Used to compute
// uptimeSeconds in health and metrics responses.
var processStart = time.Now()

// Metrics holds in-process counters for request rate, error rate, and
// rolling latency. All public methods are safe for concurrent use.
//
// Per-path aggregation must normalize dynamic segments (":id" -> ":param")
// to avoid cardinality explosions — see IncRequest. No per-path storage is
// kept today; this is a future concern.
type Metrics struct {
	requestsTotal  atomic.Int64
	requests2xx    atomic.Int64
	requests3xx    atomic.Int64
	requests4xx    atomic.Int64
	requests5xx    atomic.Int64
	errorsTotal    atomic.Int64

	mu sync.Mutex

	// rolling ring buffer of the last 1000 durations (ms, as float64)
	ring    [1000]float64
	ringPos int
	ringLen int
	sum     float64
}

var metrics = &Metrics{}

// ProcessStart returns the wall-clock time the process booted.
func ProcessStart() time.Time {
	return processStart
}

// SnapshotMetrics returns a point-in-time copy of all in-process counters.
func SnapshotMetrics() MetricsSnapshot {
	return metrics.Snapshot()
}

// IncRequest records a single completed request. It is called from the
// StructuredLogger middleware after each request. Cheap and non-allocating.
//
// Per-path aggregation must normalize dynamic segments (":id" -> ":param")
// to avoid cardinality explosions. No per-path storage is kept today.
func IncRequest(method, path string, status int, durationMs int64) {
	metrics.requestsTotal.Add(1)
	switch {
	case status >= 500:
		metrics.requests5xx.Add(1)
		metrics.errorsTotal.Add(1)
	case status >= 400:
		metrics.requests4xx.Add(1)
	case status >= 300:
		metrics.requests3xx.Add(1)
	default:
		metrics.requests2xx.Add(1)
	}
	metrics.recordDuration(float64(durationMs))
}

func (m *Metrics) recordDuration(d float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If the ring is full, the value at ringPos is about to be
	// overwritten — subtract it from the running sum first.
	if m.ringLen == len(m.ring) {
		m.sum -= m.ring[m.ringPos]
	}
	m.ring[m.ringPos] = d
	m.ringPos = (m.ringPos + 1) % len(m.ring)
	if m.ringLen < len(m.ring) {
		m.ringLen++
	}
	m.sum += d
}

// Snapshot returns a point-in-time copy of all counters.
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	avg := 0.0
	if m.ringLen > 0 {
		avg = m.sum / float64(m.ringLen)
	}
	return MetricsSnapshot{
		RequestsTotal:   m.requestsTotal.Load(),
		Requests2xx:     m.requests2xx.Load(),
		Requests3xx:     m.requests3xx.Load(),
		Requests4xx:     m.requests4xx.Load(),
		Requests5xx:     m.requests5xx.Load(),
		ErrorsTotal:     m.errorsTotal.Load(),
		AverageLatencyMs: avg,
		UptimeSeconds:   time.Since(processStart).Seconds(),
	}
}

// MetricsSnapshot is an immutable copy of the current metrics.
type MetricsSnapshot struct {
	RequestsTotal   int64
	Requests2xx     int64
	Requests3xx     int64
	Requests4xx     int64
	Requests5xx     int64
	ErrorsTotal     int64
	AverageLatencyMs float64
	UptimeSeconds   float64
}