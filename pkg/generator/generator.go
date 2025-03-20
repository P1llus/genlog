// Package generator provides log generation functionality for the genlog tool.
package generator

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/P1llus/genlog/pkg/config"
	"github.com/brianvoe/gofakeit/v7"
)

// Generator is responsible for generating fake log entries
// based on the provided configuration.
type Generator struct {
	config      *config.Config
	funcMap     map[string]any
	totalWeight int
}

// NewGenerator creates a new log generator with the given configuration.
// It initializes the function map and calculates the total template weight.
func NewGenerator(cfg *config.Config) *Generator {
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
	}

	// Initialize the function map for template rendering
	g.funcMap = g.createFuncMap(cfg.CustomTypes)

	return g
}

// GenerateLogs generates the specified number of log lines and writes them to the output file.
// It returns an error if file operations fail or if log generation encounters problems.
func (g *Generator) GenerateLogs(outputFile string, count int) error {
	// Delete file if it already exists
	if _, err := os.Stat(outputFile); err == nil {
		err = os.Remove(outputFile)
		if err != nil {
			return fmt.Errorf("error deleting output file: %w", err)
		}
		fmt.Printf("Deleted existing output file: %s\n", outputFile)
	}

	// Create output file
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	fmt.Printf("Generating %d log lines to %s...\n", count, outputFile)

	for i := 0; i < count; i++ {
		// Select a random template based on weight
		templateIdx := g.selectWeightedTemplate()
		selectedTemplate := g.config.Templates[templateIdx].Template

		// Generate a log line using the template with custom functions
		logLine, err := gofakeit.Template(selectedTemplate, &gofakeit.TemplateOptions{
			Funcs: g.funcMap,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating log line: %v\n", err)
			continue
		}

		// Write the log line to the output file
		_, err = writer.WriteString(string(logLine) + "\n")
		if err != nil {
			return fmt.Errorf("error writing to output file: %w", err)
		}

		// Show progress
		if i > 0 && i%100 == 0 {
			fmt.Printf("%d log lines generated...\n", i)
		}
	}

	fmt.Printf("Successfully generated %d log lines to %s\n", count, outputFile)
	return nil
}

// selectWeightedTemplate selects a random template index based on the weights.
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
// The addLookupFunc functionality of gofakeit is not available when rendering inside go templates.
// This is why we have to create a map of the function names along with the function used to generate the random value.
func (g *Generator) createFuncMap(customTypes map[string][]string) map[string]any {
	funcMap := make(map[string]any)

	// Add each custom type as a function that returns a random value from its slice
	for typeName, values := range customTypes {
		// Create a function to properly capture the values for each custom type
		funcMap[typeName] = g.createRandomValueFunc(values)
		fmt.Printf("Added custom function: %s with %d possible values\n", typeName, len(values))
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
// This is useful for generating log lines programmatically without writing to a file.
func (g *Generator) GenerateLogLine() (string, error) {
	templateIdx := g.selectWeightedTemplate()
	selectedTemplate := g.config.Templates[templateIdx].Template

	logLine, err := gofakeit.Template(selectedTemplate, &gofakeit.TemplateOptions{
		Funcs: g.funcMap,
	})
	if err != nil {
		return "", fmt.Errorf("error generating log line: %w", err)
	}

	return string(logLine), nil
}
