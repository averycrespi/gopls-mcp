# gopls-mcp

An MCP (Model Context Protocol) server that exposes the LSP functionality of the gopls language server, enabling LLMs to work with Go projects more effectively.

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

### Build from Source
```bash
git clone <repository-url>
cd gopls-mcp
go build -o bin/gopls-mcp ./cmd/gopls-mcp
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

#### Claude Desktop Integration

1. **Build the server:**
   ```bash
   make build
   ```

2. **Configure Claude Desktop:**
   Add the following to your Claude Desktop configuration file:

   **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`  
   **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

   ```json
   {
     "mcpServers": {
       "gopls-mcp": {
         "command": "/absolute/path/to/gopls-mcp/bin/gopls-mcp",
         "args": ["-workspace-root", "/absolute/path/to/your/go/project"]
       }
     }
   }
   ```

3. **Restart Claude Desktop** to load the new MCP server.

4. **Verify the connection:**
   - Look for the 🔌 icon in Claude Desktop indicating MCP servers are connected
   - You should see "gopls-mcp" listed as an available server
   - The server will provide 6 tools for Go code analysis

#### Claude Code Integration

If you're using Claude Code (CLI), you can add gopls-mcp to your CLAUDE.md file:

```markdown
# Project Configuration

## MCP Servers

This project uses gopls-mcp for enhanced Go development support.

### Setup
1. Build: `make build`
2. The server provides Go language analysis via gopls integration
3. Available tools: go_to_definition, find_references, hover_info, get_completion, format_code, rename_symbol

### Usage
Ask Claude to use the Go analysis tools when working with .go files in this project.
```

#### Example Usage

Once configured, you can ask Claude to:
- "Find where the `ExampleFunction` is defined"
- "Show me all references to this variable"
- "Get completion suggestions at this position"
- "Format this Go file"
- "Rename this function across the entire project"

The tools work with file paths relative to your workspace root.

## Available MCP Tools

### go_to_definition
Find where a symbol is defined.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)  
- `character` (number): Character position (0-based)

### find_references
Find all references to a symbol.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

### hover_info
Get hover information for a symbol.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

### get_completion
Get code completion suggestions.

**Parameters:**
- `file_path` (string): Path to the Go file
- `line` (number): Line number (0-based)
- `character` (number): Character position (0-based)

### format_code
Format Go code using gofmt.

**Parameters:**
- `file_path` (string): Path to the Go file

### rename_symbol
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
├── cmd/gopls-mcp/          # Main application
├── internal/
│   ├── server/             # MCP server implementation
│   └── lsp/               # LSP client wrapper
├── pkg/types/             # Shared types
└── README.md
```

### Dependencies
- [github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP server framework

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Add license information]

## Related Projects

- [gopls](https://github.com/golang/tools/tree/master/gopls) - Official Go language server
- [MCP](https://modelcontextprotocol.io/) - Model Context Protocol specification
- [mcp-go](https://github.com/mark3labs/mcp-go) - Go implementation of MCP