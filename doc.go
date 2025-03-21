/*
Package genlog provides a simple API for generating fake log data based on
customizable templates and patterns. It's designed for testing, development,
and demonstration purposes when realistic log data is needed.

# Basic Usage

The simplest way to use genlog is through a configuration file:

	import "github.com/P1llus/genlog"

	// Create a log generator from a config file
	gen, err := genlog.NewFromFile("config.yaml")
	if err != nil {
		// handle error
	}

	// Generate 1000 log lines to output.log
	err = gen.GenerateLogs("output.log", 1000)
	if err != nil {
		// handle error
	}

# Programmatic Configuration

You can also create configurations programmatically:

	cfg := &genlog.Config{
		Templates: []genlog.LogTemplate{
			{
				Template: "{{FormattedDate \"2006-01-02T15:04:05.000Z07:00\"}} [INFO] User {{username}} logged in from {{IPV4Address}}",
				Weight:   1,
			},
		},
		CustomTypes: map[string][]string{
			"username": {"john", "alice", "bob"},
		},
	}

	gen := genlog.NewFromConfig(cfg)
	logLine, err := gen.GenerateLogLine()

# Template Syntax

Templates use the golang template syntax with surrounded by double braces.
You can use functions available from gofakeit, custom types defined in configurations, and built-in custom functions:
More examples can be found on the GitHub repository.

- Basic placeholders: {{FirstName}}, {{URL}}, {{IPV4Address}}
- Custom types: {{username}} (defined in your config)
- Built-in custom functions: {{FormattedDate "2006-01-02 15:04:05"}}

# Command Line Tool

Genlog can also be used as a command line tool, and can be installed either with go-get or by downloading a pre-built binary from the releases page:

	genlog -config=myconfig.yaml -output=logs.txt -count=1000

See the GitHub repository for more information and examples:
https://github.com/P1llus/genlog
*/
package genlog
