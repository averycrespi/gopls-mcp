package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
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
	)
	return tool
}

// Handle processes the tool request
func (t *ListSymbolsInFileTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	uri := PathToUri(filePath, t.config.WorkspaceRoot)
	documentSymbols, err := t.client.GetDocumentSymbols(ctx, uri)
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("Failed to get document symbols for file: %s: %v", filePath, err),
		), nil
	}

	toolResult := results.ListSymbolsInFileToolResult{
		Arguments: results.ListSymbolsInFileToolArgs{
			FilePath: filePath,
		},
		FileSymbols: make([]results.FileSymbol, 0),
	}
	for _, docSym := range documentSymbols {
		symbolResult := t.convertDocumentSymbol(ctx, uri, docSym, filePath)
		toolResult.FileSymbols = append(toolResult.FileSymbols, symbolResult)
	}
	if len(toolResult.FileSymbols) == 0 {
		toolResult.Message = "No symbols found in file. " +
			"This could mean that the file is missing, empty, or not a Go file."
	} else {
		toolResult.Message = fmt.Sprintf("Found %d symbols in file.", len(toolResult.FileSymbols))
	}

	jsonBytes, err := json.MarshalIndent(toolResult, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tool result JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// convertDocumentSymbol converts a DocumentSymbol to FileSymbol recursively
func (t *ListSymbolsInFileTool) convertDocumentSymbol(ctx context.Context, uri string, docSym types.DocumentSymbol, filePath string) results.FileSymbol {
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

	// Try to enhance with hover information
	if hoverInfo, hoverErr := t.client.GetHoverInfo(ctx, uri, docSym.SelectionRange.Start); hoverErr == nil && hoverInfo != "" {
		result.HoverInfo = hoverInfo
	}

	// Convert children recursively
	if len(docSym.Children) > 0 {
		result.Children = make([]results.FileSymbol, len(docSym.Children))
		for i, child := range docSym.Children {
			result.Children[i] = t.convertDocumentSymbol(ctx, uri, child, filePath)
		}
	}

	return result
}
