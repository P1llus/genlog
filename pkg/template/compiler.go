package template

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CompiledTemplate represents a log template that has been parsed and compiled into
// an efficient execution plan for fast log generation.
//
// Templates are compiled once at initialization, avoiding the overhead of parsing
// template strings during hot-path log generation.
type CompiledTemplate struct {
	// Parts is the sequence of static strings and placeholders
	Parts []Part

	// Weight for weighted random selection
	Weight int

	// Original template string (for debugging)
	Original string

	// CustomFunctions contains user-registered functions available to this template
	CustomFunctions map[string]func(args ...string) string
}

// Part represents a single part of a compiled template.
type Part struct {
	// IsStatic indicates if this is a static string (true) or placeholder (false)
	IsStatic bool

	// Value is either:
	// - The static string content (if IsStatic=true)
	// - The placeholder name (if IsStatic=false)
	Value string

	// For special placeholders with arguments
	Type string   // "FormattedDate", "Number", etc.
	Args []string // Arguments for the placeholder
}

var (
	// Pattern to match placeholders: {{name}} or {{Function arg1 arg2}}
	placeholderPattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)

	// Pattern to parse function calls: FunctionName "arg1" arg2
	functionPattern = regexp.MustCompile(`^(\w+)\s+(.+)$`)
)

// CompileTemplate compiles a template string into an efficient execution plan.
func CompileTemplate(templateStr string, weight int) (*CompiledTemplate, error) {
	ct := &CompiledTemplate{
		Weight:   weight,
		Original: templateStr,
		Parts:    make([]Part, 0),
	}

	// Find all placeholders
	matches := placeholderPattern.FindAllStringSubmatchIndex(templateStr, -1)

	if len(matches) == 0 {
		// No placeholders, just a static string
		ct.Parts = append(ct.Parts, Part{
			IsStatic: true,
			Value:    templateStr,
		})
		return ct, nil
	}

	// Build parts by alternating between static strings and placeholders
	lastIdx := 0
	for _, match := range matches {
		// match[0], match[1] = full match positions
		// match[2], match[3] = captured group positions (content inside {{}})

		// Add static string before this placeholder (if any)
		if match[0] > lastIdx {
			ct.Parts = append(ct.Parts, Part{
				IsStatic: true,
				Value:    templateStr[lastIdx:match[0]],
			})
		}

		// Parse the placeholder content
		placeholder := templateStr[match[2]:match[3]]
		part, err := parsePlaceholder(placeholder)
		if err != nil {
			return nil, fmt.Errorf("failed to parse placeholder {{%s}}: %w", placeholder, err)
		}
		ct.Parts = append(ct.Parts, part)

		lastIdx = match[1]
	}

	// Add any remaining static string after the last placeholder
	if lastIdx < len(templateStr) {
		ct.Parts = append(ct.Parts, Part{
			IsStatic: true,
			Value:    templateStr[lastIdx:],
		})
	}

	return ct, nil
}

// parsePlaceholder parses a placeholder string into a Part.
func parsePlaceholder(placeholder string) (Part, error) {
	placeholder = strings.TrimSpace(placeholder)

	// Check if it's a function call (has arguments)
	if functionPattern.MatchString(placeholder) {
		matches := functionPattern.FindStringSubmatch(placeholder)
		if len(matches) != 3 {
			return Part{}, fmt.Errorf("invalid function syntax")
		}

		funcName := matches[1]
		argsStr := matches[2]

		// Parse arguments (handle quoted strings and space-separated values)
		args, err := parseArgs(argsStr)
		if err != nil {
			return Part{}, fmt.Errorf("failed to parse arguments: %w", err)
		}

		return Part{
			IsStatic: false,
			Type:     funcName,
			Args:     args,
		}, nil
	}

	// Simple placeholder without arguments
	return Part{
		IsStatic: false,
		Value:    placeholder,
	}, nil
}

