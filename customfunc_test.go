package genlog_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/P1llus/genlog"
)

// TestCustomFunctions tests the ability to register custom functions
func TestCustomFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "genlog-customfunc-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "custom-func-output.log")

	// Create a configuration with custom functions
	cfg := &genlog.Config{
		Templates: []genlog.Template{
			{
				// Use multiple custom functions in one template with string arguments
				Template: "{{Uppercase \"hello\"}} {{Repeat \"*\" \"3\"}} {{Join \"-\" \"a\" \"b\" \"c\"}}",
				Weight:   1,
			},
			{
				// Mix custom functions with built-in placeholders
				// Note: Custom functions receive evaluated placeholder values
				Template: "{{ipv4}} {{level}} - {{Reverse \"world\"}}",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"level": {"info", "warn", "error"},
		},
		CustomFunctions: map[string]genlog.TemplateFunc{
			// Simple uppercase function
			"Uppercase": func(args ...string) string {
				if len(args) == 0 {
					return ""
				}
				return strings.ToUpper(args[0])
			},
			// Repeat a string n times
			"Repeat": func(args ...string) string {
				if len(args) < 2 {
					return ""
				}
				str := args[0]
				count := 1
				fmt.Sscanf(args[1], "%d", &count)
				return strings.Repeat(str, count)
			},
			// Join strings with a separator
			"Join": func(args ...string) string {
				if len(args) < 2 {
					return ""
				}
				sep := args[0]
				return strings.Join(args[1:], sep)
			},
			// Reverse a string
			"Reverse": func(args ...string) string {
				if len(args) == 0 {
					return ""
				}
				runes := []rune(args[0])
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return string(runes)
			},
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
			MaxCount: 10,
		},
		Seed: 12345,
	}

	// Create and run generator
	gen, err := genlog.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Start(); err != nil {
		t.Fatalf("Failed to start generator: %v", err)
	}

	gen.Wait()

	if err := gen.Stop(); err != nil {
		t.Fatalf("Failed to stop generator: %v", err)
	}

	// Read and verify output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	t.Logf("Generated output:\n%s", output)

	// Verify custom functions worked
	if !strings.Contains(output, "HELLO") {
		t.Error("Expected 'HELLO' from Uppercase function")
	}
	if !strings.Contains(output, "***") {
		t.Error("Expected '***' from Repeat function")
	}
	if !strings.Contains(output, "a-b-c") {
		t.Error("Expected 'a-b-c' from Join function")
	}
	if !strings.Contains(output, "dlrow") {
		t.Error("Expected 'dlrow' (reversed 'world') from Reverse function")
	}
	// Also verify regular placeholders still work
	if !strings.Contains(output, "info") && !strings.Contains(output, "warn") && !strings.Contains(output, "error") {
		t.Error("Expected level placeholder to work alongside custom functions")
	}

	t.Log("All custom functions worked correctly!")
}

// TestCustomFunctionOverridesBuiltin tests that custom functions take priority over built-ins
func TestCustomFunctionOverridesBuiltin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "genlog-override-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "override-output.log")

	cfg := &genlog.Config{
		Templates: []genlog.Template{
			{
				// Override the built-in Number function
				Template: "Custom Number: {{Number \"5\" \"10\"}}",
				Weight:   1,
			},
		},
		CustomFunctions: map[string]genlog.TemplateFunc{
			// Override Number to always return "CUSTOM"
			"Number": func(args ...string) string {
				return "CUSTOM"
			},
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
			MaxCount: 5,
		},
	}

	gen, err := genlog.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	gen.Start()
	gen.Wait()
	gen.Stop()

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	t.Logf("Generated output:\n%s", output)

	// Verify custom function overrode the built-in
	if !strings.Contains(output, "CUSTOM") {
		t.Error("Expected 'CUSTOM' from overridden Number function")
	}

	// Make sure it's not using the built-in number generation
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Custom Number:") && !strings.Contains(line, "CUSTOM") {
			t.Errorf("Found line without CUSTOM: %s", line)
		}
	}

	t.Log("Custom function successfully overrode built-in!")
}
