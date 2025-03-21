package genlog_test

import (
	"fmt"
	"os"

	"github.com/P1llus/genlog"
)

// This example demonstrates how to create a generator from a config file
// and generate multiple log lines to a file.
func Example_generateFromFile() {
	// In real code, you would use a real config file path
	configFile := "config.yaml"

	// Create generator from config file
	gen, err := genlog.NewFromFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		return
	}

	// Generate 100 logs to output.log
	err = gen.GenerateLogs("output.log", 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating logs: %v\n", err)
		return
	}

	fmt.Println("Successfully generated logs to output.log")
	// Output: Successfully generated logs
}

// This example shows how to generate a single log line programmatically.
func Example_generateSingleLine() {
	// Create a simple configuration programmatically
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] User {{username}} logged in from {{IPV4Address}}",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"username": {"john", "alice", "bob"},
		},
	}

	// Create a generator with this config
	gen := genlog.NewFromConfig(cfg)

	// Generate a single log line,
	logLine, err := gen.GenerateLogLine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating log line: %v\n", err)
		return
	}

	fmt.Println("Generated log line successfully: ", logLine)
	// Output: Generated log line successfully
}

// This example demonstrates how to create a more complex configuration
// with multiple templates and custom types.
func Example_complexConfiguration() {
	// Create a configuration with multiple templates and weights
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] User {{username}} logged in from {{IPV4Address}}",
				Weight:   10, // Higher weight means more frequent selection
			},
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [ERROR] Failed login attempt for user {{username}} from {{IPV4Address}}",
				Weight:   3,
			},
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [WARN] High CPU usage: {{percentage}}%",
				Weight:   5,
			},
		},
		CustomTypes: map[string][]string{
			"username":   {"admin", "user", "guest", "system"},
			"percentage": {"78", "85", "91", "95", "99"},
		},
		// Optional seed for reproducible results
		Seed: 12345,
	}

	// Create generator from this config
	gen := genlog.NewFromConfig(cfg)

	// Generate 5 logs to a temporary file
	tempFile := "temp_complex_example.log"
	err := gen.GenerateLogs(tempFile, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating logs: %v\n", err)
		return
	}

	// Clean up the temp file
	os.Remove(tempFile)

	fmt.Println("Complex example completed")
	// Output: Complex example completed
}
