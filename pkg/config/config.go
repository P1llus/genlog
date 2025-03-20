// Package config provides configuration handling for the genlog tool.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure for the log generator.
type Config struct {
	// Templates is a slice of log templates with their respective weights.
	Templates []LogTemplate `yaml:"templates"`

	// CustomTypes is a map of custom type names to their possible values.
	// These can be referenced in templates and will be selected randomly.
	CustomTypes map[string][]string `yaml:"custom_types,omitempty"`

	// Seed is an optional seed value for deterministic random generation.
	Seed uint64 `yaml:"seed,omitempty"`
}

// LogTemplate represents a single log template with its selection weight.
type LogTemplate struct {
	// Template is the log template string with placeholders.
	Template string `yaml:"template"`

	// Weight determines the probability of this template being selected.
	// Higher weights increase the chance of selection.
	Weight int `yaml:"weight"`
}

// ReadConfig reads and parses the configuration file at the given path.
// It returns the parsed Config structure or an error if reading or parsing fails.
func ReadConfig(configFile string) (*Config, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	// Initialize the custom types map if it's nil
	if config.CustomTypes == nil {
		config.CustomTypes = make(map[string][]string)
	}

	return &config, nil
}
