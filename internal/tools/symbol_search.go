package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// SymbolSearchTool handles symbol search requests
type SymbolSearchTool struct {
	client types.Client
	config types.Config
}

// NewSymbolSearchTool creates a new symbol search tool
func NewSymbolSearchTool(client types.Client, config types.Config) *SymbolSearchTool {
	return &SymbolSearchTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *SymbolSearchTool) GetTool() mcp.Tool {
	tool := mcp.NewTool(ToolSymbolSearch,
		mcp.WithDescription("Search for symbols in Go code by name"),
		mcp.WithString("symbol", mcp.Required(), mcp.Description("Symbol name to search for")),
	)
	return tool
}

// Handle processes the tool request
func (t *SymbolSearchTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbol := mcp.ParseString(req, "symbol", "")
	if symbol == "" {
		return mcp.NewToolResultError("symbol parameter is required"), nil
	}

	symbols, err := t.client.FuzzyFindSymbol(ctx, symbol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to find symbols: %v", err)), nil
	}

	if len(symbols) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No symbols found matching '%s'", symbol)), nil
	}

	var results []string
	for _, sym := range symbols {
		results = append(results, fmt.Sprintf("Symbol: %s (kind: %d) at %s:%d:%d",
			sym.Name, sym.Kind, sym.Location.URI, sym.Location.Range.Start.Line, sym.Location.Range.Start.Character))
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d symbol(s) matching '%s':\n- %s",
		len(symbols), symbol, strings.Join(results, "\n- "))), nil
}
