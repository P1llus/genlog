// Package genlog provides a simple API for generating fake log data
// based on customizable templates and patterns.
//
// It can be used either as a library in your Go code or as a CLI tool.
package genlog

import (
	"fmt"

	"github.com/P1llus/genlog/pkg/config"
	"github.com/P1llus/genlog/pkg/generator"
)

// Generator represents the main log generator interface
type Generator interface {
	// Start begins generating and sending logs
	Start()
	// Stop gracefully stops the generator
	Stop() error
	// Done returns a channel that is closed when the generator has completed
	Done() chan struct{}
	// GenerateLogLine generates a single log line using a randomly selected template
	GenerateLogLine() (string, error)
}

// GeneratorStruct implements the Generator interface
type GeneratorStruct struct {
	gen *generator.Generator
}

// NewFromConfig creates a new generator from a config struct
func NewFromConfig(cfg *config.Config, maxCount int) (Generator, error) {
	gen, err := generator.NewGenerator(cfg, maxCount)
	if err != nil {
		return nil, fmt.Errorf("error creating generator: %w", err)
	}
	return &GeneratorStruct{gen: gen}, nil
}

// NewFromFile creates a new generator from a config file
func NewFromFile(configPath string, maxCount int) (Generator, error) {
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	return NewFromConfig(cfg, maxCount)
}

// Start begins generating and sending logs
func (g *GeneratorStruct) Start() {
	g.gen.Start()
}

// Stop gracefully stops the generator
func (g *GeneratorStruct) Stop() error {
	return g.gen.Stop()
}

// Done returns a channel that is closed when the generator has completed
func (g *GeneratorStruct) Done() chan struct{} {
	return g.gen.Done()
}

// GenerateLogLine generates a single log line using a randomly selected template.
// This is useful for generating log lines programmatically without writing to a file,
// such as when streaming logs directly to another system.
func (g *GeneratorStruct) GenerateLogLine() (string, error) {
	return g.gen.GenerateLogLine()
}

// Config provides access to the configuration structures.
// This allows users to create their own configurations programmatically.
type Config = config.Config

// LogTemplate represents a single log template with its selection weight.
type LogTemplate = config.LogTemplate

// OutputConfig represents a single output configuration
type OutputConfig = config.OutputConfig

// OutputType represents the type of output destination for logs
type OutputType = config.OutputType

const (
	// OutputTypeFile represents a file output destination
	OutputTypeFile = config.OutputTypeFile
	// OutputTypeUDP represents a UDP output destination
	OutputTypeUDP = config.OutputTypeUDP
)
