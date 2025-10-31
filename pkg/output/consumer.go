package output

import (
	"fmt"
	"time"
)

// Stats interface for tracking written logs.
type Stats interface {
	IncrementWritten(count int)
}

// Consumer reads log strings from a channel and writes them in batches.
type Consumer struct {
	id        int
	input     <-chan string
	output    Output
	batchSize int
	stats     Stats
}

// NewConsumer creates a new consumer worker.
func NewConsumer(
	id int,
	input <-chan string,
	output Output,
	batchSize int,
	stats Stats,
) *Consumer {
	return &Consumer{
		id:        id,
		input:     input,
		output:    output,
		batchSize: batchSize,
		stats:     stats,
	}
}

// Run starts the consumer worker loop.
// Consumers drain the input channel until it's closed, then exit.
func (c *Consumer) Run() error {
	batch := make([]string, 0, c.batchSize)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case log, ok := <-c.input:
			if !ok {
				// Channel closed, flush any remaining batch and exit
				if len(batch) > 0 {
					if err := c.output.WriteBatch(batch); err != nil {
						return fmt.Errorf("consumer %d: failed to write final batch: %w", c.id, err)
					}
					c.stats.IncrementWritten(len(batch))
				}
				return nil
			}

			batch = append(batch, log)

			// Write batch if it's full
			if len(batch) >= c.batchSize {
				if err := c.output.WriteBatch(batch); err != nil {
					return fmt.Errorf("consumer %d: failed to write batch: %w", c.id, err)
				}
				c.stats.IncrementWritten(len(batch))
				batch = batch[:0]
			}

		case <-ticker.C:
			// Flush partial batch on timeout
			if len(batch) > 0 {
				if err := c.output.WriteBatch(batch); err != nil {
					return fmt.Errorf("consumer %d: failed to write batch: %w", c.id, err)
				}
				c.stats.IncrementWritten(len(batch))
				batch = batch[:0]
			}
		}
	}
}
