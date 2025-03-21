// Package main demonstrates programmatic configuration of the genlog library.
// This example shows how to:
// 1. Create a configuration programmatically without a config file
// 2. Define multiple log templates with different weights
// 3. Define custom types for random value selection
// 4. Use a seed for reproducible results
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/P1llus/genlog"
)

func main() {
	// Create a configuration programmatically
	// This approach gives you full control over the configuration in code
	cfg := &genlog.Config{
		// Define multiple log templates with different formats and weights
		// Higher weights make templates more likely to be selected
		Templates: []genlog.LogTemplate{
			{
				// Standard timestamp format with level and message
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] {{message}}",
				Weight:   10, // This template will be selected ~66% of the time (10/15)
			},
			{
				// JSON format for structured logging
				Template: "{\"time\":\"{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}}\",\"level\":\"{{level}}\",\"msg\":\"{{message}}\"}",
				Weight:   5, // This template will be selected ~33% of the time (5/15)
			},
		},
		// Define custom types that can be used in templates
		// Values will be randomly selected from these lists
		CustomTypes: map[string][]string{
			"level": {
				"INFO", "WARN", "ERROR", "DEBUG",
			},
			"message": {
				"System starting up",
				"Connection established",
				"Transaction completed",
				"User authentication failed",
				"Resource not found",
			},
		},
		// Optional: set seed for reproducible results
		// Using the same seed will generate the same sequence of logs
		Seed: 12345,
	}

	// Create a generator from the config
	gen := genlog.NewFromConfig(cfg)

	// Output file for the generated logs
	outputFile := "advanced-output.log"

	// Delete previous output file if it exists
	if _, err := os.Stat(outputFile); err == nil {
		os.Remove(outputFile)
	}

	// Generate logs to a file
	fmt.Printf("Generating 20 log lines to %s...\n", outputFile)
	err := gen.GenerateLogs(outputFile, 20)
	if err != nil {
		log.Fatalf("Failed to generate logs: %v", err)
	}

	// Generate and print a few sample log lines
	fmt.Println("\nSample generated log lines:")
	for i := 0; i < 3; i++ {
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			log.Fatalf("Failed to generate log line: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, logLine)
	}

	fmt.Println("\nNote: Because we set a seed value (12345), these logs will be")
	fmt.Println("the same every time this example is run. Remove the seed for random logs.")
	fmt.Printf("\nCheck %s to see all generated log lines.\n", outputFile)
}
