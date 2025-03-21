package genlog_test

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
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
	gen := genlog.NewFromConfig(cfg)
	if gen == nil {
		t.Fatalf("Failed to create generator from config")
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

	gen := genlog.NewFromConfig(cfg)

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

// TestGenerateLogs tests generating multiple logs and writing to a file
func TestGenerateLogs(t *testing.T) {
	// Create a temp file for testing
	outputFile := "test_output.log"
	defer os.Remove(outputFile) // Clean up after test

	// Create a simple test configuration
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "Test log entry {{Number 1 100}}",
				Weight:   1,
			},
		},
		Seed: 12345,
	}

	gen := genlog.NewFromConfig(cfg)

	// Generate logs
	count := 10
	err := gen.GenerateLogs(outputFile, count)
	if err != nil {
		t.Fatalf("Failed to generate logs: %v", err)
	}

	// Read the output file
	content, err := ioutil.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify file content
	lines := strings.Split(string(content), "\n")
	// Last line might be empty if file ends with newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) != count {
		t.Errorf("Expected %d log lines, got %d", count, len(lines))
	}

	// Check pattern of each line
	pattern := `^Test log entry \d+$`
	for i, line := range lines {
		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			t.Fatalf("Error matching pattern on line %d: %v", i, err)
		}
		if !matched {
			t.Errorf("Line %d doesn't match expected pattern. Got: %s", i, line)
		}
	}
}

// TestNewFromFileError tests error handling when reading from a non-existent config file
func TestNewFromFileError(t *testing.T) {
	_, err := genlog.NewFromFile("nonexistent_config.yaml")
	if err == nil {
		t.Errorf("Expected error when reading from non-existent file, got nil")
	}
}

// Creates a temporary YAML config file for testing
func createTempConfigFile(t *testing.T) (string, func()) {
	content := `
templates:
  - template: "Test log message from file {{Number 1 1000}}"
    weight: 1
seed: 12345
`

	tmpfile, err := ioutil.TempFile("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpfile.Name(), func() { os.Remove(tmpfile.Name()) }
}

// TestNewFromFile tests creating a generator from a config file
func TestNewFromFile(t *testing.T) {
	// Create temporary config file
	configFile, cleanup := createTempConfigFile(t)
	defer cleanup()

	// Create generator from file
	gen, err := genlog.NewFromFile(configFile)
	if err != nil {
		t.Fatalf("Failed to create generator from file: %v", err)
	}
	if gen == nil {
		t.Fatalf("Generator created from file is nil")
	}

	// Test generating a log line
	line, err := gen.GenerateLogLine()
	if err != nil {
		t.Fatalf("Failed to generate log line: %v", err)
	}

	// Verify log line follows expected pattern
	pattern := `^Test log message from file \d+$`
	matched, err := regexp.MatchString(pattern, line)
	if err != nil {
		t.Fatalf("Error matching pattern: %v", err)
	}
	if !matched {
		t.Errorf("Generated log line doesn't match expected pattern. Got: %s", line)
	}
}
