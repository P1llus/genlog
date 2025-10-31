// Package genlog provides high-performance fake log generation.
//
// It can be used as both a library and a CLI tool for generating
// realistic log data for testing and benchmarking purposes.
//
// Basic Usage:
//
//	cfg := &genlog.Config{
//		Templates: []genlog.Template{
//			{
//				Template: "{{timestamp}} [{{level}}] {{message}}",
//				Weight:   10,
//			},
//		},
//		CustomTypes: map[string][]string{
//			"level":   {"INFO", "WARN", "ERROR"},
//			"message": {"User logged in", "Request processed"},
//		},
//		Outputs: []genlog.OutputConfig{
//			{
//				Type: "file",
//				Config: map[string]interface{}{
//					"path": "output.log",
//				},
//			},
//		},
//	}
//
//	gen, err := genlog.New(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	gen.Start()
//	gen.Wait()
package genlog

import (
	"github.com/P1llus/genlog/pkg/config"
	"github.com/P1llus/genlog/pkg/generator"
)

// Generator is the main interface for log generation.
type Generator struct {
	gen *generator.Generator
}

// Config represents the main configuration.
type Config = config.Config

// Template represents a log template with weight.
type Template = config.Template

// OutputConfig represents an output configuration.
type OutputConfig = config.OutputConfig

// Performance represents performance tuning parameters.
type Performance = config.Performance

// Limits represents generation limits.
type Limits = config.Limits

// Stats represents generation statistics.
type Stats = generator.Stats

// Snapshot represents a point-in-time statistics snapshot.
type Snapshot = generator.Snapshot

// New creates a new log generator from configuration.
func New(cfg *Config) (*Generator, error) {
	gen, err := generator.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Generator{gen: gen}, nil
}

// NewFromFile creates a new log generator from a YAML configuration file.
func NewFromFile(path string) (*Generator, error) {
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	return New(cfg)
}

// Start begins log generation.
func (g *Generator) Start() error {
	return g.gen.Start()
}

// Stop gracefully stops log generation.
func (g *Generator) Stop() error {
	return g.gen.Stop()
}

// Wait blocks until generation completes (due to configured limits).
func (g *Generator) Wait() {
	g.gen.Wait()
}

// Stats returns the current generation statistics.
func (g *Generator) Stats() *Stats {
	return g.gen.Stats()
}

// Done returns a channel that's closed when generation completes.
func (g *Generator) Done() <-chan struct{} {
	return g.gen.Done()
}
