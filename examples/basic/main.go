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
	gen, err := genlog.NewFromFile("examples/basic/config.yaml", 10) // Generate 10 logs
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Start generating logs
	gen.Start()

	// Wait for completion
	<-gen.Done()

	// Generate and print individual log lines
	// This demonstrates generating logs on-demand without writing to an output
	// Useful for adhoc log generation or when you need to generate logs for other purposes
	fmt.Println("\nGenerated individual log lines:")
	for i := 0; i < 5; i++ {
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			log.Fatalf("Failed to generate log line: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, logLine)
	}

	fmt.Println("Each run will produce different random values unless a seed is specified.")
}
