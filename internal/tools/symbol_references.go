package tools

import (
	"context"
	"fmt"

	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// SymbolReferencesTool handles symbol-references requests
type SymbolReferencesTool struct {
	client types.Client
	config types.Config
}

// NewSymbolReferencesTool creates a new symbol-references tool
func NewSymbolReferencesTool(client types.Client, config types.Config) *SymbolReferencesTool {
	return &SymbolReferencesTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *SymbolReferencesTool) GetTool() mcp.Tool {
	tool := mcp.NewTool("symbol_references",
		mcp.WithDescription("Find all references to a symbol in Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	return tool
}

// Handle processes the tool request
func (t *SymbolReferencesTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	position, err := GetPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	uri := PathToUri(filePath, t.config.WorkspaceRoot)
	locations, err := t.client.FindReferences(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to find references: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d reference(s): %+v", len(locations), locations)), nil
}
