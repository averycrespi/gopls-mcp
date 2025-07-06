package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	// DefaultFileSymbolsLimit is the default maximum number of file symbols to return
	DefaultFileSymbolsLimit = 100
)

// ListSymbolsInFileTool handles list symbols in file requests
type ListSymbolsInFileTool struct {
	client types.Client
	config types.Config
}

// NewListSymbolsInFileTool creates a new list symbols in file tool
func NewListSymbolsInFileTool(client types.Client, config types.Config) *ListSymbolsInFileTool {
	return &ListSymbolsInFileTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *ListSymbolsInFileTool) GetTool() mcp.Tool {
	tool := mcp.NewTool("list_symbols_in_file",
		mcp.WithDescription("List all symbols in a Go file, returning a list of symbols with hierarchical structure"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("limit", mcp.Description(fmt.Sprintf("Maximum number of symbols to return (default: %d)", DefaultFileSymbolsLimit))),
		mcp.WithBoolean("include_hover", mcp.Description("Whether to include hover information for symbols (default: false)")),
	)
	return tool
}

// Handle processes the tool request
func (t *ListSymbolsInFileTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		slog.Debug("MCP tool called with missing file_path parameter", "tool", "list_symbols_in_file")
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	limit := mcp.ParseInt(req, "limit", DefaultFileSymbolsLimit)
	if limit <= 0 {
		limit = DefaultFileSymbolsLimit
	}

	includeHover := mcp.ParseBoolean(req, "include_hover", false)

	slog.Debug("MCP tool called", "tool", "list_symbols_in_file", "file_path", filePath, "limit", limit, "include_hover", includeHover)

	uri := PathToUri(filePath, t.config.WorkspaceRoot)
	slog.Debug("Converted file path to URI",
		"tool", "list_symbols_in_file",
		"file_path", filePath,
		"uri", uri)

	documentSymbols, err := t.client.GetDocumentSymbols(ctx, uri)
	if err != nil {
		slog.Error("Failed to get document symbols",
			"tool", "list_symbols_in_file",
			"file_path", filePath,
			"uri", uri,
			"error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Failed to get document symbols for file: %s: %v", filePath, err),
		), nil
	}

	slog.Debug("Found document symbols from LSP",
		"tool", "list_symbols_in_file",
		"file_path", filePath,
		"symbol_count", len(documentSymbols))

	toolResult := results.ListSymbolsInFileToolResult{
		Arguments: results.ListSymbolsInFileToolArgs{
			FilePath:     filePath,
			Limit:        limit,
			IncludeHover: includeHover,
		},
		FileSymbols: make([]results.FileSymbol, 0),
	}
	for _, docSym := range documentSymbols {
		// Apply limit to prevent token overflow
		if len(toolResult.FileSymbols) >= limit {
			break
		}

		symbolResult := t.convertDocumentSymbol(ctx, uri, docSym, filePath, includeHover)
		toolResult.FileSymbols = append(toolResult.FileSymbols, symbolResult)
	}
	if len(toolResult.FileSymbols) == 0 {
		toolResult.Message = "No symbols found in file. " +
			"This could mean that the file is missing, empty, or not a Go file."
		slog.Debug("No symbols found in file",
			"tool", "list_symbols_in_file",
			"file_path", filePath)
	} else {
		toolResult.Message = fmt.Sprintf("Found %d symbols in file.", len(toolResult.FileSymbols))
		slog.Debug("Found file symbols",
			"tool", "list_symbols_in_file",
			"file_path", filePath,
			"symbol_count", len(toolResult.FileSymbols))
	}

	jsonBytes, err := json.Marshal(toolResult)
	if err != nil {
		slog.Error("Failed to marshal tool result",
			"tool", "list_symbols_in_file",
			"file_path", filePath,
			"error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tool result JSON: %v", err)), nil
	}

	slog.Debug("MCP tool completed successfully",
		"tool", "list_symbols_in_file",
		"file_path", filePath,
		"symbol_count", len(toolResult.FileSymbols),
		"response_size_bytes", len(jsonBytes))

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// convertDocumentSymbol converts a DocumentSymbol to FileSymbol recursively
func (t *ListSymbolsInFileTool) convertDocumentSymbol(ctx context.Context, uri string, docSym types.DocumentSymbol, filePath string, includeHover bool) results.FileSymbol {
	location := results.SymbolLocation{
		File:        GetRelativePath(UriToPath(PathToUri(filePath, t.config.WorkspaceRoot)), t.config.WorkspaceRoot),
		DisplayLine: docSym.SelectionRange.Start.Line + 1,      // Convert LSP coordinates to display line
		DisplayChar: docSym.SelectionRange.Start.Character + 1, // Convert LSP coordinates to display character
	}
	result := results.FileSymbol{
		Name:     docSym.Name,
		Kind:     results.NewSymbolKind(docSym.Kind),
		Location: location,
		Anchor:   location.ToAnchor(),
	}

	// Try to enhance with hover information if requested
	if includeHover {
		if hoverInfo, hoverErr := t.client.GetHoverInfo(ctx, uri, docSym.SelectionRange.Start); hoverErr == nil && hoverInfo != "" {
			result.HoverInfo = hoverInfo
		}
	}

	// Convert children recursively
	if len(docSym.Children) > 0 {
		result.Children = make([]results.FileSymbol, len(docSym.Children))
		for i, child := range docSym.Children {
			result.Children[i] = t.convertDocumentSymbol(ctx, uri, child, filePath, includeHover)
		}
	}

	return result
}
