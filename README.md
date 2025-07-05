# gopls-mcp

An MCP (Model Context Protocol) server that exposes the LSP functionality of the [gopls](https://pkg.go.dev/golang.org/x/tools/gopls) language server, enabling LLMs to work with Go projects more effectively.

## Overview

gopls-mcp bridges the gap between LLMs and Go development by providing an MCP interface to gopls (the official Go language server). This allows AI assistants to perform sophisticated Go code analysis, navigation, and refactoring operations.

## Features

### Navigation Tools
- **go_to_definition**: Find the definition of a symbol in Go code
- **find_references**: Find all references to a symbol across the project

### Code Analysis Tools
- **hover_info**: Get detailed information about symbols (types, documentation, etc.)
- **get_completion**: Get code completion suggestions at any position

### Code Transformation Tools
- **format_code**: Format Go code using gofmt standards
- **rename_symbol**: Safely rename symbols across the entire project

## Installation

### Prerequisites
- Go 1.23 or later
- gopls (Go language server) - typically installed with `go install golang.org/x/tools/gopls@latest`

### Direct Run (Recommended)
```bash
go run github.com/averycrespi/gopls-mcp@latest
```

### Build from Source
```bash
git clone https://github.com/averycrespi/gopls-mcp && cd gopls-mcp
make build
```

## Usage

### Command Line Options
```bash
./bin/gopls-mcp [options]

Options:
  -gopls-path string
        Path to the gopls binary (default "gopls")
  -workspace-root string
        Root directory of the Go workspace (default ".")
  -log-level string
        Log level (debug, info, warn, error) (default "info")
```

### MCP Client Integration

The server communicates via stdin/stdout using the MCP protocol. It can be integrated with any MCP-compatible client.

#### Claude Code Integration

Add the server to Claude Code:
```bash
claude mcp add gopls-mcp go run github.com/averycrespi/gopls-mcp@latest
```

#### Manual Configuration

Example configuration:
```json
{
  "mcpServers": {
    "gopls-mcp": {
      "command": "go run github.com/averycrespi/gopls-mcp@latest",
      "args": []
    }
  }
}
```

## Available MCP Tools

### gopls.go_to_definition
Find where a symbol is defined.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

### gopls.find_references
Find all references to a symbol.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

### gopls.hover_info
Get hover information for a symbol.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

### gopls.get_completion
Get code completion suggestions.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

### gopls.format_code
Format Go code using gofmt.

**Parameters:**
- `file_path` (string): Path to the Go file

### gopls.rename_symbol
Rename a symbol across the project.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)
- `new_name` (string): New name for the symbol

## Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────┐
│   MCP Client    │◄──►│   gopls-mcp  │◄──►│    gopls    │
│   (Claude, etc) │    │   (Server)   │    │ (Language   │
└─────────────────┘    └──────────────┘    │  Server)    │
                                           └─────────────┘
```

The server acts as a bridge:
1. Receives MCP tool calls from clients
2. Translates them to LSP requests
3. Communicates with gopls via JSON-RPC
4. Returns formatted responses to the MCP client

## Development

### Project Structure
```
gopls-mcp/
├── cmd/gopls-mcp/         # Main application entry point
├── internal/
│   ├── server/            # MCP server implementation (GoplsServer)
│   ├── client/            # LSP client wrapper and manager
│   ├── transport/         # JSON-RPC transport layer
│   └── tools/             # Individual MCP tool implementations
├── pkg/
│   ├── types/             # Shared type definitions (client, server, transport)
│   └── project/           # Project metadata
├── testdata/              # Test fixtures for integration tests
└── README.md
```

## Related Projects

- [gopls](https://github.com/golang/tools/tree/master/gopls) - Official Go language server
- [MCP](https://modelcontextprotocol.io/) - Model Context Protocol specification
- [mcp-go](https://github.com/mark3labs/mcp-go) - Go implementation of MCP
- [mcp-gopls](https://github.com/hloiseaufcms/mcp-gopls) - MCP server for gopls

## License

MIT
