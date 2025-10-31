// Package output provides output implementations and consumer workers for writing logs.
//
// This package defines:
//   - The Output interface for log destinations
//   - Consumer workers that batch and write logs
//   - Built-in implementations for file and UDP outputs
//
// New output types can be added by implementing the Output interface.
package output

import (
	"fmt"
)

// Output represents a destination for log messages.
//
// Implementations should handle batched writes efficiently and
// ensure resources are properly released in Close().
type Output interface {
	// WriteBatch writes a batch of log lines to the destination.
	// The implementation should write all logs and flush as appropriate.
	WriteBatch(logs []string) error

	// Close releases any resources held by the output (files, connections, etc.).
	// It should flush any pending data before closing.
	Close() error

	// Name returns a human-readable name for this output for logging/debugging.
	Name() string
}

// NewOutput creates a new output based on the type and configuration.
func NewOutput(outputType string, config map[string]interface{}) (Output, error) {
	switch outputType {
	case "file":
		return NewFileOutput(config)
	case "udp":
		return NewUDPOutput(config)
	default:
		return nil, fmt.Errorf("unsupported output type: %s", outputType)
	}
}
