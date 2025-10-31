package main

import (
	"fmt"
	"log"

	"github.com/P1llus/genlog"
)

func main() {
	// Create a simple configuration
	cfg := &genlog.Config{
		Templates: []genlog.Template{
			{
				Template: "{{timestamp}} [{{level}}] {{message}}",
				Weight:   10,
			},
			{
				Template: `{"time":"{{timestamp}}","level":"{{level}}","msg":"{{message}}"}`,
				Weight:   5,
			},
		},
		CustomTypes: map[string][]string{
			"level": {"INFO", "WARN", "ERROR", "DEBUG"},
			"message": {
				"User logged in successfully",
				"Request processed",
				"Database query executed",
				"Cache hit",
				"API call completed",
			},
		},
		Outputs: []genlog.OutputConfig{
			{
				Type: "file",
				Config: map[string]interface{}{
					"path": "example-output.log",
				},
			},
		},
		Limits: genlog.Limits{
			MaxCount: 10000, // Generate 10k logs
		},
	}

	// Create generator
	gen, err := genlog.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	fmt.Println("Starting log generation...")

	// Start generation
	if err := gen.Start(); err != nil {
		log.Fatalf("Failed to start generator: %v", err)
	}

	// Wait for completion
	gen.Wait()

	// Stop and cleanup
	if err := gen.Stop(); err != nil {
		log.Fatalf("Failed to stop generator: %v", err)
	}

	// Print statistics
	stats := gen.Stats().Snapshot()
	fmt.Printf("\nGeneration Complete!\n")
	fmt.Printf("  Logs Generated: %d\n", stats.Generated)
	fmt.Printf("  Logs Written:   %d\n", stats.Written)
	fmt.Printf("  Total Size:     %.2f MB\n", float64(stats.BytesGenerated)/(1024*1024))
	fmt.Printf("  Duration:       %s\n", stats.Duration)
	fmt.Printf("  Rate:           %.2f logs/sec\n", stats.LogsPerSecond)
}
