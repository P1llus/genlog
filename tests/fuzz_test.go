package genlog_test

import (
	"math/rand"
	"testing"

	"github.com/P1llus/genlog"
	"github.com/P1llus/genlog/pkg/config"
)

// FuzzGenerateWithRandomConfig tests the generator with randomly generated configurations
func FuzzGenerateWithRandomConfig(f *testing.F) {
	// Add some seed values
	f.Add(int64(42), 5, 3)
	f.Add(int64(123), 10, 5)
	f.Add(int64(0), 1, 0)

	f.Fuzz(func(t *testing.T, seed int64, templateCount int, customTypeCount int) {
		// Limit to reasonable sizes to avoid excessive resource usage
		if templateCount > 20 {
			templateCount = 20
		}
		if templateCount < 1 {
			templateCount = 1
		}
		if customTypeCount > 10 {
			customTypeCount = 10
		}
		if customTypeCount < 0 {
			customTypeCount = 0
		}

		// Create random seed if negative
		if seed < 0 {
			seed = -seed
		}

		// Create a random configuration
		cfg := createRandomConfig(uint64(seed), templateCount, customTypeCount)

		// Create generator
		gen := genlog.NewFromConfig(cfg)

		// Generate a log line
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			// It's okay to have errors for some extreme cases
			t.Logf("Error generating log line: %v", err)
			return
		}

		// Basic validation - just make sure we got a non-empty string
		if logLine == "" {
			t.Errorf("Generated empty log line")
		}
	})
}

// Helper function to create a random configuration
func createRandomConfig(seed uint64, templateCount, customTypeCount int) *config.Config {
	r := rand.New(rand.NewSource(int64(seed)))

	// Create templates
	templates := make([]config.LogTemplate, 0, templateCount)
	for i := 0; i < templateCount; i++ {
		template := createRandomTemplate(r)
		weight := r.Intn(10) + 1 // Weight between 1 and 10
		templates = append(templates, config.LogTemplate{
			Template: template,
			Weight:   weight,
		})
	}

	// Create custom types
	customTypes := make(map[string][]string)
	if customTypeCount > 0 {
		possibleTypes := []string{"level", "service", "component", "message", "username", "endpoint"}
		for i := 0; i < customTypeCount && i < len(possibleTypes); i++ {
			typeName := possibleTypes[i]
			valueCount := r.Intn(5) + 1 // 1 to 5 values per type

			values := make([]string, 0, valueCount)
			for j := 0; j < valueCount; j++ {
				values = append(values, randomString(r, 3, 10))
			}

			customTypes[typeName] = values
		}
	}

	return &config.Config{
		Templates:   templates,
		CustomTypes: customTypes,
		Seed:        seed,
	}
}

// Helper function to create a random template string
func createRandomTemplate(r *rand.Rand) string {
	// List of possible template parts
	parts := []string{
		"{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}}",
		"[{{level}}]",
		"{{service}}",
		"User: {{username}}",
		"IP: {{IPv4Address}}",
		"Port: {{Number 1000 9999}}",
		"Status: {{HTTPStatusCode}}",
		"{{message}}",
	}

	// Randomly select 2-5 parts to include
	count := r.Intn(4) + 2
	if count > len(parts) {
		count = len(parts)
	}

	template := ""
	for i := 0; i < count; i++ {
		idx := r.Intn(len(parts))
		if i > 0 {
			template += " "
		}
		template += parts[idx]
	}

	return template
}

// Helper function to generate a random string of given length range
func randomString(r *rand.Rand, minLen, maxLen int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := r.Intn(maxLen-minLen+1) + minLen

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[r.Intn(len(chars))]
	}

	return string(result)
}

// FuzzGenerateLogLine tests the GenerateLogLine function with fixed but varied configurations
func FuzzGenerateLogLine(f *testing.F) {
	// Add some seed corpus
	f.Add(uint64(1), "{{FormattedDate \"2006-01-02\"}}")
	f.Add(uint64(2), "{{IPv4Address}}")
	f.Add(uint64(3), "{{Number 1 100}}")
	f.Add(uint64(4), "{{level}}")

	f.Fuzz(func(t *testing.T, seed uint64, templateStr string) {
		// Skip empty templates
		if templateStr == "" {
			return
		}

		// Create a configuration with the provided template
		cfg := &config.Config{
			Templates: []config.LogTemplate{
				{
					Template: templateStr,
					Weight:   1,
				},
			},
			CustomTypes: map[string][]string{
				"level":    {"INFO", "WARN", "ERROR"},
				"message":  {"Test message"},
				"username": {"user1", "admin"},
			},
			Seed: seed,
		}

		// Create generator
		gen := genlog.NewFromConfig(cfg)

		// Try to generate a log line
		_, err := gen.GenerateLogLine()
		// We don't necessarily expect success for all random inputs,
		// but we want to make sure we don't panic
		if err != nil {
			t.Logf("Error generating log line for template '%s': %v", templateStr, err)
		}
	})
}
