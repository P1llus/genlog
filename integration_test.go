package genlog_test

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/P1llus/genlog"
)

// TestLibraryIntegration tests using genlog as a library
func TestLibraryIntegration(t *testing.T) {
	// Create a temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "genlog-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "test-output.log")
	expectedLines := int64(10000)

	// Create configuration
	cfg := &genlog.Config{
		Templates: []genlog.Template{
			{
				Template: "{{timestamp}} [{{level}}] {{message}}",
				Weight:   10,
			},
		},
		CustomTypes: map[string][]string{
			"level":   {"INFO", "WARN", "ERROR"},
			"message": {"Test message 1", "Test message 2"},
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
			MaxCount: expectedLines,
		},
		Seed: 12345, // Use seed for reproducibility
	}

	// Create and run generator
	gen, err := genlog.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Start(); err != nil {
		t.Fatalf("Failed to start generator: %v", err)
	}

	// Wait for completion with timeout
	select {
	case <-gen.Done():
		// Success
	case <-time.After(30 * time.Second):
		t.Fatal("Generation timed out after 30 seconds")
	}

	if err := gen.Stop(); err != nil {
		t.Fatalf("Failed to stop generator: %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", outputFile)
	}

	// Count lines in output file
	lineCount, err := countLines(outputFile)
	if err != nil {
		t.Fatalf("Failed to count lines: %v", err)
	}

	if lineCount != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, lineCount)
	}

	// Verify stats match
	stats := gen.Stats().Snapshot()
	if stats.Generated != expectedLines {
		t.Errorf("Stats: Expected %d generated, got %d", expectedLines, stats.Generated)
	}
	if stats.Written != expectedLines {
		t.Errorf("Stats: Expected %d written, got %d", expectedLines, stats.Written)
	}

	t.Logf("Successfully generated %d logs in %v (%.0f logs/sec)",
		lineCount, stats.Duration, stats.LogsPerSecond)
}

// TestCLIIntegration tests the CLI tool
func TestCLIIntegration(t *testing.T) {
	// Build the CLI binary
	tmpDir, err := os.MkdirTemp("", "genlog-cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath := filepath.Join(tmpDir, "genlog")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/genlog")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build CLI: %v\nOutput: %s", err, output)
	}

	// Test --version flag
	t.Run("version_flag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "-version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to run --version: %v\nOutput: %s", err, output)
		}
		if len(output) == 0 {
			t.Error("Expected version output, got empty")
		}
		t.Logf("Version output: %s", output)
	})

	// Test --init flag
	t.Run("init_flag", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "test-config.yaml")
		cmd := exec.Command(binaryPath, "-init", configPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to generate config: %v\nOutput: %s", err, output)
		}

		// Verify config file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatalf("Config file was not created: %s", configPath)
		}

		t.Logf("Init output: %s", output)

		// Test that init fails if file already exists
		cmd = exec.Command(binaryPath, "-init", configPath)
		output, err = cmd.CombinedOutput()
		if err == nil {
			t.Error("Expected error when file already exists, got none")
		}
		if !strings.Contains(string(output), "already exists") {
			t.Errorf("Expected 'already exists' error message, got: %s", output)
		}
		t.Logf("Expected error output: %s", output)
	})

	// Test log generation
	t.Run("generate_logs", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "gen-config.yaml")
		outputPath := filepath.Join(tmpDir, "cli-output.log")
		expectedLines := 10000

		// Create a simple config
		configContent := `templates:
  - template: '{{timestamp}} [{{level}}] {{message}}'
    weight: 1

custom_types:
  level: [INFO, WARN, ERROR]
  message: ["Test message"]

outputs:
  - type: file
    config:
      path: "` + outputPath + `"

limits:
  max_count: 10000
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Run genlog
		cmd := exec.Command(binaryPath, "-config", configPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to generate logs: %v\nOutput: %s", err, output)
		}

		// Verify output file exists
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Fatalf("Output file was not created: %s", outputPath)
		}

		// Count lines
		lineCount, err := countLines(outputPath)
		if err != nil {
			t.Fatalf("Failed to count lines: %v", err)
		}

		if lineCount != int64(expectedLines) {
			t.Errorf("Expected %d lines, got %d", expectedLines, lineCount)
		}

		t.Logf("CLI generated %d logs successfully", lineCount)
	})
}

// TestMultipleOutputs tests generation to multiple outputs simultaneously.
// Note: With the current architecture, logs are distributed among outputs,
// not replicated to each output. This test verifies that the total across
// all outputs matches the expected count.
func TestMultipleOutputs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "genlog-multi-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output1 := filepath.Join(tmpDir, "output1.log")
	output2 := filepath.Join(tmpDir, "output2.log")
	expectedTotal := int64(5000)

	cfg := &genlog.Config{
		Templates: []genlog.Template{
			{
				Template: "{{timestamp}} [{{level}}] {{message}}",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"level":   {"INFO"},
			"message": {"Test"},
		},
		Outputs: []genlog.OutputConfig{
			{
				Type: "file",
				Config: map[string]interface{}{
					"path": output1,
				},
			},
			{
				Type: "file",
				Config: map[string]interface{}{
					"path": output2,
				},
			},
		},
		Limits: genlog.Limits{
			MaxCount: expectedTotal,
		},
	}

	gen, err := genlog.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	gen.Start()
	gen.Wait()
	gen.Stop()

	// Count lines in both outputs
	count1, err := countLines(output1)
	if err != nil {
		t.Fatalf("Failed to count lines in output1: %v", err)
	}

	count2, err := countLines(output2)
	if err != nil {
		t.Fatalf("Failed to count lines in output2: %v", err)
	}

	total := count1 + count2

	// Verify that the total matches expected count
	// Logs are distributed between outputs based on which consumer grabs them
	if total != expectedTotal {
		t.Errorf("Expected total %d lines across outputs, got %d (output1: %d, output2: %d)",
			expectedTotal, total, count1, count2)
	}

	t.Logf("Successfully distributed %d logs across 2 outputs (output1: %d, output2: %d)",
		total, count1, count2)
}

// countLines counts the number of lines in a file
func countLines(filepath string) (int64, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var count int64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}
