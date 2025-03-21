package config_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/P1llus/genlog/pkg/config"
)

func TestReadConfig(t *testing.T) {
	// Create a temporary config file
	content := `
templates:
  - template: "Test template 1 {{level}} {{service}}"
    weight: 10
  - template: "Test template 2"
    weight: 5
custom_types:
  level:
    - INFO
    - ERROR
  service:
    - API
    - DB
seed: 12345
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Read the config
	cfg, err := config.ReadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Verify config values
	expectedConfig := &config.Config{
		Templates: []config.LogTemplate{
			{
				Template: "Test template 1 {{level}} {{service}}",
				Weight:   10,
			},
			{
				Template: "Test template 2",
				Weight:   5,
			},
		},
		CustomTypes: map[string][]string{
			"level":   {"INFO", "ERROR"},
			"service": {"API", "DB"},
		},
		Seed: 12345,
	}

	// Check templates
	if len(cfg.Templates) != len(expectedConfig.Templates) {
		t.Errorf("Expected %d templates, got %d", len(expectedConfig.Templates), len(cfg.Templates))
	}
	for i, tmpl := range cfg.Templates {
		if tmpl.Template != expectedConfig.Templates[i].Template || tmpl.Weight != expectedConfig.Templates[i].Weight {
			t.Errorf("Template %d mismatch. Expected %+v, got %+v", i, expectedConfig.Templates[i], tmpl)
		}
	}

	// Check custom types
	if !reflect.DeepEqual(cfg.CustomTypes, expectedConfig.CustomTypes) {
		t.Errorf("Custom types mismatch. Expected %+v, got %+v", expectedConfig.CustomTypes, cfg.CustomTypes)
	}

	// Check seed
	if cfg.Seed != expectedConfig.Seed {
		t.Errorf("Seed mismatch. Expected %d, got %d", expectedConfig.Seed, cfg.Seed)
	}
}

func TestReadConfigWithMissingFile(t *testing.T) {
	_, err := config.ReadConfig("nonexistent_file.yaml")
	if err == nil {
		t.Errorf("Expected error when reading nonexistent file, got nil")
	}
}

func TestReadConfigWithInvalidYaml(t *testing.T) {
	// Create a temporary file with invalid YAML
	content := `
templates:
  - template: "Test template 1"
    weight: 10
  - template: "Test template 2"
    weight: invalid  # This should cause an error
`
	tmpfile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Try to read the invalid config
	_, err = config.ReadConfig(tmpfile.Name())
	if err == nil {
		t.Errorf("Expected error when parsing invalid YAML, got nil")
	}
}

func TestReadConfigWithEmptyCustomTypes(t *testing.T) {
	// Create config with no custom_types field
	content := `
templates:
  - template: "Test template"
    weight: 10
seed: 12345
`
	tmpfile, err := os.CreateTemp("", "empty-custom-types-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Read the config
	cfg, err := config.ReadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// CustomTypes should be initialized to an empty map
	if cfg.CustomTypes == nil {
		t.Errorf("CustomTypes is nil, expected an initialized empty map")
	}
	if len(cfg.CustomTypes) != 0 {
		t.Errorf("CustomTypes is not empty, got %+v", cfg.CustomTypes)
	}
}
