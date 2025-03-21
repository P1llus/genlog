package generator_test

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/P1llus/genlog/pkg/config"
	"github.com/P1llus/genlog/pkg/generator"
)

func TestNewGenerator(t *testing.T) {
	// Create a simple config for testing
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "Static log line",
				Weight:   1,
			},
		},
		Seed: 12345,
	}

	// Create a new generator
	gen := generator.NewGenerator(cfg)
	if gen == nil {
		t.Fatalf("Failed to create generator")
	}

	// Test generating a log line
	logLine, err := gen.GenerateLogLine()
	if err != nil {
		t.Fatalf("Failed to generate log line: %v", err)
	}
	if logLine != "Static log line" {
		t.Errorf("Generated log line doesn't match template. Got: %s, Expected: Static log line", logLine)
	}
}

func TestGenerateLogLine(t *testing.T) {
	// Test with built-in functions
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "IP: {{IPv4Address}} Number: {{Number 1 100}}",
				Weight:   1,
			},
		},
		Seed: 12345,
	}

	gen := generator.NewGenerator(cfg)

	// Generate log line
	logLine, err := gen.GenerateLogLine()
	if err != nil {
		t.Fatalf("Failed to generate log line: %v", err)
	}

	// Validate log line format
	ipPattern := regexp.MustCompile(`IP: \d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3} Number: \d{1,3}`)
	if !ipPattern.MatchString(logLine) {
		t.Errorf("Generated log line doesn't match expected pattern. Got: %s", logLine)
	}
}

func TestGenerateLogLineWithCustomType(t *testing.T) {
	// Test with custom types
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "{{level}} - {{message}}",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"level":   {"INFO", "WARN", "ERROR"},
			"message": {"Test message 1", "Test message 2"},
		},
		Seed: 12345,
	}

	gen := generator.NewGenerator(cfg)

	// Generate log line
	logLine, err := gen.GenerateLogLine()
	if err != nil {
		t.Fatalf("Failed to generate log line: %v", err)
	}

	// Validate log line format
	pattern := regexp.MustCompile(`^(INFO|WARN|ERROR) - (Test message 1|Test message 2)$`)
	if !pattern.MatchString(logLine) {
		t.Errorf("Generated log line doesn't match expected pattern. Got: %s", logLine)
	}
}

func TestGenerateLogs(t *testing.T) {
	// Create a temp file for output
	outputFile := "test_generator_output.log"
	defer os.Remove(outputFile)

	// Create config
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "Log entry {{Number 1 1000}}",
				Weight:   1,
			},
		},
		Seed: 12345,
	}

	gen := generator.NewGenerator(cfg)

	// Generate logs
	count := 10
	err := gen.GenerateLogs(outputFile, count)
	if err != nil {
		t.Fatalf("Failed to generate logs: %v", err)
	}

	// Verify file content
	file, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		pattern := regexp.MustCompile(`^Log entry \d+$`)
		if !pattern.MatchString(line) {
			t.Errorf("Line %d doesn't match expected pattern. Got: %s", lineCount, line)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	if lineCount != count {
		t.Errorf("Expected %d log lines, got %d", count, lineCount)
	}
}

func TestSelectWeightedTemplate(t *testing.T) {
	// This test needs to test the template weight distribution
	// Create config with two templates, with vastly different weights
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "Template A",
				Weight:   90,
			},
			{
				Template: "Template B",
				Weight:   10,
			},
		},
		Seed: 12345,
	}

	gen := generator.NewGenerator(cfg)

	// Generate many log lines to test the distribution
	countA := 0
	countB := 0
	totalSamples := 1000

	for i := 0; i < totalSamples; i++ {
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			t.Fatalf("Failed to generate log line: %v", err)
		}

		switch logLine {
		case "Template A":
			countA++
		case "Template B":
			countB++
		default:
			t.Errorf("Unexpected log line: %s", logLine)
		}
	}

	// Check distribution is roughly as expected (allowing for some randomness)
	expectedA := totalSamples * 90 / 100 // 90%
	marginA := totalSamples * 5 / 100    // 5% margin of error

	t.Logf("Template A count: %d (expected ~%d)", countA, expectedA)
	t.Logf("Template B count: %d (expected ~%d)", countB, totalSamples-expectedA)

	if countA < expectedA-marginA || countA > expectedA+marginA {
		t.Errorf("Template A distribution outside expected range. Got: %d, Expected: %d±%d",
			countA, expectedA, marginA)
	}
}

