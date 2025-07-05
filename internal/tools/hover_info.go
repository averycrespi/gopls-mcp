package tools

import (
	"context"
	"fmt"

	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// HoverInfoTool handles hover-info requests
type HoverInfoTool struct {
	client types.Client
	config *types.Config
}

// NewHoverInfoTool creates a new hover-info tool
func NewHoverInfoTool(client types.Client, config *types.Config) *HoverInfoTool {
	return &HoverInfoTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *HoverInfoTool) GetTool() mcp.Tool {
	tool := mcp.NewTool(ToolHoverInfo,
		mcp.WithDescription("Get hover information for a symbol in Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	return tool
}

// Handle processes the tool request
func (t *HoverInfoTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	position, err := getPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	uri := getFileURI(filePath, t.config.WorkspaceRoot)
	hover, err := t.client.Hover(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get hover info: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Hover info: %s", hover)), nil
}
