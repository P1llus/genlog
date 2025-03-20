package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/P1llus/genlog"
)

func main() {
	// Parse command line arguments
	configFile := flag.String("config", "config.yaml", "Path to the configuration file")
	outputFile := flag.String("output", "generated.log", "Path to the output file")
	count := flag.Int("count", 10, "Number of log samples to generate")
	flag.Parse()

	// Create generator from config file
	gen, err := genlog.NewFromFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		os.Exit(1)
	}

	// Generate logs
	err = gen.GenerateLogs(*outputFile, *count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating logs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %d log entries to %s\n", *count, *outputFile)
}
