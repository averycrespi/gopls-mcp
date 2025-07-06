# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Testing
- `make build` - Build the gopls-mcp binary to `bin/gopls-mcp`
- `make test` - Run unit tests
- `make test-integration` - Run integration tests (builds binary, checks dependencies)
- `make dev` - Build and run server with current directory as workspace
- `make clean` - Clean build artifacts and caches

### Dependencies
- `make deps` - Download and tidy Go modules
- `make install-gopls` - Install gopls language server if not present
- Requires Go 1.23+ and gopls to be available

### Running the Server
- `make run WORKSPACE=/path/to/go/project` - Run with specific workspace
- `go run ./cmd/gopls-mcp -workspace-root /path/to/project` - Direct run
- `go run github.com/averycrespi/gopls-mcp@latest` - Run latest version

## Architecture

This is an MCP (Model Context Protocol) server that bridges LLMs with the Go language server (gopls).

### Core Components

**Main Flow**: MCP Client → GoplsServer → GoplsClient → Transport → gopls binary

- `cmd/gopls-mcp/main.go` - Entry point, handles CLI flags and server lifecycle
- `internal/server/server.go` - MCP server implementation (GoplsServer) with direct client usage
- `internal/client/client.go` - Gopls client that communicates with gopls via JSON-RPC
- `internal/transport/transport.go` - JSON-RPC transport layer for LSP communication
- `internal/tools/` - Individual tool implementations (one file per MCP tool)
- `internal/results/` - JSON response types and formatting utilities
- `pkg/types/` - Shared type definitions split into domain files:
  - `client.go` - LSP client interface and related types (includes Start/Stop methods)
  - `server.go` - Server interface for MCP operations (Serve method)
  - `config.go` - Configuration structure (used as value type)
  - `transport.go` - Transport interface for JSON-RPC communication (Start/Stop methods)

### Key Design Patterns

**Tool Registration**: Each MCP tool is implemented in its own file in `internal/tools/`:
- `symbol_definition.go` - `symbol_definition` → LSP WorkspaceSymbol + Definition requests
- `find_references.go` - `find_references` → LSP References request
- `symbol_search.go` - `symbol_search` → LSP WorkspaceSymbol request
- `utils.go` - Shared utilities for path handling and position parsing

**JSON Response Structure**: Structured output types in `internal/results/`:
- `symbol_kind.go` - SymbolKind enum with LSP mapping (file, function, struct, etc.)
- `symbol_location.go` - Location information with file paths and positions
- `source_context.go` - Source code context with line highlighting
- `symbol_search.go` - Search result types with metadata and documentation
- `symbol_definition.go` - Definition result types with multiple definition support

**Interface Design**: The codebase uses clean interfaces to separate concerns:
- `types.Client` - Defines LSP client operations including Start/Stop (implemented by GoplsClient)
- `types.Server` - Defines MCP server operations with Serve method (implemented by GoplsServer)
- `types.Transport` - Defines JSON-RPC transport operations with Start/Stop methods (implemented by JsonRpcTransport)
- `types.Config` - Configuration structure used as value type for better immutability

**Value vs Pointer Types**: The design uses value types for Config to ensure immutability and prevent accidental modifications, while interfaces provide the abstraction layer for different implementations.

**Path Handling**: All file paths are converted to absolute paths and file:// URIs for LSP communication.

**Error Handling**: LSP errors are wrapped and returned as MCP tool result errors with improved logging and specific error messages.

**JSON Output**: All symbol tools return structured JSON responses with:
- Type-safe SymbolKind enums (function, struct, method, etc.)
- Rich metadata including documentation from hover info
- Source code context with line numbers and highlighting
- Relative file paths from workspace root

## MCP Integration

The server uses the `github.com/mark3labs/mcp-go` framework and communicates via stdin/stdout. It can be integrated with Claude Code using:

```bash
claude mcp add gopls-mcp go run github.com/averycrespi/gopls-mcp@latest
```

## Development Guidelines

**IMPORTANT**: After making any code changes, always run these commands to ensure code quality:

1. `make test` - Run unit tests to verify functionality
2. `go fmt ./...` - Format Go code (if not using gofmt integration)
3. `golangci-lint run` - Lint Go code for potential issues

These commands must be run before committing changes to ensure the codebase remains stable and follows Go best practices.
