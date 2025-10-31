# genlog

[![Go Reference](https://pkg.go.dev/badge/github.com/P1llus/genlog.svg)](https://pkg.go.dev/github.com/P1llus/genlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/P1llus/genlog)](https://goreportcard.com/report/github.com/P1llus/genlog)

A high-performance fake log generator for testing and benchmarking purposes.

## Installation

### As a Command-Line Tool

```bash
go install github.com/P1llus/genlog/cmd/genlog@latest
```

### As a Library

```bash
go get github.com/P1llus/genlog
```

## Quick Start

### Command-Line Usage

```bash
# Generate a sample configuration file (creates config.yaml)
genlog -init

# Or specify a custom filename
genlog -init my-config.yaml

# Edit the config to customize templates and outputs
vim config.yaml  # or your custom filename

# Generate logs
genlog -config config.yaml
```

### Library Usage

```go
package main

import (
    "log"
    "github.com/P1llus/genlog"
)

func main() {
    cfg := &genlog.Config{
        Templates: []genlog.Template{
            {
                Template: "{{timestamp}} [{{level}}] {{message}}",
                Weight:   10,
            },
        },
        CustomTypes: map[string][]string{
            "level":   {"INFO", "WARN", "ERROR"},
            "message": {"User logged in", "Request processed"},
        },
        Outputs: []genlog.OutputConfig{
            {
                Type: "file",
                Config: map[string]interface{}{
                    "path": "output.log",
                },
            },
        },
        Limits: genlog.Limits{
            MaxCount: 1000000, // Generate 1M logs
        },
    }

    gen, err := genlog.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    gen.Start()
    gen.Wait() // Wait for completion
    gen.Stop()
}
```

## Custom Functions

Library users can register custom Go functions to use in templates:

```go
cfg := &genlog.Config{
    Templates: []genlog.Template{
        {
            Template: "{{Uppercase \"alert\"}} - {{ipv4}} - {{level}}",
            Weight:   1,
        },
    },
    CustomFunctions: map[string]genlog.TemplateFunc{
        // Convert string to uppercase
        "Uppercase": func(args ...string) string {
            if len(args) == 0 {
                return ""
            }
            return strings.ToUpper(args[0])
        },
        // Hash a string (simple example)
        "Hash": func(args ...string) string {
            if len(args) == 0 {
                return ""
            }
            return fmt.Sprintf("%x", []byte(args[0]))[:8]
        },
        // Repeat a string N times
        "Repeat": func(args ...string) string {
            if len(args) < 2 {
                return ""
            }
            count, _ := strconv.Atoi(args[1])
            return strings.Repeat(args[0], count)
        },
    },
    // ... rest of config
}
```

Custom functions:
- Accept variable string arguments: `func(args ...string) string`
- Can override built-in functions (FormattedDate, Number)
- Are only available when using genlog as a library (not from YAML)

See `examples/customfunc/` for a complete working example.

## Configuration

### YAML Configuration Example

```yaml
# Optional: Set seed for reproducible generation
seed: 12345

# Log templates with weights (higher weight = higher selection probability)
templates:
  - template: '{{FormattedDate "2006-01-02T15:04:05.000Z07:00"}} [{{level}}] {{message}}'
    weight: 10
  - template: '{"timestamp":"{{timestamp}}","level":"{{level}}","msg":"{{message}}"}'
    weight: 5

# Custom types that can be referenced in templates
custom_types:
  level:
    - INFO
    - WARN
    - ERROR
  message:
    - "User authenticated"
    - "Request processed"
    - "Database query executed"

# Performance tuning (optional)
performance:
  producer_workers: 8      # Goroutines generating logs (default: NumCPU)
  consumer_workers: 2      # Goroutines writing logs (default: 2)
  channel_buffer: 100000   # Buffer size between producers/consumers
  write_batch_size: 1000   # Logs per batch write
  value_pool_size: 10000   # Pre-generated values per type

# Output destinations
outputs:
  - type: file
    config:
      path: "output.log"
  - type: udp
    config:
      address: "localhost:514"

# Limits (generation stops when ANY limit is reached)
limits:
  max_count: 1000000      # Total logs to generate
  max_duration: "5m"      # Maximum runtime
  max_size: 1073741824    # Maximum bytes (1GB)
```

## Template Syntax

### Built-in Placeholders

Templates use `{{placeholder}}` syntax. Common placeholders:

**Network:**
- `{{ipv4}}`, `{{IPv4Address}}` - Random IPv4 address
- `{{ipv6}}`, `{{IPv6Address}}` - Random IPv6 address
- `{{domain}}`, `{{DomainName}}` - Random domain name
- `{{url}}`, `{{URL}}` - Random URL

**User Info:**
- `{{username}}`, `{{Username}}` - Random username
- `{{email}}`, `{{Email}}` - Random email address

**HTTP:**
- `{{useragent}}`, `{{UserAgent}}` - Random user agent string
- `{{httpmethod}}`, `{{HTTPMethod}}` - HTTP method (GET, POST, etc.)
- `{{httpstatus}}`, `{{HTTPStatusCode}}` - HTTP status code

**Logging:**
- `{{loglevel}}`, `{{LogLevel}}`, `{{level}}` - Log level
- `{{timestamp}}`, `{{Timestamp}}` - ISO 8601 timestamp

**Numbers:**
- `{{number}}`, `{{Number}}` - Random number
- `{{port}}`, `{{Port}}` - Random port number (1-65535)

### Special Functions

**FormattedDate**
```
{{FormattedDate "2006-01-02 15:04:05"}}
{{FormattedDate "Jan 2, 2006"}}
```

**Number Range**
```
{{Number 1 100}}      # Random number between 1 and 100
{{Number 1000 9999}}  # Random 4-digit number
```

### Custom Types

Define your own placeholder types in the configuration:

```yaml
custom_types:
  service:
    - API
    - Database
    - Cache
  severity:
    - critical
    - high
    - medium
    - low
```

Use them in templates:
```
{{timestamp}} {{service}}: {{severity}} alert detected
```

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────┐
│                   Generator                          │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │Producer 1│  │Producer 2│  │Producer N│         │
│  │(generate)│  │(generate)│  │(generate)│         │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘         │
│       └─────────────┼─────────────┘                │
│                     ▼                               │
│          ┌────────────────────┐                    │
│          │  Buffered Channel  │                    │
│          │  (100k capacity)   │                    │
│          └────────────────────┘                    │
│                     │                               │
│       ┌─────────────┼─────────────┐                │
│       ▼             ▼             ▼                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │Consumer 1│  │Consumer 2│  │Consumer M│         │
│  │ (write)  │  │ (write)  │  │ (write)  │         │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘         │
│       ▼             ▼             ▼                │
│     File          UDP           TCP                │
└─────────────────────────────────────────────────────┘
```

## Output Types

### File Output

```yaml
outputs:
  - type: file
    config:
      path: "output.log"
```

### UDP Output

```yaml
outputs:
  - type: udp
    config:
      address: "localhost:514"
```

## Examples

The `examples/` directory contains working examples demonstrating various features:

### Simple Example
```bash
cd examples/simple
go run main.go
```
Demonstrates basic programmatic configuration and library usage.

### Custom Functions Example
```bash
cd examples/customfunc
go run main.go
```
Demonstrates registering custom Go functions for use in templates.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see LICENSE file for details.

