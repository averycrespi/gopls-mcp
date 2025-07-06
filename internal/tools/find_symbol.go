package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// FindSymbolTool handles symbol search requests
type FindSymbolTool struct {
	client types.Client
	config types.Config
}

// NewFindSymbolTool creates a new symbol search tool
func NewFindSymbolTool(client types.Client, config types.Config) *FindSymbolTool {
	return &FindSymbolTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *FindSymbolTool) GetTool() mcp.Tool {
	tool := mcp.NewTool(ToolFindSymbol,
		mcp.WithDescription("Find symbols in Go code by name"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Symbol name or pattern to search for")),
	)
	return tool
}

// Handle processes the tool request
func (t *FindSymbolTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := mcp.ParseString(req, "query", "")
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	symbols, err := t.client.FindSymbol(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to find symbols: %v", err)), nil
	}

	if len(symbols) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No symbols found matching '%s'", query)), nil
	}

	var results []string
	for _, sym := range symbols {
		results = append(results, fmt.Sprintf("Symbol: %s (kind: %d) at %s:%d:%d", 
			sym.Name, sym.Kind, sym.Location.URI, sym.Location.Range.Start.Line, sym.Location.Range.Start.Character))
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d symbol(s) matching '%s':\n- %s", 
		len(symbols), query, strings.Join(results, "\n- "))), nil
}