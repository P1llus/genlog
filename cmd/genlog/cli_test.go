package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIBasicUsage tests the basic functionality of the CLI
func TestCLIBasicUsage(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "genlog-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a config file
	configFile := filepath.Join(tempDir, "config.yaml")
	err = createTestConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Set output file
	outputFile := filepath.Join(tempDir, "output.log")

	// Build the CLI
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "genlog"))
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v\nOutput: %s", err, buildOutput)
	}

	// Run the CLI
	cmd := exec.Command(
		filepath.Join(tempDir, "genlog"),
		"--config="+configFile,
		"--output="+outputFile,
		"--count=10",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run CLI: %v\nOutput: %s", err, output)
	}

	// Verify output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Read output file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Count lines
	lines := strings.Split(string(content), "\n")
	// Remove last empty line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Verify line count
	if len(lines) != 10 {
		t.Errorf("Expected 10 log lines, got %d", len(lines))
	}

	// Log output for debugging
	t.Logf("CLI output: %s", output)
}

// TestCLIInvalidConfig tests the CLI behavior with an invalid config
func TestCLIInvalidConfig(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "genlog-cli-invalid-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an invalid config file
	configFile := filepath.Join(tempDir, "invalid_config.yaml")
	err = os.WriteFile(configFile, []byte("invalid: yaml: content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}

	// Build the CLI
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "genlog"))
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v\nOutput: %s", err, buildOutput)
	}

	// Run the CLI with invalid config
	cmd := exec.Command(
		filepath.Join(tempDir, "genlog"),
		"--config="+configFile,
		"--output="+filepath.Join(tempDir, "output.log"),
		"--count=10",
	)

	output, err := cmd.CombinedOutput()

	// The command should fail with an error
	if err == nil {
		t.Errorf("CLI did not fail with invalid config")
	}

	// Error message should indicate config problem
	if !strings.Contains(string(output), "config") && !strings.Contains(string(output), "error") {
		t.Errorf("Error message doesn't mention config problem: %s", output)
	}
}

// Helper function to create a valid test config file
func createTestConfig(filePath string) error {
	content := `
templates:
  - template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} Test log message {{Number 1 100}}"
    weight: 1
seed: 12345
`
	return os.WriteFile(filePath, []byte(content), 0o644)
}
