package genlog_test

import (
	"regexp"
	"testing"

	"github.com/P1llus/genlog"
)

// TestNewFromConfig tests creating a generator from a programmatically created config
func TestNewFromConfig(t *testing.T) {
	// Create a simple test configuration
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "Test log message",
				Weight:   1,
			},
		},
	}

	// Create generator from config
	gen, err := genlog.NewFromConfig(cfg, 4)
	if err != nil {
		t.Fatalf("Failed to create generator from config: %v", err)
	}

	// Test generating a log line
	line, err := gen.GenerateLogLine()
	if err != nil {
		t.Fatalf("Failed to generate log line: %v", err)
	}
	if line != "Test log message" {
		t.Errorf("Generated log line doesn't match template. Got: %s, Expected: Test log message", line)
	}
}

// TestGenerateLogLine tests the functionality of generating individual log lines
func TestGenerateLogLine(t *testing.T) {
	// Create a test configuration with multiple templates and custom types
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "[{{level}}] Simple test message",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"level": {"INFO", "DEBUG", "ERROR"},
		},
		// Set a seed for deterministic testing
		Seed: 12345,
	}

	gen, err := genlog.NewFromConfig(cfg, 4)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Generate a few log lines and verify they match expected pattern
	for i := 0; i < 5; i++ {
		line, err := gen.GenerateLogLine()
		if err != nil {
			t.Fatalf("Failed to generate log line: %v", err)
		}

		// Verify the log line matches the expected pattern
		pattern := `^\[(INFO|DEBUG|ERROR)\] Simple test message$`
		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			t.Fatalf("Error matching pattern: %v", err)
		}
		if !matched {
			t.Errorf("Generated log line doesn't match expected pattern. Got: %s", line)
		}
	}
}

// TestNewFromFileError tests error handling when reading from a non-existent config file
func TestNewFromFileError(t *testing.T) {
	_, err := genlog.NewFromFile("nonexistent_config.yaml", 4)
	if err == nil {
		t.Errorf("Expected error when reading from non-existent file, got nil")
	}
}
