package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/P1llus/genlog"
)

func main() {
	// Parse command line arguments
	configFile := flag.String("config", "config.yaml", "Path to the configuration file")
	count := flag.Int("count", 1000, "Number of logs to generate (0 for infinite)")
	flag.Parse()

	// Create generator from config file
	gen, err := genlog.NewFromFile(*configFile, *count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the generator
	fmt.Printf("Starting log generation... (count: %d)\n", *count)
	gen.Start()

	// Wait for either interrupt signal or completion
	if *count > 0 {
		fmt.Printf("Waiting for %d logs to be generated...\n", *count)
		// Wait for either signal or completion
		select {
		case <-sigChan:
			fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
		case <-gen.Done():
			fmt.Printf("\nSuccessfully generated %d logs!\n", *count)
		}
	} else {
		fmt.Println("Generating logs indefinitely. Press Ctrl+C to stop.")
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
	}

	// Stop the generator
	if err := gen.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping generator: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Log generation stopped successfully")
}
