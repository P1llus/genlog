package generator

import (
	"time"

	"github.com/P1llus/genlog/pkg/config"
)

// Limiter checks if generation should stop based on configured limits.
type Limiter struct {
	maxCount    int64
	maxDuration time.Duration
	maxSize     int64
	startTime   time.Time
}

// NewLimiter creates a new limiter from configuration.
func NewLimiter(limits config.Limits) *Limiter {
	return &Limiter{
		maxCount:    limits.MaxCount,
		maxDuration: limits.MaxDuration,
		maxSize:     limits.MaxSize,
		startTime:   time.Now(),
	}
}

// HasLimits returns true if any limits are configured.
func (l *Limiter) HasLimits() bool {
	return l.maxCount > 0 || l.maxDuration > 0 || l.maxSize > 0
}

// HasDurationOrSizeLimits returns true if duration or size limits are configured.
// Count limits are handled by producers directly, so we only need active monitoring
// for duration and size limits.
func (l *Limiter) HasDurationOrSizeLimits() bool {
	return l.maxDuration > 0 || l.maxSize > 0
}

// ShouldStop checks if any limit has been reached.
func (l *Limiter) ShouldStop(stats *Stats) bool {
	// Check count limit
	if l.maxCount > 0 && stats.Generated() >= l.maxCount {
		return true
	}

	// Check duration limit
	if l.maxDuration > 0 && time.Since(l.startTime) >= l.maxDuration {
		return true
	}

	// Check size limit
	if l.maxSize > 0 && stats.BytesGenerated() >= l.maxSize {
		return true
	}

	return false
}
