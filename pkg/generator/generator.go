// Package generator provides log generation functionality for the genlog tool.
// It contains the core logic for creating fake log entries based on templates
// and random data generation.
package generator

import (
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/P1llus/genlog/pkg/config"
	"github.com/P1llus/genlog/pkg/output"
	"github.com/brianvoe/gofakeit/v7"
)

// Generator is responsible for generating fake log entries
// based on the provided configuration. It handles template selection,
// random value generation, and output management.
type Generator struct {
	config      *config.Config
	funcMap     template.FuncMap
	totalWeight int
	workers     []*output.Worker
	stopChan    chan struct{}
	wg          sync.WaitGroup
	maxCount    int
	doneChan    chan struct{} // Channel to signal completion
}

// NewGenerator creates a new log generator with the given configuration.
// It initializes the function map for template rendering and calculates
// the total template weight for weighted random selection.
//
// The function map includes all custom types from the configuration,
// making them available as placeholders in templates.
func NewGenerator(cfg *config.Config, maxCount int) (*Generator, error) {
	// Set the seed for randomization if provided
	if cfg.Seed != 0 {
		gofakeit.Seed(cfg.Seed)
	}

	// Calculate total weight for template selection
	totalWeight := 0
	for _, tpl := range cfg.Templates {
		totalWeight += tpl.Weight
	}

	// Create the generator instance
	g := &Generator{
		config:      cfg,
		totalWeight: totalWeight,
		stopChan:    make(chan struct{}),
		doneChan:    make(chan struct{}),
		maxCount:    maxCount,
	}

	// Initialize the function map for template rendering
	g.funcMap = g.createFuncMap(cfg.CustomTypes)

	// Initialize outputs and workers
	if err := g.initializeOutputs(); err != nil {
		return nil, fmt.Errorf("error initializing outputs: %w", err)
	}

	return g, nil
}

// initializeOutputs sets up all configured outputs and their workers
func (g *Generator) initializeOutputs() error {
	for _, outputCfg := range g.config.Outputs {
		// Calculate max count per worker
		maxCountPerWorker := 0
		if g.maxCount > 0 {
			maxCountPerWorker = g.maxCount / outputCfg.Workers
			if g.maxCount%outputCfg.Workers != 0 {
				maxCountPerWorker++ // Round up to ensure we generate at least maxCount
			}
		}

		// Create workers for this output
		for i := 0; i < outputCfg.Workers; i++ {
			// Create the output with worker ID
			out, err := output.NewOutput(outputCfg, i)
			if err != nil {
				return fmt.Errorf("error creating output %s: %w", outputCfg.Type, err)
			}

			worker := output.NewWorker(out, g, 100, maxCountPerWorker, g.stopChan) // Batch size of 100
			g.workers = append(g.workers, worker)
		}
	}

	return nil
}

// Done returns a channel that is closed when the generator has completed
// generating the requested number of logs
func (g *Generator) Done() chan struct{} {
	return g.doneChan
}

// Start begins generating and sending logs to all configured outputs
func (g *Generator) Start() {
	for _, worker := range g.workers {
		g.wg.Add(1)
		go func(w *output.Worker) {
			defer g.wg.Done()
			w.Start()
		}(worker)
	}

	// If we have a max count, start a goroutine to monitor completion
	if g.maxCount > 0 {
		go func() {
			g.wg.Wait()
			close(g.doneChan)
		}()
	}
}

// Stop gracefully stops all workers and closes outputs
func (g *Generator) Stop() error {
	close(g.stopChan)
	g.wg.Wait()

	// Close all outputs
	for _, worker := range g.workers {
		if err := worker.Output.Close(); err != nil {
			return fmt.Errorf("error closing output: %w", err)
		}
	}

	return nil
}

// selectWeightedTemplate selects a random template index based on the weights.
// Templates with higher weights have a proportionally higher chance of being selected.
func (g *Generator) selectWeightedTemplate() int {
	if g.totalWeight <= 0 || len(g.config.Templates) == 0 {
		return 0
	}

	r := gofakeit.IntRange(0, g.totalWeight-1)
	sum := 0
	for i, tpl := range g.config.Templates {
		sum += tpl.Weight
		if r < sum {
			return i
		}
	}
	return 0
}

// createFuncMap creates a map of functions that can be used in templates.
// The map includes:
// 1. All custom types from the configuration, each as a function returning a random value
// 2. Built-in helper functions like FormattedDate
//
// Note: The addLookupFunc functionality of gofakeit is not available when rendering
// inside go templates. This is why we have to create a map of the function names along
// with the function used to generate the random value.
func (g *Generator) createFuncMap(customTypes map[string][]string) template.FuncMap {
	funcMap := make(map[string]any)

	// Add each custom type as a function that returns a random value from its slice
	for typeName, values := range customTypes {
		// Create a function to properly capture the values for each custom type
		funcMap[typeName] = g.createRandomValueFunc(values)
	}

	// Add built-in helper functions
	funcMap["FormattedDate"] = func(format string) string {
		// Generate a random date within a reasonable range
		minDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		maxDate := time.Now() // Current time

		randomDate := gofakeit.DateRange(minDate, maxDate)
		return randomDate.Format(format)
	}

	return funcMap
}

// createRandomValueFunc creates a function that returns a random value from the given slice.
// This is used to translate configured custom types to a funcMap for the template functions.
func (g *Generator) createRandomValueFunc(values []string) func() string {
	return func() string {
		if len(values) == 0 {
			return ""
		}
		index := gofakeit.IntRange(0, len(values)-1)
		return values[index]
	}
}

// GenerateLogLine generates a single log line using a randomly selected template.
// This is useful for generating log lines programmatically without writing to a file,
// such as when streaming logs directly to another system.
func (g *Generator) GenerateLogLine() (string, error) {
	// First check if we have any templates
	if len(g.config.Templates) == 0 {
		return "", fmt.Errorf("no templates available")
	}

	templateIdx := g.selectWeightedTemplate()
	selectedTemplate := g.config.Templates[templateIdx].Template

	logLine, err := gofakeit.Template(selectedTemplate, &gofakeit.TemplateOptions{
		Funcs: g.funcMap,
	})
	if err != nil {
		return "", fmt.Errorf("error generating log line: %w", err)
	}

	return logLine, nil
}
