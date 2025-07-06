# Developer Guide

This guide provides detailed information for developers contributing to gopls-mcp.

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

## Project Structure
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

## Core Components

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

## Key Design Patterns

### Tool Registration
Each MCP tool is implemented in its own file in `internal/tools/`:
- `find_symbol_definitions_by_name.go` - `find_symbol_definitions_by_name` → LSP WorkspaceSymbol + Definition requests with anchor generation
- `find_symbol_references_by_anchor.go` - `find_symbol_references_by_anchor` → LSP References requests using precise anchor locations
- `list_symbols_in_file.go` - `list_symbols_in_file` → LSP DocumentSymbol requests with hierarchical support and anchor generation
- `rename_symbol_by_anchor.go` - `rename_symbol_by_anchor` → LSP PrepareRename + Rename requests for safe symbol renaming
- `utils.go` - Shared utilities for path handling and position parsing

### JSON Response Structure
Structured output types in `internal/results/`:
- `symbol_kind.go` - SymbolKind enum with LSP mapping (file, function, struct, etc.)
- `symbol_location.go` - Location information with file paths and positions, plus anchor conversion
- `symbol_anchor.go` - SymbolAnchor type for precise symbol identification with format `go://FILE#LINE:CHAR` (1-indexed coordinates)
- `find_symbol_definitions_by_name.go` - FindSymbolDefinitionsByNameToolResult with standardized structure (message, arguments with symbol_name, SymbolDefinition array)
- `find_symbol_references_by_anchor.go` - FindSymbolReferencesByAnchorToolResult with standardized structure (message, arguments with symbol_anchor, SymbolReference array)
- `list_symbols_in_file.go` - ListSymbolsInFileToolResult with standardized structure (message, arguments with file_path, hierarchical FileSymbol array)
- `rename_symbol_by_anchor.go` - RenameSymbolByAnchorToolResult with standardized structure (message, arguments with symbol_anchor and new_name, FileEdit array with detailed edit information)

### Interface Design
The codebase uses clean interfaces to separate concerns:
- `types.Client` - Defines LSP client operations including Start/Stop (implemented by GoplsClient)
- `types.Server` - Defines MCP server operations with Serve method (implemented by GoplsServer)
- `types.Transport` - Defines JSON-RPC transport operations with Start/Stop methods (implemented by JsonRpcTransport)
- `types.Config` - Configuration structure used as value type for better immutability

**Value vs Pointer Types**: The design uses value types for Config to ensure immutability and prevent accidental modifications, while interfaces provide the abstraction layer for different implementations.

### Path Handling
All file paths are converted to absolute paths and file:// URIs for LSP communication.

### Error Handling
LSP errors are wrapped and returned as MCP tool result errors with structured logging and specific error messages.

### Structured Logging
The codebase uses Go's built-in `slog` package for comprehensive logging:
- **JSON format**: All logs are structured JSON sent to stderr for easy parsing
- **Configurable levels**: debug, info, warn, error controlled by `--log-level` flag
- **Contextual information**: Request IDs, timing, file paths, symbol counts, error details
- **Performance metrics**: JSON-RPC request/response timing, gopls process tracking
- **Debug tracing**: Tool execution workflows, LSP method calls, parameter validation
- **Source locations**: File/line information included in debug mode for troubleshooting

### JSON Output
All symbol tools return standardized JSON responses with:
- **Consistent Structure**: `message`, `arguments`, and tool-specific results (`definitions`, `references`, `file_symbols`)
- **Input Echo**: Arguments field echoes back input parameters for validation and debugging
- Type-safe SymbolKind enums (function, struct, method, etc.) where applicable
- Rich metadata including hover info from the language server (for definition tools)
- Relative file paths from workspace root
- Symbol anchors for precise identification (`go://FILE#LINE:CHAR` format)
- Descriptive messages and metadata (e.g., file paths, symbol counts, reference counts)

### Symbol Anchor System
Enables precise symbol identification and eliminates ambiguity:
- Format: `go://FILE#LINE:CHAR` with display coordinates (matches editor display)
- Generated for all SymbolDefinition and FileSymbol results
- Used by `find_symbol_references_by_anchor` for exact reference finding, returning anchors for each reference location
- Converts to LSP coordinates internally for protocol operations via `ToFilePosition()`
- Uses `DisplayLine` and `DisplayChar` fields for clarity throughout codebase
- Validates anchor format and coordinates before processing

### Hierarchical Symbol Support
The `list_symbols_in_file` tool provides full hierarchical support for Go symbols:
- Enabled by declaring `hierarchicalDocumentSymbolSupport: true` in the LSP client capabilities
- Structs include their fields and methods as children
- Interfaces include their method signatures as children
- Functions may include nested function declarations as children
- Recursive hover info collection for all symbol levels

## Development Commands

### Building and Testing
- `make build` - Build the gopls-mcp binary to `bin/gopls-mcp`
- `make test` - Run unit tests
- `make test-integration` - Run integration tests (builds binary, checks dependencies)
- `make run` - Run server (see Running the Server section)
- `make clean` - Clean build artifacts and caches

