package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// FileSymbolsTool handles file symbols requests
type FileSymbolsTool struct {
	client types.Client
	config types.Config
}

// NewFileSymbolsTool creates a new file symbols tool
func NewFileSymbolsTool(client types.Client, config types.Config) *FileSymbolsTool {
	return &FileSymbolsTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *FileSymbolsTool) GetTool() mcp.Tool {
	tool := mcp.NewTool("file_symbols",
		mcp.WithDescription("Get all symbols in a Go file"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
	)
	return tool
}

// Handle processes the tool request
func (t *FileSymbolsTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	uri := PathToUri(filePath, t.config.WorkspaceRoot)
	documentSymbols, err := t.client.GetDocumentSymbols(ctx, uri)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get document symbols: %v", err)), nil
	}

	// Convert DocumentSymbol to FileSymbolResult
	var symbolResults []results.FileSymbolResult
	for _, docSym := range documentSymbols {
		symbolResult := t.convertDocumentSymbol(ctx, uri, docSym, filePath)
		symbolResults = append(symbolResults, symbolResult)
	}

	if len(symbolResults) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No symbols found in file: %s", filePath)), nil
	}

	jsonBytes, err := json.MarshalIndent(symbolResults, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// convertDocumentSymbol converts a DocumentSymbol to FileSymbolResult recursively
func (t *FileSymbolsTool) convertDocumentSymbol(ctx context.Context, uri string, docSym types.DocumentSymbol, filePath string) results.FileSymbolResult {
	result := results.FileSymbolResult{
		Name: docSym.Name,
		Kind: results.NewSymbolKind(docSym.Kind),
		Location: results.SymbolLocation{
			File:      GetRelativePath(UriToPath(PathToUri(filePath, t.config.WorkspaceRoot)), t.config.WorkspaceRoot),
			Line:      docSym.SelectionRange.Start.Line + 1,      // convert to 1-indexed line numbers
			Character: docSym.SelectionRange.Start.Character + 1, // convert to 1-indexed character numbers
		},
	}

	// Try to enhance with hover information
	if hoverInfo, hoverErr := t.client.GetHoverInfo(ctx, uri, docSym.SelectionRange.Start); hoverErr == nil && hoverInfo != "" {
		result.HoverInfo = hoverInfo
	}

	// Convert children recursively
	if len(docSym.Children) > 0 {
		result.Children = make([]results.FileSymbolResult, len(docSym.Children))
		for i, child := range docSym.Children {
			result.Children[i] = t.convertDocumentSymbol(ctx, uri, child, filePath)
		}
	}

	return result
}
