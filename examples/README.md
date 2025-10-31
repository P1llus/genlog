# Examples

This directory contains working examples demonstrating various genlog features.

## Running Examples

Each example is in its own directory and can be run independently:

### Simple Example

Basic library usage with programmatic configuration:

```bash
cd examples/simple
go run main.go
```

Demonstrates:
- Creating a configuration programmatically
- Multiple template formats
- Custom types
- File output
- Statistics

### Custom Functions Example

Advanced usage with custom Go functions:

```bash
cd examples/customfunc
go run main.go
```

Demonstrates:
- Registering custom template functions
- String transformation functions
- Overriding built-in functions
- Multiple function types (Uppercase, Hash, Repeat, etc.)

## Example Structure

Each example directory contains:
- `main.go` - The runnable example code
- Generated output files (automatically created)

## Clean Up

Examples write to temporary directories and clean up automatically. No manual cleanup needed!

## Adding New Examples

To add a new example:

1. Create a new directory: `examples/myexample/`
2. Add `main.go` with `package main`
3. Update this README with the new example
4. Test with: `cd examples/myexample && go run main.go`

