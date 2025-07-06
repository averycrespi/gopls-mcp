# gopls-mcp

An MCP (Model Context Protocol) server that exposes the LSP functionality of the [gopls](https://pkg.go.dev/golang.org/x/tools/gopls) language server.

---

> ⚠️ WARNING: This project is under active development. Breaking changes may be released without notice.

> ⚠️ WARNING: MCP servers execute with full filesystem access and can run arbitrary commands. Only use trusted servers and review code before installation. This server executes gopls and reads Go source files within the specified workspace.

## Purpose

Working with Go code requires understanding its semantic structure, not just text patterns. Traditional text-based operations like grep and find are insufficient for Go development because:

- **Semantic Understanding**: Go symbols have context-dependent meanings. A function name `Add` could refer to different functions across packages, methods on different types, or even variables.
- **Precision**: Text search returns false positives and misses semantic relationships like interface implementations, type aliases, and embedded fields.
- **Scope Awareness**: Go's scoping rules (package, function, block) determine symbol visibility and meaning, which text search cannot understand.
- **Type Information**: Go's type system provides rich metadata (methods, fields, interfaces) that's essential for code navigation and understanding.

gopls-mcp bridges this gap by leveraging the Go language server's semantic analysis capabilities:

- **Symbol Resolution**: Finds exact symbol definitions, not just text matches
- **Reference Tracking**: Identifies all actual references to a symbol, excluding false positives
- **Hierarchical Structure**: Understands Go's package/type/method hierarchy
- **Type-Aware Navigation**: Follows Go's semantic relationships between symbols

This enables LLMs to work with Go code the same way IDEs do - with full semantic understanding rather than pattern matching.

## Tools

| Tool                               | Purpose                                           | Input                                        | Output                                                  |
| ---------------------------------- | ------------------------------------------------- | -------------------------------------------- | ------------------------------------------------------- |
| `list_symbols_in_file`             | List all symbols in a Go file with hierarchy      | `file_path` (string)                         | Hierarchical list of file symbols                       |
| `find_symbol_definitions_by_name`  | Find symbol definitions by name with fuzzy search | `symbol_name` (string)                       | List of symbol definitions which fuzzily-match the name |
| `find_symbol_references_by_anchor` | Find all references to a specific symbol instance | `symbol_anchor` (string)                     | List of symbol references for the anchor                |
| `rename_symbol_by_anchor`          | Rename a symbol across the entire workspace       | `symbol_anchor` (string), `new_name` (string) | List of name changes per file (old→new)                 |

All tools return structured JSON responses with precise location information and symbol anchors for disambiguation.

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
git clone https://github.com/averycrespi/gopls-mcp
cd gopls-mcp
make build
```

## Usage

### Command Line Options
```bash
./bin/gopls-mcp [flags]

Flags:
      --gopls-path string       Path to the gopls binary (default "gopls")
      --log-level string        Log level (debug, info, warn, error) (default "info")
      --workspace-root string   Root directory of the Go workspace (default ".")
  -h, --help                    help for gopls-mcp
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

### Note: Symbol Anchors

Symbol anchors provide a precise way to identify specific symbol instances in Go code. They use the format:
```
go://FILE#LINE:CHAR
```

Where:
- `FILE`: Relative path to the file from workspace root
- `LINE`: Display line number (starts at 1, matches editor display)
- `CHAR`: Display character position (starts at 1, matches editor display)

**Example:** `go://calculator.go#6:6`

Anchors use display coordinates that match what you see in your editor. They are included in all symbol results and enable precise reference finding without ambiguity when multiple symbols share the same name.

### Tool: find_symbol_definitions_by_name
Find the definitions of a symbol by name in the Go workspace, returning a list of symbol definitions with fuzzy search.

**Parameters:**
- `symbol_name` (string): Symbol name to find the definitions for, with fuzzy matching

**Response:** JSON object containing:
- `message`: Summary message about the results (e.g., "Found 3 symbol definitions in the Go workspace." or "No symbol definitions found in the Go workspace.")
- `arguments`: Input arguments echoed back with:
  - `symbol_name`: The searched symbol name
- `definitions`: Array of symbol definition objects (may be empty), each containing:
  - `name`: Symbol name
  - `kind`: Symbol type (function, struct, method, etc.)
  - `location`: File path, line, and character position
  - `anchor`: Symbol anchor in format `go://FILE#LINE:CHAR` (display coordinates)
  - `hover_info`: Hover information from the language server (if available)

### Tool: list_symbols_in_file
List all symbols in a Go file, returning a list of symbols with hierarchical structure.

**Parameters:**
- `file_path` (string): Path to the Go file

**Response:** JSON object containing:
- `message`: Summary message about the results (e.g., "Found 8 symbols in file." or "No symbols found in file.")
- `arguments`: Input arguments echoed back with:
  - `file_path`: The path to the analyzed file
- `file_symbols`: Array of file symbol objects (may be empty), each containing:
  - `name`: Symbol name
  - `kind`: Symbol type (function, struct, method, etc.)
  - `location`: File path, line, and character position
  - `anchor`: Symbol anchor in format `go://FILE#LINE:CHAR` (display coordinates)
  - `hover_info`: Hover information from the language server (if available)
  - `children`: Array of child symbols (for hierarchical symbols like structs with fields, methods, etc.)

The tool provides full hierarchical support for Go symbols. For example:
- Struct symbols include their fields and methods as children
- Interface symbols include their method signatures as children
- Function symbols may include nested function declarations as children

This hierarchical structure is enabled by the LSP client's `hierarchicalDocumentSymbolSupport` capability.

### Tool: find_symbol_references_by_anchor
Find all references to a symbol by its precise anchor location in the Go workspace.

**Parameters:**
- `symbol_anchor` (string): Symbol anchor in format `go://FILE#LINE:CHAR` (display coordinates)

**Response:** JSON object containing:
- `message`: Summary message about the results (e.g., "Found 8 references for the symbol anchor." or "No references found for the symbol anchor.")
- `arguments`: Input arguments echoed back with:
  - `symbol_anchor`: The input symbol anchor used for the search
- `references`: Array of reference objects, each containing:
  - `location`: Reference location with:
    - `file`: Relative file path from workspace root
    - `line`: Display line number (starts at 1, matches editor display)
    - `character`: Display character position (starts at 1, matches editor display)
  - `anchor`: Symbol anchor for this specific reference location in format `go://FILE#LINE:CHAR`

**Note:** This tool requires a precise anchor from the output of `find_symbol_definitions_by_name` or `list_symbols_in_file` tools to identify the exact symbol instance.

### Tool: rename_symbol_by_anchor
Rename a symbol by its precise anchor location across the entire Go workspace.

**Parameters:**
- `symbol_anchor` (string): Symbol anchor in format `go://FILE#LINE:CHAR` (display coordinates)
- `new_name` (string): New name for the symbol (must be a valid Go identifier)

**Response:** JSON object containing:
- `message`: Summary message about the results (e.g., "Successfully renamed symbol to 'NewName' with 5 edits across 3 files.")
- `arguments`: Input arguments echoed back with:
  - `symbol_anchor`: The input symbol anchor used for the rename
  - `new_name`: The new name for the symbol
- `file_edits`: Array of file edit objects (may be empty), each containing:
  - `file`: Relative file path from workspace root
  - `edits`: Array of edit objects, each with:
    - `start_line`: Display line number where edit starts (1-indexed)
    - `start_character`: Display character position where edit starts (1-indexed)
    - `end_line`: Display line number where edit ends (1-indexed)
    - `end_character`: Display character position where edit ends (1-indexed)
    - `old_text`: The text being replaced
    - `new_text`: The replacement text

**Notes:** 
- This tool uses gopls's rename functionality which includes validation to prevent breaking changes
- The rename will fail if it would introduce compilation errors
- Go keywords cannot be used as new names
- The tool performs a prepareRename check first to ensure the rename is valid

## Development

For detailed information about the architecture, design patterns, and development guidelines, please see [DEVELOPERS.md](DEVELOPERS.md).

## Related Projects

- [gopls](https://github.com/golang/tools/tree/master/gopls) - Official Go language server
- [MCP](https://modelcontextprotocol.io/) - Model Context Protocol specification
- [mcp-go](https://github.com/mark3labs/mcp-go) - Go implementation of MCP
- [mcp-gopls](https://github.com/hloiseaufcms/mcp-gopls) - MCP server for gopls

## License

MIT
