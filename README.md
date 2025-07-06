# gopls-mcp

An MCP (Model Context Protocol) server that exposes the LSP functionality of the [gopls](https://pkg.go.dev/golang.org/x/tools/gopls) language server, enabling LLMs to work with Go projects more effectively.

## Overview

gopls-mcp bridges the gap between LLMs and Go development by providing an MCP interface to gopls (the official Go language server). This allows AI assistants to perform sophisticated Go code analysis, navigation, and refactoring operations.

## Features

### Navigation Tools
- **symbol_definition**: Find the definition of a symbol by name
- **find_references**: Find all references to a symbol across the project


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

All tools return structured JSON responses for easy programmatic consumption.

### symbol_definition
Find the definition of a symbol by name.

**Parameters:**
- `symbol_name` (string): Symbol name to find the definition for

**Response:** JSON object containing:
- `query`: The search query used
- `count`: Number of symbols found
- `symbols`: Array of symbol definition entries with location, documentation, and source code context

### find_references
Find all references to a symbol.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)


## Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────┐
│   MCP Client    │◄──►│   gopls-mcp  │◄──►│    gopls    │
│   (Claude, etc) │    │   (Server)   │    │ (Language   │
└─────────────────┘    └──────────────┘    │  Server)    │
                                           └─────────────┘
```

The server acts as a bridge:
1. GoplsServer.Serve() receives MCP tool calls from clients
2. Tools translate requests to LSP operations via GoplsClient interface
3. GoplsClient communicates with gopls via JsonRpcTransport
4. Returns formatted responses to the MCP client

### Design Principles
- **Interface-based design**: Clean separation of concerns through well-defined interfaces
- **Value types for immutability**: Config is passed as value type to prevent modification
- **Direct tool injection**: Tools receive client reference at creation time for simplicity

## Development

### Project Structure
```
gopls-mcp/
├── cmd/gopls-mcp/         # Main application entry point
├── internal/
│   ├── server/            # MCP server implementation (GoplsServer)
│   ├── client/            # LSP client implementation (GoplsClient)
│   ├── transport/         # JSON-RPC transport layer (JsonRpcTransport)
│   ├── tools/             # Individual MCP tool implementations
│   └── results/           # JSON response types and formatting
├── pkg/
│   ├── types/             # Shared type definitions (client, server, config, transport)
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
