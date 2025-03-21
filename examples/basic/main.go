// Package main demonstrates basic usage of the genlog library.
// This example shows how to create a log generator from a configuration file,
// generate multiple log lines to a file, and generate individual log lines.
package main

import (
	"fmt"
	"log"

	"github.com/P1llus/genlog"
)

func main() {
	// Create a generator directly from a config file (simplest approach)
	// The config file contains templates and custom types for log generation
	gen, err := genlog.NewFromFile("examples/basic/config.yaml")
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Generate 10 log lines and write them to a file
	// This is useful for batch generation of test data
	fmt.Println("Generating logs to output.log...")
	err = gen.GenerateLogs("output.log", 10)
	if err != nil {
		log.Fatalf("Failed to generate logs: %v", err)
	}

	// Generate and print individual log lines
	// This demonstrates generating logs on-demand without writing to a file
	// Useful for streaming logs or integrating with other systems
	fmt.Println("\nGenerated individual log lines:")
	for i := 0; i < 5; i++ {
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			log.Fatalf("Failed to generate log line: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, logLine)
	}

	fmt.Println("\nTip: Check output.log to see all generated log lines.")
	fmt.Println("Each run will produce different random values unless a seed is specified.")
}
