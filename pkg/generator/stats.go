package generator

import (
	"sync/atomic"
	"time"
)

// Stats holds real-time statistics about log generation using atomic operations for thread-safety.
//
// All counters use atomic operations to allow safe concurrent access from multiple goroutines
// without locking overhead.
type Stats struct {
	generated atomic.Int64 // Total logs generated
	written   atomic.Int64 // Total logs written to outputs
	bytesGen  atomic.Int64 // Approximate bytes generated
	startTime time.Time
}

// NewStats creates a new Stats instance.
func NewStats() *Stats {
	return &Stats{
		startTime: time.Now(),
	}
}

// IncrementGenerated increments the generated counter.
func (s *Stats) IncrementGenerated() {
	s.generated.Add(1)
}

// IncrementWritten increments the written counter.
func (s *Stats) IncrementWritten(count int) {
	s.written.Add(int64(count))
}

// AddBytes adds to the bytes generated counter.
func (s *Stats) AddBytes(bytes int) {
	s.bytesGen.Add(int64(bytes))
}

// Generated returns the total number of logs generated.
func (s *Stats) Generated() int64 {
	return s.generated.Load()
}

// Written returns the total number of logs written.
func (s *Stats) Written() int64 {
	return s.written.Load()
}

// BytesGenerated returns the approximate bytes generated.
func (s *Stats) BytesGenerated() int64 {
	return s.bytesGen.Load()
}

// Duration returns the elapsed time since generation started.
func (s *Stats) Duration() time.Duration {
	return time.Since(s.startTime)
}

// LogsPerSecond returns the current generation rate.
func (s *Stats) LogsPerSecond() float64 {
	duration := time.Since(s.startTime).Seconds()
	if duration == 0 {
		return 0
	}
	return float64(s.generated.Load()) / duration
}

// Snapshot represents a point-in-time view of generation statistics.
//
// Unlike Stats which uses atomic operations, Snapshot is a plain struct
// containing calculated values like rates and derived metrics.
type Snapshot struct {
	Generated      int64
	Written        int64
	BytesGenerated int64
	Duration       time.Duration
	LogsPerSecond  float64
	BytesPerSecond float64
}

// Snapshot returns a point-in-time snapshot of current statistics with calculated rates.
//
// This method atomically reads all counters and calculates derived metrics like
// logs per second and bytes per second based on elapsed time.
func (s *Stats) Snapshot() Snapshot {
	duration := time.Since(s.startTime)
	generated := s.generated.Load()
	written := s.written.Load()
	bytes := s.bytesGen.Load()

	seconds := duration.Seconds()
	logsPerSec := 0.0
	bytesPerSec := 0.0
	if seconds > 0 {
		logsPerSec = float64(generated) / seconds
		bytesPerSec = float64(bytes) / seconds
	}

	return Snapshot{
		Generated:      generated,
		Written:        written,
		BytesGenerated: bytes,
		Duration:       duration,
		LogsPerSecond:  logsPerSec,
		BytesPerSecond: bytesPerSec,
	}
}
