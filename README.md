# genlog

A flexible and powerful fake log generator for testing and development purposes.

[![Go Reference](https://pkg.go.dev/badge/github.com/P1llus/genlog.svg)](https://pkg.go.dev/github.com/P1llus/genlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/P1llus/genlog)](https://goreportcard.com/report/github.com/P1llus/genlog)

Generate realistic log data with customizable templates for testing log processing, visualization tools, and SIEM systems.

## Features

- 📝 Template-based log generation with customizable patterns and outputs
- 📊 Weighted template distribution for realistic log patterns
- 🧩 Support for custom data types and values
- 🔄 Deterministic generation with optional seeds for reproducible results
- 💻 Easy-to-use command-line interface
- 📦 Available as a Go package for integration into existing projects

## Installation

### As a command-line tool

Directly with go install:

```bash
go install github.com/P1llus/genlog/cmd/genlog@latest
```

Or download the latest release for your platform from the [releases page](https://github.com/P1llus/genlog/releases).

```bash
curl -L "https://github.com/P1llus/genlog/releases/latest/download/genlog_linux_amd64" -o genlog
chmod +x genlog
```

### As a library

```bash
go get github.com/P1llus/genlog
```

## Quick Start

### Command-line

```bash
# Generate 100 log lines using the default configuration (expects config.yaml in the current directory)
# Example config can be found further down in the README
genlog --count=100

# Specify a custom configuration file and output location
genlog --config=myconfig.yaml --output=app.log --count=1000
```

### As a library

`genlog` can be used in two ways - with a YAML configuration file or with a programmatically created configuration.

#### Basic Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/P1llus/genlog"
)

func main() {
	// Create a generator from a config file
	gen, err := genlog.NewFromFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Generate logs to a file
	err = gen.GenerateLogs("output.log", 10)
	if err != nil {
		log.Fatalf("Failed to generate logs: %v", err)
	}

	// Generate and print individual log lines
	fmt.Println("Generated individual log line:")
	logLine, err := gen.GenerateLogLine()
	if err != nil {
		log.Fatalf("Failed to generate log line: %v", err)
	}
	fmt.Println(logLine)
}
```

#### Programmatic Configuration

```go
package main

import (
	"fmt"
	"log"

	"github.com/P1llus/genlog"
)

func main() {
	// Create a configuration programmatically
	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [{{level}}] {{message}}",
				Weight:   10,
			},
			{
				Template: "{\"time\":\"{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}}\",\"level\":\"{{level}}\",\"msg\":\"{{message}}\"}",
				Weight:   5,
			},
		},
		CustomTypes: map[string][]string{
			"level": {
				"INFO", "WARN", "ERROR", "DEBUG",
			},
			"message": {
				"System starting up",
				"Connection established",
				"Transaction completed",
				"User authentication failed",
				"Resource not found",
			},
		},
		Seed: 12345, // Optional: set for reproducible results
	}

	// Create a generator from the config
	gen := genlog.NewFromConfig(cfg)

	// Generate logs to a file
	err := gen.GenerateLogs("advanced-output.log", 20)
	if err != nil {
		log.Fatalf("Failed to generate logs: %v", err)
	}

	// Generate a sample log line
	logLine, err := gen.GenerateLogLine()
	if err != nil {
		log.Fatalf("Failed to generate log line: %v", err)
	}
	fmt.Printf("Sample log: %s\n", logLine)
}
```

## Configuration File

`genlog` uses YAML for configuration. Here's an example:

```yaml
# Optional seed for reproducible generation
# seed: 12345

# Log templates with weights
templates:
  - template: '{{FormattedDate "Jan 2 2006 15:04:05"}} {{ServerName}} CiscoASA[{{Number 100 999}}]: %ASA-6-305011: Built dynamic TCP translation from inside:{{IPv4Address}}/{{Number 1000 9999}} to outside:{{IPv4Address}}/{{Number 1000 9999}}'
    weight: 10
  - template: '{{FormattedDate "Jan 2 2006 15:04:05"}} {{ServerName}} CiscoASA[{{Number 100 999}}]: %ASA-6-302016: Teardown UDP connection {{Number 10000 99999}} for outside:{{IPv4Address}}/{{Number 1 65535}} to inside:{{IPv4Address}}/{{Number 1 65535}} duration {{Hour}}:{{Minute}}:{{Second}} bytes {{Number 100 9999}}'
    weight: 8

# Custom types that can be referenced in templates
custom_types:
  ServerName:
    - localhost
    - firewall-01
    - fw-edge-01
  AccessListName:
    - incoming
    - guest_access
```

## Template Syntax

Templates use placeholders in double curly braces `{{ }}` to insert randomly generated values. The available placeholders include:

### Built-in Fake Data Functions

`genlog` leverages [gofakeit](https://github.com/brianvoe/gofakeit) to provide a wide range of functions:

#### Common Categories:

- **Personal**: `{{Name}}`, `{{FirstName}}`, `{{LastName}}`, `{{Email}}`, `{{Phone}}`
- **Internet**: `{{URL}}`, `{{DomainName}}`, `{{IPv4Address}}`, `{{UserAgent}}`
- **Log Specific**: `{{LogLevel}}`, `{{HTTPMethod}}`, `{{HTTPStatusCode}}`
- **Numbers**: `{{Number 1 100}}`, `{{Int32}}`, `{{Float32}}`
- **Date/Time**: `{{Second}}`, `{{Minute}}`, `{{Hour}}`, `{{Month}}`

#### Example Template:

```yaml
templates:
	template: '{{FormattedDate "2006-01-02"}} {{LogLevel}} - User {{Username}} connected from {{IPv4Address}} using {{UserAgent}}'
	weight: 10
```

### Custom Types

Custom types defined in your config file can be used directly in templates:

```yaml
templates:
	template: '{{FormattedDate "Jan 2 2006 15:04:05"}} {{ServerName}} {{AccessListName}} request from {{IPv4Address}}'
	weight: 10
```

### Custom Built-in Functions:

- `{{FormattedDate "format"}}`: Generates a random date in the specified format using Go's date formatting syntax.

## Advanced Examples

Check the `examples/` directory for more advanced usage patterns:

- `examples/basic/`: Simple configuration with basic templates
- `examples/config/`: Programmatic configuration example

## Documentation

For complete documentation, visit the [Go package documentation](https://pkg.go.dev/github.com/P1llus/genlog).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see the LICENSE file for details.
