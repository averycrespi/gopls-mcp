package tools

import (
	"context"
	"fmt"

	"gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// GoToDefinitionTool handles go-to-definition requests
type GoToDefinitionTool struct {
	lspClient types.LSPClient
	config    *types.Config
}

// NewGoToDefinitionTool creates a new go-to-definition tool
func NewGoToDefinitionTool(lspClient types.LSPClient, config *types.Config) *GoToDefinitionTool {
	return &GoToDefinitionTool{
		lspClient: lspClient,
		config:    config,
	}
}

// GetTool returns the MCP tool definition
func (t *GoToDefinitionTool) GetTool() *mcp.Tool {
	tool := mcp.NewTool(ToolGoToDefinition,
		mcp.WithDescription("Find the definition of a symbol in Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	return &tool
}

// Handle processes the tool request
func (t *GoToDefinitionTool) Handle(ctx context.Context, lspClient types.LSPClient, config *types.Config, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	locations, err := lspClient.GoToDefinition(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get definition: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d definition(s): %+v", len(locations), locations)), nil
}
