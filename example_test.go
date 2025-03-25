package genlog_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/P1llus/genlog"
)

// This example demonstrates how to create a generator from a config file
// and generate logs to multiple outputs.
func Example_generateFromFile() {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "config-example-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary config file
	configContent := `
templates:
  - template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] {{message}}"
    weight: 1
custom_types:
  message:
    - "Test message 1"
    - "Test message 2"
outputs:
  - type: file
    workers: 1
    config:
      filename: "` + filepath.Join(tmpDir, "output.log") + `"
seed: 12345
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		return
	}

	// Create generator from config file with a max count of 100 logs
	gen, err := genlog.NewFromFile(configFile, 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		return
	}

	// Start generating logs
	gen.Start()

	// Wait for completion
	<-gen.Done()

	// Stop the generator
	if err := gen.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping generator: %v\n", err)
		return
	}

	fmt.Println("Successfully generated logs")
	// Output:
	// Successfully generated logs
}

// This example shows how to generate a single log line programmatically.
func Example_generateSingleLine() {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "single-line-example-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple configuration programmatically
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] User {{username}} logged in from {{IPv4Address}}",
				Weight:   1,
			},
		},
		Seed: 12345,
		CustomTypes: map[string][]string{
			"username": {"alice"},
		},
		Outputs: []genlog.OutputConfig{
			{
				Type:    genlog.OutputTypeFile,
				Workers: 1,
				Config: map[string]interface{}{
					"filename": filepath.Join(tmpDir, "output.log"),
				},
			},
		},
	}

	// Create a generator with this config
	gen, err := genlog.NewFromConfig(cfg, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		return
	}

	// Generate a single log line
	_, err = gen.GenerateLogLine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating log line: %v\n", err)
		return
	}

	fmt.Printf("Generated log line")
	// Output: Generated log line
}

// This example demonstrates how to create a more complex configuration
// with multiple templates, custom types, and outputs.
func Example_complexConfiguration() {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "complex-example-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create a configuration with multiple templates and outputs
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] User {{username}} logged in from {{IPv4Address}}",
				Weight:   10,
			},
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [ERROR] Failed login attempt for user {{username}} from {{IPv4Address}}",
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
		Outputs: []genlog.OutputConfig{
			{
				Type:    genlog.OutputTypeFile,
				Workers: 2,
				Config: map[string]interface{}{
					"filename": filepath.Join(tmpDir, "app.log"),
				},
			},
			//{
			//	Type:    genlog.OutputTypeUDP,
			//	Workers: 1,
			//	Config: map[string]interface{}{
			//		"address": "localhost:514",
			//	},
			//},
		},
		Seed: 12345, // Optional seed for reproducible results
	}

	// Create generator from this config
	gen, err := genlog.NewFromConfig(cfg, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		return
	}

	// Start generating logs
	gen.Start()

	// Wait for completion
	<-gen.Done()

	// Stop the generator
	if err := gen.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping generator: %v\n", err)
		return
	}

	fmt.Println("Complex example completed")
	// Output: Complex example completed
}
