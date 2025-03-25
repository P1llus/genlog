package generator

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/P1llus/genlog/pkg/config"
	"github.com/P1llus/genlog/pkg/output"
)

func TestNewGenerator(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "test template",
				Weight:   1,
			},
		},
		Outputs: []config.OutputConfig{
			{
				Type:    config.OutputTypeFile,
				Workers: 1,
				Config: map[string]interface{}{
					"filename": filepath.Join(tmpDir, "test.log"),
				},
			},
		},
	}

	gen, err := NewGenerator(cfg, 10)
	if err != nil {
		t.Fatalf("NewGenerator failed: %v", err)
	}
	if gen == nil {
		t.Fatal("Expected generator to be non-nil")
	}
	if gen.maxCount != 10 {
		t.Errorf("Expected maxCount 10, got %d", gen.maxCount)
	}
}

func TestGenerateLogLine(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] {{message}}",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"message": {"test message"},
		},
		Outputs: []config.OutputConfig{
			{
				Type:    config.OutputTypeFile,
				Workers: 1,
				Config: map[string]interface{}{
					"filename": filepath.Join(tmpDir, "test.log"),
				},
			},
		},
		Seed: 12345, // Use fixed seed for reproducibility
	}

	gen, err := NewGenerator(cfg, 1)
	if err != nil {
		t.Fatalf("NewGenerator failed: %v", err)
	}

	// Generate a log line
	logLine, err := gen.GenerateLogLine()
	if err != nil {
		t.Fatalf("GenerateLogLine failed: %v", err)
	}
	if logLine == "" {
		t.Error("Expected non-empty log line")
	}

	// Verify the log line format
	if len(logLine) < 30 { // Minimum expected length for timestamp + message
		t.Errorf("Log line too short: %s", logLine)
	}
}

func TestSelectWeightedTemplate(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "template 1",
				Weight:   10,
			},
			{
				Template: "template 2",
				Weight:   5,
			},
		},
		Outputs: []config.OutputConfig{
			{
				Type:    config.OutputTypeFile,
				Workers: 1,
				Config: map[string]interface{}{
					"filename": filepath.Join(tmpDir, "test.log"),
				},
			},
		},
		Seed: 12345, // Use fixed seed for reproducibility
	}

	gen, err := NewGenerator(cfg, 1)
	if err != nil {
		t.Fatalf("NewGenerator failed: %v", err)
	}

	// Test template selection multiple times
	selectedTemplates := make(map[int]int)
	for i := 0; i < 1000; i++ {
		idx := gen.selectWeightedTemplate()
		selectedTemplates[idx]++
	}

	// Verify that the first template (weight 10) is selected roughly twice as often
	// as the second template (weight 5)
	ratio := float64(selectedTemplates[0]) / float64(selectedTemplates[1])
	if ratio < 1.5 || ratio > 2.5 {
		t.Errorf("Template selection ratio outside expected range: %f", ratio)
	}
}

func TestStartAndStop(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "test template",
				Weight:   1,
			},
		},
		Outputs: []config.OutputConfig{
			{
				Type:    config.OutputTypeFile,
				Workers: 1,
				Config: map[string]interface{}{
					"filename": filepath.Join(tmpDir, "test.log"),
				},
			},
		},
	}

	gen, err := NewGenerator(cfg, 5)
	if err != nil {
		t.Fatalf("NewGenerator failed: %v", err)
	}

	// Start the generator
	gen.Start()

	// Wait for completion with timeout
	select {
	case <-gen.Done():
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Generator did not complete within timeout")
	}

	// Stop the generator
	if err := gen.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestCreateFuncMap(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		CustomTypes: map[string][]string{
			"test_type": {"value1", "value2"},
		},
		Templates: []config.LogTemplate{
			{
				Template: "{{test_type}}",
				Weight:   1,
			},
		},
		Outputs: []config.OutputConfig{
			{
				Type:    config.OutputTypeFile,
				Workers: 1,
				Config: map[string]interface{}{
					"filename": filepath.Join(tmpDir, "test.log"),
				},
			},
		},
	}

	gen, err := NewGenerator(cfg, 1)
	if err != nil {
		t.Fatalf("NewGenerator failed: %v", err)
	}

	// Test custom type function
	if fn, ok := gen.funcMap["test_type"].(func() string); ok {
		value := fn()
		if value != "value1" && value != "value2" {
			t.Errorf("Unexpected value from custom type function: %s", value)
		}
	} else {
		t.Error("Custom type function not found in funcMap")
	}

	// Test FormattedDate function
	if fn, ok := gen.funcMap["FormattedDate"].(func(string) string); ok {
		date := fn("2006-01-02")
		if date == "" {
			t.Error("FormattedDate returned empty string")
		}
	} else {
		t.Error("FormattedDate function not found in funcMap")
	}
}

// BenchmarkGeneratorFileOutput measures the performance of generating and writing logs using a Worker
func BenchmarkGeneratorFileOutput(b *testing.B) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create output config
	cfg := config.OutputConfig{
		Type:    config.OutputTypeFile,
		Workers: 1,
		Config: map[string]interface{}{
			"filename": filepath.Join(tmpDir, "benchmark.log"),
		},
	}

	// Create output
	out, err := output.NewOutput(cfg, 0)
	if err != nil {
		b.Fatalf("NewOutput failed: %v", err)
	}
	defer out.Close()

	// Create generator config
	genConfig := &config.Config{
		Seed: 12345,
		Templates: []config.LogTemplate{
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
		Outputs: []config.OutputConfig{
			{
				Type:    config.OutputTypeFile,
				Workers: 1,
				Config: map[string]interface{}{
					"filename": filepath.Join(tmpDir, "benchmark.log"),
				},
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

	// Create generator
	gen, err := NewGenerator(genConfig, 0)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	// Create worker
	stopChan := make(chan struct{})
	worker := output.NewWorker(out, gen, 100, 0, stopChan)

	// Reset the benchmark timer before the actual benchmark
	b.ResetTimer()

	// Start worker
	go worker.Start()

	// Wait for completion
	<-time.After(time.Duration(b.N) * time.Microsecond)
	close(stopChan)

	time.Sleep(100 * time.Millisecond)
}