func TestFormattedDate(t *testing.T) {
	// Test the FormattedDate function
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: `{{FormattedDate "2006-01-02"}}`,
				Weight:   1,
			},
		},
		Seed: 12345,
	}

	gen := generator.NewGenerator(cfg)

	// Generate log line
	logLine, err := gen.GenerateLogLine()
	if err != nil {
		t.Fatalf("Failed to generate log line: %v", err)
	}

	// Validate log line format (YYYY-MM-DD)
	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !datePattern.MatchString(logLine) {
		t.Errorf("Generated date doesn't match expected format. Got: %s", logLine)
	}
}

func TestZeroWeightTemplates(t *testing.T) {
	// Test that templates with zero weight are never selected
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "Template A",
				Weight:   1,
			},
			{
				Template: "Template B (should never appear)",
				Weight:   0,
			},
		},
		Seed: 12345,
	}

	gen := generator.NewGenerator(cfg)

	// Generate multiple log lines
	for i := 0; i < 100; i++ {
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			t.Fatalf("Failed to generate log line: %v", err)
		}

		// Zero-weight template should never be selected
		if logLine == "Template B (should never appear)" {
			t.Errorf("Zero-weight template was selected: %s", logLine)
		}
	}
}

func TestEmptyTemplates(t *testing.T) {
	// Test with an empty templates list
	cfg := &config.Config{
		Templates: []config.LogTemplate{},
		Seed:      12345,
	}

	gen := generator.NewGenerator(cfg)

	// Generate log line - this should return an empty string with no error
	// since there are no templates to select from
	logLine, err := gen.GenerateLogLine()

	// The behavior might depend on implementation - either expect an error
	// or an empty string
	if err == nil && logLine != "" {
		t.Errorf("Expected empty string or error with empty templates, got: %s", logLine)
	}
}

func TestMultipleTemplatesConsistency(t *testing.T) {
	// Test that with a fixed seed, template selection is consistent
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "Template 1",
				Weight:   1,
			},
			{
				Template: "Template 2",
				Weight:   1,
			},
			{
				Template: "Template 3",
				Weight:   1,
			},
		},
		Seed: 42, // Fixed seed
	}

	// Create first generator
	gen1 := generator.NewGenerator(cfg)

	// Generate a sequence of log lines
	sequence1 := make([]string, 10)
	for i := 0; i < 10; i++ {
		logLine, err := gen1.GenerateLogLine()
		if err != nil {
			t.Fatalf("Failed to generate log line: %v", err)
		}
		sequence1[i] = logLine
	}

	// Create a second generator with the same config and seed
	gen2 := generator.NewGenerator(cfg)

	// Generate another sequence of log lines
	sequence2 := make([]string, 10)
	for i := 0; i < 10; i++ {
		logLine, err := gen2.GenerateLogLine()
		if err != nil {
			t.Fatalf("Failed to generate log line: %v", err)
		}
		sequence2[i] = logLine
	}

	// The sequences should be identical with the same seed
	for i := 0; i < 10; i++ {
		if sequence1[i] != sequence2[i] {
			t.Errorf("Sequence mismatch at position %d: %s != %s",
				i, sequence1[i], sequence2[i])
		}
	}
}

func TestOverwriteExistingFile(t *testing.T) {
	// Create a file to be overwritten
	outputFile := "overwrite_test.log"
	initialContent := "This should be overwritten\n"

	err := os.WriteFile(outputFile, []byte(initialContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}
	defer os.Remove(outputFile)

	// Create config and generator
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "New log entry",
				Weight:   1,
			},
		},
		Seed: 12345,
	}

	gen := generator.NewGenerator(cfg)

	// Generate logs (should overwrite)
	count := 5
	err = gen.GenerateLogs(outputFile, count)
	if err != nil {
		t.Fatalf("Failed to generate logs: %v", err)
	}

	// Read file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify file was overwritten
	if string(content) == initialContent {
		t.Errorf("File was not overwritten")
	}

	// Count lines
	lines := strings.Split(string(content), "\n")
	// Remove empty line at end if exists
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) != count {
		t.Errorf("Expected %d lines, got %d", count, len(lines))
	}
}

func TestInvalidTemplateOutput(t *testing.T) {
	// Test with an invalid template
	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "Test {{InvalidFunction}}",
				Weight:   1,
			},
		},
		Seed: 12345,
	}

	gen := generator.NewGenerator(cfg)

	// Attempt to generate a log line with the invalid template
	// The expected behavior depends on the implementation:
	// - It might return an error
	// - It might substitute with an empty string or some placeholder
	_, err := gen.GenerateLogLine()
	// We're primarily checking that it doesn't panic
	if err != nil {
		t.Logf("Got expected error for invalid template: %v", err)
	}
}
