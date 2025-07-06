package tools

import (
	"context"
	"fmt"

	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetCompletionTool handles code completion requests
type GetCompletionTool struct {
	client types.Client
	config types.Config
}

// NewGetCompletionTool creates a new code completion tool
func NewGetCompletionTool(client types.Client, config types.Config) *GetCompletionTool {
	return &GetCompletionTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *GetCompletionTool) GetTool() mcp.Tool {
	tool := mcp.NewTool(ToolGetCompletion,
		mcp.WithDescription("Get code completion suggestions for Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	return tool
}

// Handle processes the tool request
func (t *GetCompletionTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}

	position, err := GetPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	uri := PathToUri(filePath, t.config.WorkspaceRoot)
	completions, err := t.client.GetCompletion(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get completions: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d completion(s): %+v", len(completions), completions)), nil
}
