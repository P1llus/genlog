package genlog_test

import (
	"os"
	"path/filepath"
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
		Outputs: []genlog.OutputConfig{
			{
				Type: genlog.OutputTypeFile,
				Config: map[string]interface{}{
					"filename": "test.log",
				},
				Workers: 1,
			},
		},
		CustomTypes: map[string][]string{
			"username": {"user1", "user2", "user3"},
		},
		Seed: 12345,
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
		Outputs: []genlog.OutputConfig{
			{
				Type: genlog.OutputTypeFile,
				Config: map[string]interface{}{
					"filename": "test.log",
				},
				Workers: 1,
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

// testConfig represents a fixed configuration for benchmarks
var testConfig = &genlog.Config{
	Seed: 12345,
	Templates: []genlog.LogTemplate{
		{
			Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] {{username}} - {{message}}",
			Weight:   5,
		},
		{
			Template: "{{FormattedDate \"Jan 2 15:04:05\"}} {{level}} [{{service}}] {{IPv4Address}} {{username}}: {{message}}",
			Weight:   3,
		},
		{
			Template: "{\"timestamp\":\"{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}}\",\"level\":\"{{level}}\",\"service\":\"{{service}}\",\"message\":\"{{message}}\",\"user\":\"{{username}}\",\"ip\":\"{{IPv4Address}}\"}",
			Weight:   2,
		},
	},
	CustomTypes: map[string][]string{
		"level": {
			"INFO",
			"WARNING",
			"ERROR",
			"DEBUG",
			"TRACE",
		},
		"service": {
			"API",
			"AUTH",
			"DATABASE",
			"CACHE",
			"FRONTEND",
		},
		"username": {
			"admin",
			"system",
			"app",
			"service_account",
			"anonymous",
		},
		"message": {
			"User authenticated successfully",
			"Failed login attempt - invalid credentials",
			"Permission denied to resource",
			"Resource accessed successfully",
			"API rate limit exceeded",
			"Database connection timeout",
			"Cache invalidation completed",
			"Request processed in 235ms",
		},
	},
}

// BenchmarkGenerateLogs measures the performance of generating logs
func BenchmarkGenerateLogs(b *testing.B) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a copy of the test config with the temporary directory
	cfg := *testConfig
	cfg.Outputs = []genlog.OutputConfig{
		{
			Type:    genlog.OutputTypeFile,
			Workers: 1,
			Config: map[string]interface{}{
				"filename": filepath.Join(tmpDir, "benchmark.log"),
			},
		},
	}

	// Create generator from test config
	gen, err := genlog.NewFromConfig(&cfg, 100)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	// Reset the benchmark timer before the actual benchmark
	b.ResetTimer()

	// Generate logs
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateLogLine()
		if err != nil {
			b.Fatalf("Failed to generate log line: %v", err)
		}
	}
}
