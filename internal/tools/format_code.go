package tools

import (
	"context"
	"fmt"

	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// FormatCodeTool handles code formatting requests
type FormatCodeTool struct {
	client types.Client
	config *types.Config
}

// NewFormatCodeTool creates a new code formatting tool
func NewFormatCodeTool(client types.Client, config *types.Config) *FormatCodeTool {
	return &FormatCodeTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *FormatCodeTool) GetTool() mcp.Tool {
	tool := mcp.NewTool(ToolFormatCode,
		mcp.WithDescription("Format Go code using gofmt"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
	)
	return tool
}

// Handle processes the tool request
func (t *FormatCodeTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	uri := getFileURI(filePath, t.config.WorkspaceRoot)
	edits, err := t.client.FormatDocument(ctx, uri)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format document: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Formatting complete. Applied %d edit(s): %+v", len(edits), edits)), nil
}
