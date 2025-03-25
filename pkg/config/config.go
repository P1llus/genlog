// Package config provides configuration handling for the genlog tool.
// It contains structures and functions for loading, parsing, and representing
// the configuration needed for generating log entries.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// OutputType represents the type of output destination for logs
type OutputType string

const (
	OutputTypeFile OutputType = "file"
	OutputTypeUDP  OutputType = "udp"
)

// OutputConfig represents a single output configuration
type OutputConfig struct {
	// Type specifies the kind of output (file, udp, etc.)
	Type OutputType `yaml:"type"`

	// Workers specifies the number of concurrent workers for this output
	Workers int `yaml:"workers"`

	// BatchSize specifies the number of logs to write in a single batch
	BatchSize int `yaml:"batch_size"`

	// Config contains the type-specific configuration
	Config map[string]any `yaml:"config"`
}

// FileOutputConfig represents configuration specific to file outputs
type FileOutputConfig struct {
	// Filename is the path where logs will be written
	Filename string `yaml:"filename"`
}

// UDPOutputConfig represents configuration specific to UDP outputs
type UDPOutputConfig struct {
	// Address is the UDP destination address (e.g., "localhost:514")
	Address string `yaml:"address"`
}

// Config represents the main configuration structure for the log generator.
// It defines the templates to use, any custom type definitions, and optional
// seed value for deterministic generation.
//
// Example YAML configuration:
//
//		templates:
//		  - template: '{{FormattedDate "2006-01-02T15:04:05.000Z07:00"}} [{{level}}] {{username}} - {{message}}'
//		    weight: 10
//		  - template: '{{FormattedDate "Jan 2 15:04:05"}} {{level}} [{{service}}] {{IPv4Address}} {{username}}: {{message}}'
//		    weight: 3
//		custom_types:
//	   level:
//			  - INFO
//			  - WARNING
//			  - ERROR
//			  - DEBUG
//			  - TRACE
//	 service:
//	   - API
//	   - AUTH
//	   - DATABASE
//	   - CACHE
//	   - FRONTEND
//	 username:
//	   - admin
//	   - system
//	   - app
//	   - service_account
//	   - anonymous
//	 message:
//	     - "User authenticated successfully"
//	     - "Failed login attempt - invalid credentials"
//	     - "Permission denied to resource"
//	     - "Resource accessed successfully"
//	     - "API rate limit exceeded"
//	     - "Database connection timeout"
//	     - "Cache invalidation completed"
//	     - "Request processed in 235ms"
//		seed: 12345  # Optional, for reproducible results
type Config struct {
	// Templates is a slice of log templates with their respective weights.
	// At least one template is required for log generation.
	Templates []LogTemplate `yaml:"templates"`

	// Outputs defines the destinations where logs will be sent
	Outputs []OutputConfig `yaml:"outputs"`

	// CustomTypes is a map of custom type names to their possible values.
	// These can be referenced in templates and will be selected randomly.
	// For example, a custom type "username" can be used in templates as {username}.
	CustomTypes map[string][]string `yaml:"custom_types,omitempty"`

	// Seed is an optional seed value for deterministic random generation.
	// Using the same seed will produce the same sequence of logs.
	// If omitted or set to 0, a random seed will be used.
	Seed uint64 `yaml:"seed,omitempty"`
}

// LogTemplate represents a single log template with its selection weight.
// Templates use the gofakeit syntax for placeholders, such as {name}, {ipv4}, etc.
type LogTemplate struct {
	// Template is the log template string with placeholders.
	// Placeholders are enclosed in curly braces, e.g., {name}, {email}.
	// Placeholders can be:
	// - Built-in gofakeit functions: {name}, {email}, {ipv4}, etc.
	// - Custom types defined in the configuration: {username}, {severity}, etc.
	// - Special functions: {FormattedDate("2006-01-02 15:04:05")}
	Template string `yaml:"template"`

	// Weight determines the probability of this template being selected.
	// Higher weights increase the chance of selection.
	// For example, if template A has weight 10 and template B has weight 5,
	// template A will be selected roughly twice as often as template B.
	Weight int `yaml:"weight"`
}

// ReadConfig reads and parses the configuration file at the given path.
// It returns the parsed Config structure or an error if reading or parsing fails.
//
// Example:
//
//	cfg, err := config.ReadConfig("config.yaml")
//	if err != nil {
//		// handle error
//	}
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

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Templates) == 0 {
		return fmt.Errorf("no templates configured")
	}
	if len(c.Outputs) == 0 {
		return fmt.Errorf("no outputs configured")
	}
	for _, output := range c.Outputs {
		if output.BatchSize == 0 {
			output.BatchSize = 100
		}
		if output.Workers == 0 {
			output.Workers = 1
		}
		switch output.Type {
		case OutputTypeFile:
			if _, ok := output.Config["filename"].(string); !ok {
				return fmt.Errorf("filename is required for file output")
			}
		case OutputTypeUDP:
			if _, ok := output.Config["address"].(string); !ok {
				return fmt.Errorf("address is required for UDP output")
			}
		default:
			return fmt.Errorf("unsupported output type: %s", output.Type)
		}
	}
	return nil
}
