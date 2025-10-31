// Package main demonstrates using custom functions in genlog templates.
//
// Custom functions allow you to register your own Go functions that can be called
// from within log templates, enabling custom formatting, transformations, and logic.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/P1llus/genlog"
)

func main() {
	// Create temporary directory for output
	tmpDir, err := os.MkdirTemp("", "genlog-customfunc-example-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "custom-functions.log")

	// Create configuration with custom functions
	cfg := &genlog.Config{
		Templates: []genlog.Template{
			{
				// Use custom functions for formatting
				Template: "[{{Uppercase \"alert\"}}] {{Timestamp}} - {{ipv4}}",
				Weight:   5,
			},
			{
				// Multiple custom functions and placeholders
				Template: "{{level}} | {{Hash \"user123\" \"8\"}} | {{message}}",
				Weight:   3,
			},
			{
				// Custom functions can work with strings
				Template: "{{Repeat \"=\" \"20\"}} {{Uppercase \"security event\"}} {{Repeat \"=\" \"20\"}}",
				Weight:   2,
			},
		},
		CustomTypes: map[string][]string{
			"level":   {"info", "warning", "error", "critical"},
			"message": {"Login attempt", "API call", "Database query", "Cache refresh"},
		},
		// Register custom functions
		CustomFunctions: map[string]genlog.TemplateFunc{
			// Convert string to uppercase
			"Uppercase": func(args ...string) string {
				if len(args) == 0 {
					return ""
				}
				return strings.ToUpper(args[0])
			},

			// Pad string to specified width
			"Pad": func(args ...string) string {
				if len(args) < 2 {
					return ""
				}
				str := args[0]
				width := 10
				fmt.Sscanf(args[1], "%d", &width)
				return fmt.Sprintf("%-*s", width, str)
			},

			// Mask IP address (keep first N octets)
			"Mask": func(args ...string) string {
				if len(args) < 2 {
					return ""
				}
				ip := args[0]
				keep := 2
				fmt.Sscanf(args[1], "%d", &keep)

				parts := strings.Split(ip, ".")
				if len(parts) != 4 {
					return ip
				}

				for i := keep; i < 4; i++ {
					parts[i] = "***"
				}
				return strings.Join(parts, ".")
			},

			// Generate simple hash (first N chars of a string hash)
			"Hash": func(args ...string) string {
				if len(args) < 2 {
					return ""
				}
				str := args[0]
				length := 8
				fmt.Sscanf(args[1], "%d", &length)

				// Simple hash-like string (for demo purposes)
				hash := fmt.Sprintf("%x", []byte(str))
				if len(hash) > length {
					hash = hash[:length]
				}
				return hash
			},

			// Repeat string N times
			"Repeat": func(args ...string) string {
				if len(args) < 2 {
					return ""
				}
				str := args[0]
				count := 1
				fmt.Sscanf(args[1], "%d", &count)
				return strings.Repeat(str, count)
			},

			// Custom timestamp formatter
			"Timestamp": func(args ...string) string {
				return time.Now().Format("2006-01-02 15:04:05.000")
			},
		},
		Outputs: []genlog.OutputConfig{
			{
				Type: "file",
				Config: map[string]interface{}{
					"path": outputFile,
				},
			},
		},
		Limits: genlog.Limits{
			MaxCount: 20,
		},
		Seed: 12345, // For reproducible output
	}

	// Create and run generator
	gen, err := genlog.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	fmt.Println("Generating logs with custom functions...")
	gen.Start()
	gen.Wait()
	gen.Stop()

	// Display results
	content, err := os.ReadFile(outputFile)
	if err != nil {
		log.Fatalf("Failed to read output: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Generated Logs:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println(string(content))
	fmt.Println(strings.Repeat("=", 70))

	stats := gen.Stats().Snapshot()
	fmt.Printf("\nGenerated %d logs in %v\n", stats.Generated, stats.Duration)
	fmt.Println("\nCustom functions demonstrated:")
	fmt.Println("  • Uppercase: Convert text to uppercase")
	fmt.Println("  • Pad: Pad strings to fixed width")
	fmt.Println("  • Mask: Mask IP addresses for privacy")
	fmt.Println("  • Hash: Generate hash-like strings")
	fmt.Println("  • Repeat: Repeat strings")
	fmt.Println("  • Timestamp: Custom timestamp formatting")
}
