package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"gopls-mcp/pkg/types"
)

// HoverInfoTool handles hover-info requests
type HoverInfoTool struct {
	lspClient types.LSPClient
	config    *types.Config
}

// NewHoverInfoTool creates a new hover-info tool
func NewHoverInfoTool(lspClient types.LSPClient, config *types.Config) *HoverInfoTool {
	return &HoverInfoTool{
		lspClient: lspClient,
		config:    config,
	}
}

// GetTool returns the MCP tool definition
func (t *HoverInfoTool) GetTool() *mcp.Tool {
	tool := mcp.NewTool("gopls.hover_info",
		mcp.WithDescription("Get hover information for a symbol in Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	return &tool
}

// Handle processes the tool request
func (t *HoverInfoTool) Handle(ctx context.Context, lspClient types.LSPClient, config *types.Config, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	hover, err := lspClient.Hover(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get hover info: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Hover info: %s", hover)), nil
}