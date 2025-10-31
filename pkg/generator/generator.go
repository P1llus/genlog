// Package generator orchestrates the log generation pipeline with producer and consumer workers.
//
// This package implements a high-performance producer-consumer architecture where:
//   - Producer workers generate log strings using pre-compiled templates and value pools
//   - Logs are sent through a buffered channel for decoupling
//   - Consumer workers batch and write logs to configured outputs
//
// The architecture enables millions of logs per second throughput with minimal memory overhead.
package generator

import (
	"fmt"
	"sync"
	"time"

	"github.com/P1llus/genlog/pkg/config"
	"github.com/P1llus/genlog/pkg/output"
	"github.com/P1llus/genlog/pkg/template"
)

// Generator orchestrates the entire log generation pipeline from initialization to shutdown.
//
// It manages:
//   - Template compilation and value pool initialization
//   - Producer workers that generate log strings
//   - Consumer workers that batch and write logs
//   - Statistics tracking
//   - Graceful shutdown coordination
type Generator struct {
	config     *config.Config
	compiler   *template.Compiler
	producers  []*Producer
	consumers  []*output.Consumer
	outputs    []output.Output
	pipeline   chan string
	stopChan   chan struct{}
	doneChan   chan struct{}
	stats      *Stats
	producerWg sync.WaitGroup
	consumerWg sync.WaitGroup
	limiter    *Limiter
	stopOnce   sync.Once
}

// New creates a new generator from the provided configuration.
//
// This function:
//   - Validates configuration and applies defaults
//   - Compiles all templates into efficient execution plans
//   - Creates value pools for each producer worker
//   - Initializes the pipeline channel
//   - Sets up all configured outputs
//   - Creates producer and consumer workers
//
// The generator is ready to start but does not begin generating logs until Start() is called.
func New(cfg *config.Config) (*Generator, error) {
	// Validate config and apply defaults
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create template compiler
	templates := make([]struct {
		Template string
		Weight   int
	}, len(cfg.Templates))

	for i, t := range cfg.Templates {
		templates[i].Template = t.Template
		templates[i].Weight = t.Weight
	}

	compiler, err := template.NewCompiler(templates)
	if err != nil {
		return nil, fmt.Errorf("failed to compile templates: %w", err)
	}

	// Create pipeline channel
	pipeline := make(chan string, cfg.Performance.ChannelBuffer)

	// Create stats
	stats := NewStats()

	// Create outputs
	outputs := make([]output.Output, len(cfg.Outputs))
	for i, outputCfg := range cfg.Outputs {
		out, err := output.NewOutput(outputCfg.Type, outputCfg.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create output %d: %w", i, err)
		}
		outputs[i] = out
	}

	// Create generator
	g := &Generator{
		config:   cfg,
		compiler: compiler,
		outputs:  outputs,
		pipeline: pipeline,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		stats:    stats,
		limiter:  NewLimiter(cfg.Limits),
	}

	// Create producers
	g.producers = make([]*Producer, cfg.Performance.ProducerWorkers)

	// Calculate per-producer limit if max_count is set
	var maxCountPerProducer int64 = 0
	if cfg.Limits.MaxCount > 0 {
		// Divide total count among producers, rounding up to ensure we hit the limit
		maxCountPerProducer = (cfg.Limits.MaxCount + int64(cfg.Performance.ProducerWorkers) - 1) / int64(cfg.Performance.ProducerWorkers)
	}

	for i := 0; i < cfg.Performance.ProducerWorkers; i++ {
		// Each producer gets its own value pool with unique seed
		seed := cfg.Seed + int64(i)
		valuePool := template.NewValuePool(seed, cfg.Performance.ValuePoolSize, cfg.CustomTypes)

		g.producers[i] = NewProducer(
			i,
			compiler,
			valuePool,
			pipeline,
			g.stopChan,
			stats,
			maxCountPerProducer,
		)
	}

	// Create consumers (one per output)
	g.consumers = make([]*output.Consumer, 0)
	for i, out := range outputs {
		consumer := output.NewConsumer(
			i,
			pipeline,
			out,
			cfg.Performance.WriteBatchSize,
			stats,
		)
		g.consumers = append(g.consumers, consumer)
	}

	return g, nil
}

// Start begins log generation by launching all producer and consumer workers.
//
// Producers immediately begin generating logs and sending them to the pipeline channel.
// Consumers read from the channel, batch logs, and write them to outputs.
//
// If limits are configured, a monitor goroutine is started to detect when limits are reached.
func (g *Generator) Start() error {
	// Start all producers
	for _, producer := range g.producers {
		g.producerWg.Add(1)
		go func(p *Producer) {
			defer g.producerWg.Done()
			p.Run()
		}(producer)
	}

	// Start all consumers
	for _, consumer := range g.consumers {
		g.consumerWg.Add(1)
		go func(c *output.Consumer) {
			defer g.consumerWg.Done()
			if err := c.Run(); err != nil {
				fmt.Printf("Consumer error: %v\n", err)
			}
		}(consumer)
	}

	// Start limit monitor for duration/size limits
	// (count limits are handled by producers themselves)
	if g.limiter.HasDurationOrSizeLimits() {
		go g.monitorLimits()
	} else if g.config.Limits.MaxCount > 0 {
		// If only count limit, monitor producers completion
		go func() {
			g.producerWg.Wait()
			close(g.doneChan)
		}()
	}

	return nil
}

// monitorLimits checks if any limits have been reached and stops generation.
func (g *Generator) monitorLimits() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-g.stopChan:
			return
		case <-ticker.C:
			if g.limiter.ShouldStop(g.stats) {
				close(g.doneChan)
				g.Stop()
				return
			}
		}
	}
}

// Stop gracefully stops all workers and ensures data integrity.
//
// The shutdown sequence:
//  1. Signal all workers to stop via stopChan
//  2. Wait for producers to finish generating
//  3. Close the pipeline channel (signals "no more data")
//  4. Wait for consumers to drain all remaining logs
//  5. Close all output destinations
//
// This ensures all generated logs are written before shutdown completes.
// Stop can be called multiple times safely; subsequent calls are no-ops.
func (g *Generator) Stop() error {
	var err error

	// Ensure we only stop once
	g.stopOnce.Do(func() {
		// Step 1: Signal all workers to stop
		close(g.stopChan)

		// Step 2: Wait for all producers to finish
		g.producerWg.Wait()

		// Step 3: Close the pipeline channel (signals consumers no more data is coming)
		close(g.pipeline)

		// Step 4: Wait for all consumers to drain the channel and finish
		g.consumerWg.Wait()

		// Step 5: Close all outputs
		for _, out := range g.outputs {
			if closeErr := out.Close(); closeErr != nil {
				err = fmt.Errorf("failed to close output %s: %w", out.Name(), closeErr)
			}
		}
	})

	return err
}

// Wait blocks until generation completes.
//
// Generation completes when:
//   - A configured limit is reached (count, duration, or size)
//   - Stop() is called
//
// This is useful when you want to start generation and wait for it to finish.
func (g *Generator) Wait() {
	<-g.doneChan
}

// Stats returns the current generation statistics.
//
// Statistics include:
//   - Total logs generated
//   - Total logs written
//   - Bytes generated
//   - Generation duration
//   - Logs per second rate
func (g *Generator) Stats() *Stats {
	return g.stats
}

// Done returns a channel that's closed when generation completes.
//
// This allows integration with select statements for coordinating with other goroutines.
func (g *Generator) Done() <-chan struct{} {
	return g.doneChan
}
