# Integration Test Fixtures

This directory contains test fixtures for the gopls-mcp integration tests.

## Purpose

This standalone Go module provides known, stable symbols for testing LSP functionality through the MCP server. Unlike testing against the main codebase where symbol positions can change, these files have predictable content and positions.

## Files

- **`go.mod`** - Standalone Go module definition
- **`main.go`** - Main function demonstrating usage of all types and functions
- **`calculator.go`** - Calculator struct with methods (Add, Subtract, Multiply, Divide, etc.)
- **`types.go`** - Custom types, interfaces, and constants (Operation enum, Processor interface, BasicProcessor)
- **`utils.go`** - Utility functions and struct (MathUtils, global functions like Factorial, Fibonacci)

## Integration Test Coverage

The integration tests use these files to test:

- **Go to Definition**: Jump from function calls to their definitions
- **Hover Info**: Get documentation and type information for symbols
- **Find References**: Find all usages of types and functions across files
- **Code Completion**: Get completion suggestions after typing `calc.`, etc.
- **Code Formatting**: Format Go code using gopls

## Symbol Positions

The files are designed with known symbol positions for reliable testing:

- Line numbers and character positions are carefully documented in tests
- Symbols are placed at predictable locations
- Multiple references exist across files for comprehensive testing

## Maintenance

When modifying these files:
1. Update corresponding integration test positions if line numbers change
2. Ensure the module remains compilable with `go build ./...`
3. Test that integration tests still pass with `make test-integration`

## Example Test Scenarios

- `NewCalculator` function call in main.go → definition in calculator.go
- `calc.Add` method call → hover shows method signature and documentation
- `Calculator` type → finds 9+ references across the codebase
- `calc.` → completion shows Add, Subtract, Multiply, Divide methods
