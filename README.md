# genlog

A flexible and powerful fake log generator for testing and development purposes.

[![Go Reference](https://pkg.go.dev/badge/github.com/P1llus/genlog.svg)](https://pkg.go.dev/github.com/P1llus/genlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/P1llus/genlog)](https://goreportcard.com/report/github.com/P1llus/genlog)

## Features

- Template-based log generation with customizable patterns and outputs
- Weighted template distribution for realistic log patterns
- Support for custom data types and values
- Deterministic generation with optional seeds for reproducible results
- Easy-to-use command-line interface
- Also available as a Go package for integration into existing projects

## Installation

### As a command-line tool

```bash
go install github.com/P1llus/genlog/cmd/genlog@latest
```

### As a library

```bash
go get github.com/P1llus/genlog
```

## Usage

### Command-line

```bash
# Generate 100 log lines using the default configuration (expects config.yaml in the current directory)
genlog --count=100

# Specify a custom configuration file and output location
genlog --config=myconfig.yaml --output=app.log --count=1000
```

### As a library

genlog can be used in two ways - with the simplified direct import or by using the underlying packages.

#### Simplified Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/P1llus/genlog"
)

func main() {
	// Create a generator directly from a config file (simplest approach)
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

#### Advanced Usage

```go
// Package main demonstrates usage with manually created configuration.
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

	// Generate and print a few sample log lines
	fmt.Println("Sample generated log lines:")
	for i := 0; i < 3; i++ {
		logLine, err := gen.GenerateLogLine()
		if err != nil {
			log.Fatalf("Failed to generate log line: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, logLine)
	}
}

```

## Configuration

genlog uses YAML for configuration. Here's an example:

```yaml
# Optional seed for reproducible generation, uncomment to use
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

  Protocol:
    - tcp
    - udp
    - http

  Zone:
    - inside
    - outside
    - dmz
    - public

  DomainUsername:
    - "LOCAL\\username"
    - "DOMAIN\\admin"
    - "CORP\\user"
    - "LOCAL\\sysadmin"

  VpnGroup:
    - VPN_USERS
    - ADMIN_VPN

  Username:
    - example.user
    - admin.user
    - guest.access

  HexCode:
    - "0x93d0e533"
    - "0xbc56e123"
```

## Template Functions

### Built-in Functions

genlog leverages [gofakeit](https://github.com/brianvoe/gofakeit) to provide a wide range of built-in functions for generating realistic fake data. Some commonly used functions include:

- **Personal Information**: `Name`, `FirstName`, `LastName`, `Email`, `Phone`, `Username`
- **Internet**: `URL`, `DomainName`, `IPv4Address`, `IPv6Address`, `MacAddress`, `UserAgent`
- **Log Related**: `LogLevel`, `HTTPMethod`, `HTTPStatusCode`, `HTTPVersion`
- **Text**: `Word`, `Sentence`, `Paragraph`, `Quote`, `UUID`
- **Numbers**: `Number`, `Int32`, `Int64`, `Float32`, `Float64`, `Digit`
- **Date/Time**: `Date`, `NanoSecond`, `Second`, `Minute`, `Hour`, `Month`, `WeekDay`, `Year`, `TimeZone`
- **Location**: `Latitude`, `Longitude`, `Country`, `City`, `State`, `StreetName`, `Zip`
- **Error**: `Error`, `ErrorHTTP`, `ErrorDatabase`

For a complete list of available functions, please refer to the [gofakeit functions documentation](https://github.com/brianvoe/gofakeit?tab=readme-ov-file#functions).

### Custom Functions

In addition to the gofakeit functions, genlog provides the following custom functions:

- All custom types defined in your configuration file, they are referenced by a key and a list of possible values.
- `FormattedDate`: Generates a random date in the specified format using golangs date formatting syntax.

### Example Template Usage

```
{{Name}} logged in from {{IPv4Address}} at {{FormattedDate "2006-01-02T15:04:05"}}
{{Username}} {{HTTPMethod}} {{URL}} {{HTTPStatusCode}}
User {{UUID}} accessed resource from {{City}}, {{Country}}
[{{LogLevel "short"}}] Connection from {{IPv4Address}} failed: {{ErrorDatabase}}
```

## License

MIT License - see the LICENSE file for details.
