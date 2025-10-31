// Package config provides configuration structures and parsing for the genlog log generator.
//
// This package defines all configuration options including templates, outputs,
// performance tuning, and generation limits.
package config

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// TemplateFunc is the signature for custom template functions.
// Functions receive variable arguments as strings and return a string.
// Example: func(args ...string) string { return strings.Join(args, "-") }
type TemplateFunc func(args ...string) string

// Config represents the main configuration for the log generator.
type Config struct {
	// Templates defines the log templates with their weights
	Templates []Template `yaml:"templates"`

	// CustomTypes defines custom placeholder types that can be referenced in templates
	CustomTypes map[string][]string `yaml:"custom_types,omitempty"`

	// CustomFunctions allows registering Go functions that can be called from templates.
	// Function signature: func(args ...string) string
	// Example: {{MyFunc "arg1" "arg2"}} will call CustomFunctions["MyFunc"]("arg1", "arg2")
	// Note: This field cannot be set via YAML, only programmatically
	CustomFunctions map[string]TemplateFunc `yaml:"-"`

	// Performance tuning parameters
	Performance Performance `yaml:"performance,omitempty"`

	// Outputs defines where logs should be written
	Outputs []OutputConfig `yaml:"outputs"`

	// Limits define when generation should stop (whichever is hit first)
	Limits Limits `yaml:"limits,omitempty"`

	// Seed for reproducible random generation (optional)
	Seed int64 `yaml:"seed,omitempty"`
}

// Template represents a log template with its selection weight.
type Template struct {
	// Template is the log format string with placeholders like {{username}}, {{ipv4}}, etc.
	Template string `yaml:"template"`

	// Weight determines the probability of this template being selected
	// Higher weights = higher selection probability
	Weight int `yaml:"weight"`
}

// Performance contains parameters for tuning generation performance.
type Performance struct {
	// ProducerWorkers is the number of goroutines generating logs
	// Default: number of CPU cores
	ProducerWorkers int `yaml:"producer_workers,omitempty"`

	// ConsumerWorkers is the number of goroutines writing logs
	// Default: 2
	ConsumerWorkers int `yaml:"consumer_workers,omitempty"`

	// ChannelBuffer is the size of the channel between producers and consumers
	// Larger = more memory but better throughput
	// Default: 100000
	ChannelBuffer int `yaml:"channel_buffer,omitempty"`

	// WriteBatchSize is the number of logs to batch before writing
	// Default: 1000
	WriteBatchSize int `yaml:"write_batch_size,omitempty"`

	// ValuePoolSize is the number of pre-generated values per type
	// Default: 10000
	ValuePoolSize int `yaml:"value_pool_size,omitempty"`
}

// OutputConfig defines a single output destination.
type OutputConfig struct {
	// Type is the output type (file, udp, tcp)
	Type string `yaml:"type"`

	// Config contains type-specific configuration
	Config map[string]interface{} `yaml:"config"`
}

// Limits define when log generation should stop.
// If multiple limits are specified, generation stops when ANY limit is reached.
type Limits struct {
	// MaxCount is the maximum number of logs to generate
	MaxCount int64 `yaml:"max_count,omitempty"`

	// MaxDuration is the maximum time to run generation
	// Examples: "1h", "30m", "1h30m"
	MaxDuration time.Duration `yaml:"max_duration,omitempty"`

	// MaxSize is the approximate maximum total size in bytes
	MaxSize int64 `yaml:"max_size,omitempty"`
}

// LoadFromFile reads and parses a YAML configuration file.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults
	cfg.applyDefaults()

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// applyDefaults sets default values for optional fields.
func (c *Config) applyDefaults() {
	// Initialize CustomTypes if nil
	if c.CustomTypes == nil {
		c.CustomTypes = make(map[string][]string)
	}

	// Performance defaults
	if c.Performance.ProducerWorkers == 0 {
		c.Performance.ProducerWorkers = runtime.NumCPU()
	}
	if c.Performance.ConsumerWorkers == 0 {
		c.Performance.ConsumerWorkers = 2
	}
	if c.Performance.ChannelBuffer == 0 {
		c.Performance.ChannelBuffer = 100000
	}
	if c.Performance.WriteBatchSize == 0 {
		c.Performance.WriteBatchSize = 1000
	}
	if c.Performance.ValuePoolSize == 0 {
		c.Performance.ValuePoolSize = 10000
	}
}

// Validate checks if the configuration is valid and applies defaults.
func (c *Config) Validate() error {
	// Apply defaults first
	c.applyDefaults()

	if len(c.Templates) == 0 {
		return fmt.Errorf("at least one template is required")
	}

	// Validate templates
	for i, tmpl := range c.Templates {
		if tmpl.Template == "" {
			return fmt.Errorf("template %d: template string is empty", i)
		}
		if tmpl.Weight <= 0 {
			return fmt.Errorf("template %d: weight must be positive", i)
		}
	}

	// Validate outputs
	if len(c.Outputs) == 0 {
		return fmt.Errorf("at least one output is required")
	}

	for i, output := range c.Outputs {
		if output.Type == "" {
			return fmt.Errorf("output %d: type is required", i)
		}
		if output.Config == nil {
			return fmt.Errorf("output %d: config is required", i)
		}
	}

	// Validate performance settings
	if c.Performance.ProducerWorkers < 1 {
		return fmt.Errorf("producer_workers must be at least 1")
	}
	if c.Performance.ConsumerWorkers < 1 {
		return fmt.Errorf("consumer_workers must be at least 1")
	}
	if c.Performance.ChannelBuffer < 1 {
		return fmt.Errorf("channel_buffer must be at least 1")
	}
	if c.Performance.WriteBatchSize < 1 {
		return fmt.Errorf("write_batch_size must be at least 1")
	}
	if c.Performance.ValuePoolSize < 100 {
		return fmt.Errorf("value_pool_size must be at least 100")
	}

	return nil
}

// TotalWeight returns the sum of all template weights.
func (c *Config) TotalWeight() int {
	total := 0
	for _, tmpl := range c.Templates {
		total += tmpl.Weight
	}
	return total
}