### MCP Tool Testing
- `make test-find-symbol-definitions-by-name` - Test find_symbol_definitions_by_name tool with pretty-printed JSON output
- `make test-find-symbol-references-by-anchor` - Test find_symbol_references_by_anchor tool with pretty-printed JSON output
- `make test-list-symbols-in-file` - Test list_symbols_in_file tool with pretty-printed JSON output
- `make test-rename-symbol-by-anchor` - Test rename_symbol_by_anchor tool with automatic backup/restore
- Uses `scripts/test-mcp-tool.sh` for JSON extraction and formatting
- Uses `scripts/test-rename-tool.sh` for rename testing with file backup/restore

### Dependencies
- `make deps` - Download and tidy Go modules
- `make install-gopls` - Install gopls language server if not present
- Requires Go 1.23+ and gopls to be available
- `jq` - Required for pretty-printing JSON in MCP tool tests

### Running the Server
- `make run WORKSPACE=/path/to/go/project` - Run with specific workspace
- `go run ./cmd/gopls-mcp --workspace-root /path/to/project` - Direct run
- `go run ./cmd/gopls-mcp --workspace-root /path/to/project --log-level debug` - Run with debug logging
- `go run github.com/averycrespi/gopls-mcp@latest` - Run latest version

## Logging and Debugging

The server uses structured JSON logging to stderr, making it easy to parse and analyze logs programmatically. The `--log-level` flag controls the verbosity:

### Log Levels

- **`error`**: Only critical errors that prevent operation
- **`warn`**: Warning messages (currently unused, reserved for future use)
- **`info`**: High-level operational messages (startup, configuration)
- **`debug`**: Detailed execution traces with timing and context

### Debug Logging Features

When using `--log-level debug`, you get comprehensive visibility into:

**Server Operations:**
- Server startup and tool registration
- Gopls process management (PID tracking)
- MCP tool invocation and completion

**LSP Communication:**
- JSON-RPC request/response timing in milliseconds
- Method names and request IDs for correlation
- Transport lifecycle events

**Tool Execution:**
- Parameter validation and parsing
- Symbol search and resolution steps
- File processing and URI conversion
- Result counts and performance metrics

**Example Debug Output:**
```json
{"time":"2025-07-06T14:40:31.400726-07:00","level":"DEBUG","source":{"function":"github.com/averycrespi/gopls-mcp/internal/server.NewGoplsServer","file":"/Users/avery/Workspace/gopls-mcp/internal/server/server.go","line":27},"msg":"Creating new Gopls MCP server","project_name":"gopls-mcp","project_version":"0.0.1","gopls_path":"gopls","workspace_root":"/Users/avery/Workspace/gopls-mcp"}

{"time":"2025-07-06T14:40:31.402073-07:00","level":"DEBUG","source":{"function":"github.com/averycrespi/gopls-mcp/internal/transport.(*JsonRpcTransport).SendRequest","file":"/Users/avery/Workspace/gopls-mcp/internal/transport/transport.go","line":153},"msg":"Sending JSON-RPC request","request_id":1,"method":"initialize"}
```

### Troubleshooting

For troubleshooting issues:

1. **Connection problems**: Use `--log-level debug` to see JSON-RPC communication
2. **Tool failures**: Debug logs show parameter parsing and LSP interaction
3. **Performance issues**: Check request timing and symbol counts in debug output
4. **Gopls issues**: Monitor gopls process startup and LSP method calls

### Debugging Development Issues

**Using Structured Logging for Development:**
- Use `--log-level debug` when developing or troubleshooting
- Debug logs include source file locations, timing, and detailed context
- JSON-RPC communication is fully logged with request IDs for correlation
- Tool execution shows parameter parsing, LSP interactions, and result processing

**Common Debugging Scenarios:**
- **Tool failures**: Check parameter validation and LSP method calls in debug logs
- **Performance issues**: Monitor request timing and symbol counts
- **LSP communication problems**: Review JSON-RPC request/response sequences
- **Symbol resolution**: Trace workspace symbol searches and definition lookups

**Logging Integration:**
- All components use `slog.Debug()`, `slog.Info()`, `slog.Error()` with structured fields
- Avoid redundant error logging where errors are already returned and handled
- Include relevant context (file paths, symbol names, request IDs) in log messages
- Use consistent field names across components for easier log analysis

## Testing

The project includes both unit and integration tests:

- Unit tests: Focus on individual components
- Integration tests: Verify the full tool workflow including gopls interaction
- Test fixtures in `testdata/` provide sample Go projects for testing

Run all tests with `make test test-integration`.

### Testing Safety

For tools that modify files (like `rename_symbol_by_anchor`), special safety measures are in place:

- **Automatic Backup**: The `test-rename-symbol-by-anchor` target automatically backs up `testdata/example` before running tests
- **Guaranteed Restore**: A bash trap ensures the original files are restored even if the test fails or is interrupted
- **Isolated Testing**: Each test run uses a unique backup directory to avoid conflicts
- **Visual Feedback**: Tests show before/after states and clearly indicate whether changes occurred

The backup/restore mechanism ensures that the test environment remains pristine and tests can be run repeatedly without side effects.