// parseArgs parses function arguments, handling quoted strings.
func parseArgs(argsStr string) ([]string, error) {
	args := make([]string, 0)
	current := ""
	inQuote := false
	quoteChar := rune(0)

	for i, ch := range argsStr {
		switch {
		case (ch == '"' || ch == '\'') && !inQuote:
			inQuote = true
			quoteChar = ch
		case ch == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
			args = append(args, current)
			current = ""
		case ch == ' ' && !inQuote:
			if current != "" {
				args = append(args, current)
				current = ""
			}
		default:
			current += string(ch)
		}

		// Handle end of string
		if i == len(argsStr)-1 && current != "" {
			args = append(args, current)
		}
	}

	if inQuote {
		return nil, fmt.Errorf("unterminated quoted string")
	}

	return args, nil
}

// Execute generates a log line from this compiled template using the value pool.
func (ct *CompiledTemplate) Execute(vp *ValuePool, builder *strings.Builder) {
	builder.Reset()

	for _, part := range ct.Parts {
		if part.IsStatic {
			builder.WriteString(part.Value)
		} else {
			// Check for custom functions first (user-defined take priority)
			if part.Type != "" && ct.CustomFunctions != nil {
				if customFunc, exists := ct.CustomFunctions[part.Type]; exists {
					result := customFunc(part.Args...)
					builder.WriteString(result)
					continue
				}
			}

			// Handle built-in special function types
			switch part.Type {
			case "FormattedDate":
				// FormattedDate "2006-01-02 15:04:05"
				format := time.RFC3339 // default
				if len(part.Args) > 0 {
					format = part.Args[0]
				}
				builder.WriteString(vp.GetFormattedDate(format))

			case "Number":
				// Number min max
				if len(part.Args) >= 2 {
					min, err1 := strconv.Atoi(part.Args[0])
					max, err2 := strconv.Atoi(part.Args[1])
					if err1 == nil && err2 == nil {
						builder.WriteString(vp.GetNumberRange(min, max))
					} else {
						builder.WriteString(vp.Get("number"))
					}
				} else {
					builder.WriteString(vp.Get("number"))
				}

			default:
				// Simple placeholder lookup from value pool
				builder.WriteString(vp.Get(part.Value))
			}
		}
	}
}

// Compiler manages a collection of compiled templates.
type Compiler struct {
	templates       []*CompiledTemplate
	totalWeight     int
	customFunctions map[string]func(args ...string) string
}

// NewCompiler creates a new template compiler from a list of template strings.
func NewCompiler(templates []struct {
	Template string
	Weight   int
}, customFunctions map[string]func(args ...string) string) (*Compiler, error) {
	c := &Compiler{
		templates:       make([]*CompiledTemplate, 0, len(templates)),
		customFunctions: customFunctions,
	}

	for i, tmpl := range templates {
		compiled, err := CompileTemplate(tmpl.Template, tmpl.Weight)
		if err != nil {
			return nil, fmt.Errorf("failed to compile template %d: %w", i, err)
		}
		// Attach custom functions to each compiled template
		compiled.CustomFunctions = customFunctions
		c.templates = append(c.templates, compiled)
		c.totalWeight += tmpl.Weight
	}

	if len(c.templates) == 0 {
		return nil, fmt.Errorf("no templates provided")
	}

	return c, nil
}

// SelectTemplate selects a random template based on weights.
func (c *Compiler) SelectTemplate(rng *rand.Rand) *CompiledTemplate {
	if len(c.templates) == 1 {
		return c.templates[0]
	}

	r := rng.Intn(c.totalWeight)
	sum := 0
	for _, tmpl := range c.templates {
		sum += tmpl.Weight
		if r < sum {
			return tmpl
		}
	}

	// Fallback to first template (shouldn't happen)
	return c.templates[0]
}

// Templates returns all compiled templates.
func (c *Compiler) Templates() []*CompiledTemplate {
	return c.templates
}
