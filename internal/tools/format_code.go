package tools

import (
	"context"
	"fmt"

	"gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// FormatCodeTool handles code formatting requests
type FormatCodeTool struct {
	lspClient types.LSPClient
	config    *types.Config
}

// NewFormatCodeTool creates a new code formatting tool
func NewFormatCodeTool(lspClient types.LSPClient, config *types.Config) *FormatCodeTool {
	return &FormatCodeTool{
		lspClient: lspClient,
		config:    config,
	}
}

// GetTool returns the MCP tool definition
func (t *FormatCodeTool) GetTool() *mcp.Tool {
	tool := mcp.NewTool(ToolFormatCode,
		mcp.WithDescription("Format Go code using gofmt"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
	)
	return &tool
}

// Handle processes the tool request
func (t *FormatCodeTool) Handle(ctx context.Context, lspClient types.LSPClient, config *types.Config, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if lspClient == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}

	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	uri := getFileURI(filePath, config.WorkspaceRoot)
	edits, err := lspClient.FormatDocument(ctx, uri)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format document: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Formatting complete. Applied %d edit(s): %+v", len(edits), edits)), nil
}
