# gopls-mcp

An MCP (Model Context Protocol) server that exposes the LSP functionality of the [gopls](https://pkg.go.dev/golang.org/x/tools/gopls) language server, enabling LLMs to work with Go projects more effectively.

## Tools

gopls-mcp provides the following tools:
- **find_symbol_definitions_by_name**: Find the definition of a symbol by name.
- **find_symbol_references_by_anchor**: Find all references to a symbol by anchor name.
- **list_symbols_in_file**: Get all symbols in a Go file.

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

### find_symbol_definitions_by_name
Find the definitions of a symbol by name in the Go workspace, returning a list of symbol definitions with fuzzy search.

**Parameters:**
- `symbol_name` (string): Symbol name to find the definitions for, with fuzzy matching

**Response:** JSON object containing:
- `symbol_name`: The searched symbol name
- `message`: Summary message about the results (e.g., "Found 3 symbol definitions in the Go workspace." or "No symbol definitions found in the Go workspace.")
- `definitions`: Array of symbol definition objects (may be empty), each containing:
  - `name`: Symbol name
  - `kind`: Symbol type (function, struct, method, etc.)
  - `location`: File path, line, and character position
  - `hover_info`: Hover information from the language server (if available)

### find_symbol_references_by_anchor
Find all references to a symbol by anchor name in the Go workspace.

**Parameters:**
- `symbol_name` (string): Symbol name to find references for (case-insensitive matching)

**Response:** JSON array of symbol reference objects, each containing:
- `name`: Symbol name
- `kind`: Symbol type (function, struct, method, etc.)
- `location`: File path, line, and character position of the symbol definition
- `hover_info`: Hover information from the language server (if available)
- `references`: Array of locations where the symbol is referenced

### list_symbols_in_file
List all symbols in a Go file, returning a list of symbols with hierarchical structure.

**Parameters:**
- `file_path` (string): Path to the Go file

**Response:** JSON object containing:
- `file_path`: The path to the analyzed file
- `message`: Summary message about the results (e.g., "Found 8 symbols in file." or "No symbols found in file.")
- `file_symbols`: Array of file symbol objects (may be empty), each containing:
  - `name`: Symbol name
  - `kind`: Symbol type (function, struct, method, etc.)
  - `location`: File path, line, and character position
  - `hover_info`: Hover information from the language server (if available)
  - `children`: Array of child symbols (for hierarchical symbols like structs with fields, methods, etc.)

The tool provides full hierarchical support for Go symbols. For example:
- Struct symbols include their fields and methods as children
- Interface symbols include their method signatures as children
- Function symbols may include nested function declarations as children

This hierarchical structure is enabled by the LSP client's `hierarchicalDocumentSymbolSupport` capability.


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
