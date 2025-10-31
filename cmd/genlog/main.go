package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/P1llus/genlog"
)

var (
	// version and commit are set via ldflags during build
	version = "dev"
	commit  = "none"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	initConfig := flag.Bool("init", false, "Generate a sample configuration file (usage: -init [filename])")
	flag.Parse()

	// Show version if requested
	if *showVersion {
		fmt.Printf("genlog version %s (commit: %s)\n", version, commit)
		return
	}

	// Generate sample config if requested
	if *initConfig {
		// Determine output filename: positional arg > -config flag > default
		outputFile := "config.yaml"
		if flag.NArg() > 0 {
			// Use positional argument if provided
			outputFile = flag.Arg(0)
		} else if *configFile != "config.yaml" {
			// Use -config flag if it was explicitly set
			outputFile = *configFile
		}

		// Check if file already exists
		if _, err := os.Stat(outputFile); err == nil {
			fmt.Fprintf(os.Stderr, "Error: File '%s' already exists\n", outputFile)
			fmt.Fprintf(os.Stderr, "Remove the existing file or specify a different name.\n")
			os.Exit(1)
		}

		if err := generateSampleConfig(outputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Sample configuration written to %s\n", outputFile)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit the configuration: vim %s\n", outputFile)
		fmt.Printf("  2. Generate logs: genlog -config %s\n", outputFile)
		return
	}

	// Load configuration
	gen, err := genlog.NewFromFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start generation
	fmt.Println("Starting log generation...")
	if err := gen.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting generator: %v\n", err)
		os.Exit(1)
	}

	// Start statistics reporting
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-gen.Done():
				return
			case <-ticker.C:
				stats := gen.Stats().Snapshot()
				fmt.Printf("\rGenerated: %d logs | %.2f logs/sec | %.2f MB | %.2f MB/sec",
					stats.Generated,
					stats.LogsPerSecond,
					float64(stats.BytesGenerated)/(1024*1024),
					stats.BytesPerSecond/(1024*1024),
				)
			}
		}
	}()

	// Wait for either completion or interrupt
	select {
	case <-gen.Done():
		fmt.Println("\n\nGeneration completed!")
	case <-sigChan:
		fmt.Println("\n\nReceived interrupt signal, shutting down gracefully...")
	}

	// Stop the generator
	if err := gen.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping generator: %v\n", err)
		os.Exit(1)
	}

	// Print final statistics
	stats := gen.Stats().Snapshot()
	fmt.Printf("\nFinal Statistics:\n")
	fmt.Printf("  Total Logs:      %d\n", stats.Generated)
	fmt.Printf("  Total Written:   %d\n", stats.Written)
	fmt.Printf("  Total Size:      %.2f MB\n", float64(stats.BytesGenerated)/(1024*1024))
	fmt.Printf("  Duration:        %s\n", stats.Duration.Round(time.Millisecond))
	fmt.Printf("  Avg Rate:        %.2f logs/sec\n", stats.LogsPerSecond)
	fmt.Printf("  Avg Throughput:  %.2f MB/sec\n", stats.BytesPerSecond/(1024*1024))

	fmt.Println("\nLog generation stopped successfully")
}

func generateSampleConfig(filename string) error {
	sampleConfig := `# genlog configuration file
# Generate realistic log data for testing and development

# Optional: Set seed for reproducible generation
# seed: 12345

# Log templates with weights
# Higher weight = higher probability of selection
templates:
  - template: '{{FormattedDate "2006-01-02T15:04:05.000Z07:00"}} [{{level}}] {{username}} - {{message}}'
    weight: 10
  - template: '{{FormattedDate "Jan 2 15:04:05"}} {{level}} [{{service}}] {{ipv4}} {{username}}: {{message}}'
    weight: 5
  - template: '{"timestamp":"{{FormattedDate "2006-01-02T15:04:05.000Z07:00"}}","level":"{{level}}","service":"{{service}}","message":"{{message}}","user":"{{username}}","ip":"{{ipv4}}"}'
    weight: 3

# Custom types that can be referenced in templates
custom_types:
  level:
    - INFO
    - WARNING
    - ERROR
    - DEBUG
  service:
    - API
    - AUTH
    - DATABASE
    - CACHE
    - FRONTEND
  username:
    - admin
    - user
    - service_account
    - system
  message:
    - "User authenticated successfully"
    - "Failed login attempt"
    - "Permission denied"
    - "Resource accessed"
    - "Request processed"

# Performance tuning (optional - these are sensible defaults)
performance:
  # Number of goroutines generating logs (default: NumCPU)
  # producer_workers: 8
  
  # Number of goroutines writing logs (default: 2)
  # consumer_workers: 2
  
  # Channel buffer size (default: 100000)
  # channel_buffer: 100000
  
  # Logs to batch before writing (default: 1000)
  # write_batch_size: 1000
  
  # Pre-generated random values per type (default: 10000)
  # value_pool_size: 10000

# Output destinations
outputs:
  - type: file
    config:
      path: "output.log"

# Optional: Generation limits (stops when ANY limit is reached)
limits:
  # Maximum number of logs to generate
  max_count: 100000
  
  # Maximum duration (examples: "5m", "1h", "30s")
  # max_duration: "5m"
  
  # Maximum total size in bytes
  # max_size: 1073741824  # 1GB
`

	return os.WriteFile(filename, []byte(sampleConfig), 0644)
}
