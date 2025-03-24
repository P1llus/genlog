package config

import (
	"os"
	"testing"
)

func TestReadConfig(t *testing.T) {
	// Create a temporary config file
	content := `
templates:
  - template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] {{message}}"
    weight: 1
custom_types:
  message:
    - "Test message 1"
    - "Test message 2"
outputs:
  - type: file
    workers: 1
    config:
      filename: "test.log"
seed: 12345
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test reading the config
	cfg, err := ReadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("ReadConfig failed: %v", err)
	}

	// Verify the config contents
	if len(cfg.Templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(cfg.Templates))
	}
	if cfg.Templates[0].Weight != 1 {
		t.Errorf("Expected weight 1, got %d", cfg.Templates[0].Weight)
	}
	if len(cfg.CustomTypes["message"]) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(cfg.CustomTypes["message"]))
	}
	if len(cfg.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(cfg.Outputs))
	}
	if cfg.Outputs[0].Type != OutputTypeFile {
		t.Errorf("Expected output type file, got %s", cfg.Outputs[0].Type)
	}
	if cfg.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", cfg.Seed)
	}
}

func TestReadConfigInvalidFile(t *testing.T) {
	_, err := ReadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestReadConfigInvalidYAML(t *testing.T) {
	// Create a temporary config file with invalid YAML
	content := `
templates:
  - template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] {{message}}"
    weight: invalid
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = ReadConfig(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Templates: []LogTemplate{
					{
						Template: "test template",
						Weight:   1,
					},
				},
				Outputs: []OutputConfig{
					{
						Type:    OutputTypeFile,
						Workers: 1,
						Config: map[string]interface{}{
							"filename": "test.log",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no templates",
			config: &Config{
				Outputs: []OutputConfig{
					{
						Type:    OutputTypeFile,
						Workers: 1,
						Config: map[string]interface{}{
							"filename": "test.log",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no outputs",
			config: &Config{
				Templates: []LogTemplate{
					{
						Template: "test template",
						Weight:   1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid output type",
			config: &Config{
				Templates: []LogTemplate{
					{
						Template: "test template",
						Weight:   1,
					},
				},
				Outputs: []OutputConfig{
					{
						Type:    "invalid",
						Workers: 1,
						Config: map[string]interface{}{
							"filename": "test.log",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test assumes we add a Validate method to the Config struct
			// You might want to add this method to the Config struct
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
