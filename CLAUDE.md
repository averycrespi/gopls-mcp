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

**Main Flow**: MCP Client → MCP Server → LSP Manager → gopls LSP Client → gopls binary

- `cmd/gopls-mcp/main.go` - Entry point, handles CLI flags and server lifecycle
- `internal/server/server.go` - MCP server implementation with 6 registered tools
- `internal/lsp/manager.go` - Manages LSP client lifecycle and thread safety
- `internal/lsp/client.go` - LSP client that communicates with gopls via JSON-RPC
- `pkg/types/types.go` - Shared types for LSP operations

### Key Design Patterns

**Tool Registration**: Each MCP tool maps to a specific LSP operation:
- `go_to_definition` → LSP Definition request
- `find_references` → LSP References request  
- `hover_info` → LSP Hover request
- `get_completion` → LSP Completion request
- `format_code` → LSP DocumentFormatting request
- `rename_symbol` → LSP Rename request

**Path Handling**: All file paths are converted to absolute paths and file:// URIs for LSP communication.

**Error Handling**: LSP errors are wrapped and returned as MCP tool result errors.

## MCP Integration

The server uses the `github.com/mark3labs/mcp-go` framework and communicates via stdin/stdout. It can be integrated with Claude Code using:

```bash
claude mcp add gopls-mcp go run github.com/averycrespi/gopls-mcp@latest
```

All tools require `file_path`, `line`, and `character` parameters (0-based indexing) except `format_code` which only needs `file_path`.