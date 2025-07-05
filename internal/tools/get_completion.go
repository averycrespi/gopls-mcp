package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"gopls-mcp/pkg/types"
)

// GetCompletionTool handles code completion requests
type GetCompletionTool struct {
	lspClient types.LSPClient
	config    *types.Config
}

// NewGetCompletionTool creates a new code completion tool
func NewGetCompletionTool(lspClient types.LSPClient, config *types.Config) *GetCompletionTool {
	return &GetCompletionTool{
		lspClient: lspClient,
		config:    config,
	}
}

// GetTool returns the MCP tool definition
func (t *GetCompletionTool) GetTool() *mcp.Tool {
	tool := mcp.NewTool(ToolGetCompletion,
		mcp.WithDescription("Get code completion suggestions for Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	return &tool
}

// Handle processes the tool request
func (t *GetCompletionTool) Handle(ctx context.Context, lspClient types.LSPClient, config *types.Config, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	completions, err := lspClient.GetCompletion(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get completions: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d completion(s): %+v", len(completions), completions)), nil
}
