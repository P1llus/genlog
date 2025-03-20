// Package loggen provides a simple API for generating fake log data
// based on customizable templates and patterns.
//
// It can be used either as a library in your Go code or as a CLI tool.
package genlog

import (
	"github.com/P1llus/genlog/pkg/config"
	"github.com/P1llus/genlog/pkg/generator"
)

// Generator represents a log generator instance.
// It provides methods for generating fake log data.
type Generator struct {
	gen *generator.Generator
}

// NewFromConfig creates a new log generator from the provided configuration.
// This is useful when you want to programmatically create a configuration.
func NewFromConfig(cfg *config.Config) *Generator {
	return &Generator{
		gen: generator.NewGenerator(cfg),
	}
}

// NewFromFile creates a new log generator by reading configuration from a file.
// This is the simplest way to get started with genlog.
func NewFromFile(configFile string) (*Generator, error) {
	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		return nil, err
	}

	return &Generator{
		gen: generator.NewGenerator(cfg),
	}, nil
}

// GenerateLogs generates the specified number of log lines and writes them to the output file.
func (g *Generator) GenerateLogs(outputFile string, count int) error {
	return g.gen.GenerateLogs(outputFile, count)
}

// GenerateLogLine generates a single log line using a randomly selected template.
func (g *Generator) GenerateLogLine() (string, error) {
	return g.gen.GenerateLogLine()
}

// Config provides access to the configuration structures.
// This allows users to create their own configurations programmatically.
type Config = config.Config

// LogTemplate represents a single log template with its selection weight.
type LogTemplate = config.LogTemplate
