// Package main demonstrates usage with manually created configuration.
package main

import (
	"fmt"
	"log"

	"github.com/P1llus/genlog"
)

func main() {
	// Create a configuration programmatically
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] {{message}}",
				Weight:   10,
			},
			{
				Template: "{\"time\":\"{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}}\",\"level\":\"{{level}}\",\"msg\":\"{{message}}\"}",
				Weight:   5,
			},
		},
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
		Seed: 12345, // Optional: set for reproducible results
	}

	// Create a generator from the config
	gen := genlog.NewFromConfig(cfg)

	// Generate logs to a file
	err := gen.GenerateLogs("advanced-output.log", 20)
	if err != nil {
		log.Fatalf("Failed to generate logs: %v", err)
	}

	// Generate and print a few sample log lines
	fmt.Println("Sample generated log lines:")
	for i := 0; i < 3; i++ {
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			log.Fatalf("Failed to generate log line: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, logLine)
	}
}
