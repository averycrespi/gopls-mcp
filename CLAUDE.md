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

**Main Flow**: MCP Client → MCP Server → Client Manager → Gopls Client → Transport → gopls binary

- `cmd/gopls-mcp/main.go` - Entry point, handles CLI flags and server lifecycle
- `internal/server/server.go` - MCP server implementation (GoplsServer) and tool registration
- `internal/client/manager.go` - Manages LSP client lifecycle and thread safety
- `internal/client/client.go` - Gopls client that communicates with gopls via JSON-RPC
- `internal/transport/transport.go` - JSON-RPC transport layer for LSP communication
- `internal/tools/` - Individual tool implementations (one file per MCP tool)
- `pkg/types/` - Shared type definitions split into domain files:
  - `client.go` - LSP client interface and related types
  - `server.go` - Server interface and configuration types
  - `transport.go` - Transport interface for JSON-RPC communication

### Key Design Patterns

**Tool Registration**: Each MCP tool is implemented in its own file in `internal/tools/`:
- `go_to_definition.go` - `gopls.go_to_definition` → LSP Definition request
- `find_references.go` - `gopls.find_references` → LSP References request
- `hover_info.go` - `gopls.hover_info` → LSP Hover request
- `get_completion.go` - `gopls.get_completion` → LSP Completion request
- `format_code.go` - `gopls.format_code` → LSP DocumentFormatting request
- `rename_symbol.go` - `gopls.rename_symbol` → LSP Rename request
- `tools.go` - Shared utilities for path handling and position parsing

**Architecture Pattern**: Tools are registered in `server.go` with a wrapper that injects the Client interface dynamically, allowing tools to be stateless. The GoplsServer implements the Server interface and coordinates between MCP tools and the LSP client.

**Interface Design**: The codebase uses clean interfaces to separate concerns:
- `types.Client` - Defines LSP client operations (implemented by GoplsClient)
- `types.Server` - Defines MCP server operations (implemented by GoplsServer)  
- `types.Transport` - Defines JSON-RPC transport operations (implemented by JsonRpcTransport)

**Path Handling**: All file paths are converted to absolute paths and file:// URIs for LSP communication.

**Error Handling**: LSP errors are wrapped and returned as MCP tool result errors with improved logging and specific error messages.

## MCP Integration

The server uses the `github.com/mark3labs/mcp-go` framework and communicates via stdin/stdout. It can be integrated with Claude Code using:

```bash
claude mcp add gopls-mcp go run github.com/averycrespi/gopls-mcp@latest
```

All tools require `file_path`, `line`, and `character` parameters (0-based indexing) except `gopls.format_code` which only needs `file_path`.

## Development Guidelines

**IMPORTANT**: After making any code changes, always run these commands to ensure code quality:

1. `make test` - Run unit tests to verify functionality
2. `go fmt ./...` - Format Go code (if not using gofmt integration)
3. `golangci-lint run` - Lint Go code for potential issues

These commands must be run before committing changes to ensure the codebase remains stable and follows Go best practices.
