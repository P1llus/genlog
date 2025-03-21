package genlog_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/P1llus/genlog"
)

// TestIntegrationComplete tests a complete workflow from configuration to log generation
func TestIntegrationComplete(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "genlog-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test configuration
	configFile := filepath.Join(tempDir, "test_config.yaml")
	err = createComplexConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create generator from config file
	gen, err := genlog.NewFromFile(configFile)
	if err != nil {
		t.Fatalf("Failed to create generator from config file: %v", err)
	}

	// Generate logs
	outputFile := filepath.Join(tempDir, "output.log")
	logCount := 100
	err = gen.GenerateLogs(outputFile, logCount)
	if err != nil {
		t.Fatalf("Failed to generate logs: %v", err)
	}

	// Verify log file exists and has correct number of lines
	verifyLogFile(t, outputFile, logCount)

	// Verify log patterns match expected templates
	verifyLogPatterns(t, outputFile)
}

// Creates a complex configuration file for testing
func createComplexConfig(filePath string) error {
	content := `
templates:
  - template: '{{FormattedDate "2006-01-02T15:04:05.000Z07:00"}} [{{level}}] {{ServiceName}} - User {{UserName}} accessed resource from {{IPv4Address}}'
    weight: 10
  - template: '{{FormattedDate "2006-01-02T15:04:05.000Z07:00"}} [{{level}}] {{ServiceName}} - Failed login attempt for user {{UserName}} from {{IPv4Address}}, reason: {{FailReason}}'
    weight: 5
  - template: '{{FormattedDate "2006-01-02T15:04:05.000Z07:00"}} [{{level}}] {{ServiceName}} - System event: {{SysEvent}} (id: {{Number 10000 99999}})'
    weight: 3

custom_types:
  level:
    - INFO
    - WARN
    - ERROR
    - DEBUG
  ServiceName:
    - Authentication
    - API
    - Database
    - Frontend
    - Backend
  UserName:
    - alice
    - bob
    - charlie
    - admin
    - system
  FailReason:
    - "Invalid credentials"
    - "Account locked"
    - "Password expired"
    - "Two-factor authentication failed"
  SysEvent:
    - "Resource limit reached"
    - "Service restarted"
    - "Configuration updated"
    - "Backup completed"
    - "Database connection established"

seed: 12345
`
	return os.WriteFile(filePath, []byte(content), 0o644)
}

// Verifies that the log file exists and has the expected number of lines
func verifyLogFile(t *testing.T, filePath string, expectedCount int) {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	if lineCount != expectedCount {
		t.Errorf("Expected %d log lines, got %d", expectedCount, lineCount)
	}
}

// Verifies that log lines match expected patterns based on the templates
func verifyLogPatterns(t *testing.T, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	// Define patterns to match based on the templates
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z \[(INFO|WARN|ERROR|DEBUG)\] (Authentication|API|Database|Frontend|Backend) - User (alice|bob|charlie|admin|system) accessed resource from \d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`),
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z \[(INFO|WARN|ERROR|DEBUG)\] (Authentication|API|Database|Frontend|Backend) - Failed login attempt for user (alice|bob|charlie|admin|system) from \d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}, reason: (Invalid credentials|Account locked|Password expired|Two-factor authentication failed)$`),
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z \[(INFO|WARN|ERROR|DEBUG)\] (Authentication|API|Database|Frontend|Backend) - System event: (Resource limit reached|Service restarted|Configuration updated|Backup completed|Database connection established) \(id: \d{5}\)$`),
	}

	scanner := bufio.NewScanner(file)
	lineCount := 0
	matchedLines := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Check if line matches any of our expected patterns
		matched := false
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				matched = true
				matchedLines++
				break
			}
		}

		if !matched {
			t.Errorf("Line %d does not match any expected pattern: %s", lineCount, line)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}

	t.Logf("Successfully matched %d/%d log lines to expected patterns", matchedLines, lineCount)
}

// TestIntegrationWeightedDistribution tests that template weights affect distribution
func TestIntegrationWeightedDistribution(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a config with two templates with different weights
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "Template A",
				Weight:   80, // 80% weight
			},
			{
				Template: "Template B",
				Weight:   20, // 20% weight
			},
		},
		Seed: 12345,
	}

	gen := genlog.NewFromConfig(cfg)

	// Generate a large number of log lines to test distribution
	tempFile := "temp_weight_test.log"
	defer os.Remove(tempFile)

	logCount := 1000
	err := gen.GenerateLogs(tempFile, logCount)
	if err != nil {
		t.Fatalf("Failed to generate logs: %v", err)
	}

	// Count occurrences of each template
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	countA := 0
	countB := 0
	for _, line := range lines {
		switch line {
		case "Template A":
			countA++
		case "Template B":
			countB++
		default:
			t.Errorf("Unexpected log line: %s", line)
		}
	}

	// Check that distribution is roughly as expected (with some margin of error)
	// For 80/20 weight with 1000 logs, we expect approximately 800 A and 200 B
	expectedA := int(float64(logCount) * 0.8)
	expectedB := int(float64(logCount) * 0.2)

	// Allow for a 10% margin of error due to randomness
	margin := int(float64(logCount) * 0.1)

	t.Logf("Template A count: %d (expected ~%d)", countA, expectedA)
	t.Logf("Template B count: %d (expected ~%d)", countB, expectedB)

	if countA < expectedA-margin || countA > expectedA+margin {
		t.Errorf("Template A distribution outside expected range. Got: %d, Expected: %d±%d", countA, expectedA, margin)
	}

	if countB < expectedB-margin || countB > expectedB+margin {
		t.Errorf("Template B distribution outside expected range. Got: %d, Expected: %d±%d", countB, expectedB, margin)
	}
}
