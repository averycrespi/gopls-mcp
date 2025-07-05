package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"gopls-mcp/pkg/types"
)

// FindReferencesTool handles find-references requests
type FindReferencesTool struct {
	lspClient types.LSPClient
	config    *types.Config
}

// NewFindReferencesTool creates a new find-references tool
func NewFindReferencesTool(lspClient types.LSPClient, config *types.Config) *FindReferencesTool {
	return &FindReferencesTool{
		lspClient: lspClient,
		config:    config,
	}
}

// GetTool returns the MCP tool definition
func (t *FindReferencesTool) GetTool() *mcp.Tool {
	tool := mcp.NewTool(ToolFindReferences,
		mcp.WithDescription("Find all references to a symbol in Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	return &tool
}

// Handle processes the tool request
func (t *FindReferencesTool) Handle(ctx context.Context, lspClient types.LSPClient, config *types.Config, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if lspClient == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}

	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	position, err := getPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	uri := getFileURI(filePath, config.WorkspaceRoot)
	locations, err := lspClient.FindReferences(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to find references: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d reference(s): %+v", len(locations), locations)), nil
}