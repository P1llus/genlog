// Package main demonstrates basic usage of the genlog library.
package main

import (
	"fmt"
	"log"

	"github.com/P1llus/genlog"
)

func main() {
	// Create a generator directly from a config file (simplest approach)
	gen, err := genlog.NewFromFile("examples/basic/config.yaml")
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Generate 10 log lines to a file
	err = gen.GenerateLogs("output.log", 10)
	if err != nil {
		log.Fatalf("Failed to generate logs: %v", err)
	}

	// Generate and print individual log lines
	fmt.Println("\nGenerated individual log lines:")
	for i := 0; i < 5; i++ {
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			log.Fatalf("Failed to generate log line: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, logLine)
	}
}
