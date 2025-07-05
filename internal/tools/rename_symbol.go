package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"gopls-mcp/pkg/types"
)

// RenameSymbolTool handles symbol renaming requests
type RenameSymbolTool struct {
	lspClient types.LSPClient
	config    *types.Config
}

// NewRenameSymbolTool creates a new symbol renaming tool
func NewRenameSymbolTool(lspClient types.LSPClient, config *types.Config) *RenameSymbolTool {
	return &RenameSymbolTool{
		lspClient: lspClient,
		config:    config,
	}
}

// GetTool returns the MCP tool definition
func (t *RenameSymbolTool) GetTool() *mcp.Tool {
	tool := mcp.NewTool(ToolRenameSymbol,
		mcp.WithDescription("Rename a symbol across the Go project"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
		mcp.WithString("new_name", mcp.Required(), mcp.Description("New name for the symbol")),
	)
	return &tool
}

// Handle processes the tool request
func (t *RenameSymbolTool) Handle(ctx context.Context, lspClient types.LSPClient, config *types.Config, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if lspClient == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}

	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	newName := mcp.ParseString(req, "new_name", "")
	if newName == "" {
		return mcp.NewToolResultError("new_name parameter is required"), nil
	}

	position, err := getPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	uri := getFileURI(filePath, config.WorkspaceRoot)
	changes, err := lspClient.RenameSymbol(ctx, uri, position, newName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to rename symbol: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Rename complete. Changed %d file(s): %+v", len(changes), changes)), nil
}