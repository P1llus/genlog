package genlog_test

import (
	"os"
	"testing"

	"github.com/P1llus/genlog"
)

// BenchmarkGenerateLogLine benchmarks the performance of generating a single log line
func BenchmarkGenerateLogLine(b *testing.B) {
	// Create a simple configuration
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] {{message}}",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"level":   {"INFO", "WARN", "ERROR", "DEBUG"},
			"message": {"System starting up", "Connection established", "User authentication successful"},
		},
		Seed: 12345,
	}

	gen := genlog.NewFromConfig(cfg)

	// Reset timer before the loop
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateLogLine()
		if err != nil {
			b.Fatalf("Failed to generate log line: %v", err)
		}
	}
}

// BenchmarkGenerateLogLineComplex benchmarks generating a log line with a more complex template
func BenchmarkGenerateLogLineComplex(b *testing.B) {
	// Create a more complex configuration
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: `{{FormattedDate "2006-01-02T15:04:05.000Z07:00"}} [{{level}}] {{service}} - User {{Name}} ({{Email}}) connected from {{IPv4Address}}:{{Number 1000 65535}} using {{Browser}} {{UserAgent}}`,
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"level":   {"INFO", "WARN", "ERROR", "DEBUG"},
			"service": {"Auth", "API", "Web", "Mobile"},
			"Browser": {"Chrome", "Firefox", "Safari", "Edge"},
		},
		Seed: 12345,
	}

	gen := genlog.NewFromConfig(cfg)

	// Reset timer before the loop
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateLogLine()
		if err != nil {
			b.Fatalf("Failed to generate log line: %v", err)
		}
	}
}

// BenchmarkGenerateLogs benchmarks generating multiple logs to a file
func BenchmarkGenerateLogs(b *testing.B) {
	// Create a simple configuration
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] {{message}}",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"level":   {"INFO", "WARN", "ERROR", "DEBUG"},
			"message": {"System starting up", "Connection established", "User authentication successful"},
		},
		Seed: 12345,
	}

	gen := genlog.NewFromConfig(cfg)
	outputFile := "benchmark_output.log"

	// Clean up file between runs and after test
	defer os.Remove(outputFile)

	// Reset timer before the loop
	b.ResetTimer()

	// Set a fixed number of log lines per benchmark iteration
	const logsPerIteration = 100

	for i := 0; i < b.N; i++ {
		// Stop timer during cleanup
		b.StopTimer()
		os.Remove(outputFile)
		b.StartTimer()

		err := gen.GenerateLogs(outputFile, logsPerIteration)
		if err != nil {
			b.Fatalf("Failed to generate logs: %v", err)
		}
	}
}

// BenchmarkMultipleTemplates benchmarks generating logs with multiple template options
func BenchmarkMultipleTemplates(b *testing.B) {
	// Create configuration with multiple templates
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] Basic log message",
				Weight:   10,
			},
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] User {{username}} logged in from {{IPv4Address}}",
				Weight:   5,
			},
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] API request to {{endpoint}} completed in {{Number 1 500}}ms with status {{HTTPStatusCode}}",
				Weight:   3,
			},
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] Database query \"{{query}}\" took {{Number 1 1000}}ms to execute",
				Weight:   2,
			},
		},
		CustomTypes: map[string][]string{
			"level":    {"INFO", "WARN", "ERROR", "DEBUG"},
			"username": {"admin", "user1", "guest", "system"},
			"endpoint": {"/api/users", "/api/products", "/api/auth", "/api/settings"},
			"query":    {"SELECT * FROM users", "UPDATE products SET price = 10", "INSERT INTO logs VALUES (...)"},
		},
		Seed: 12345,
	}

	gen := genlog.NewFromConfig(cfg)

	// Reset timer before the loop
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateLogLine()
		if err != nil {
			b.Fatalf("Failed to generate log line: %v", err)
		}
	}
}
