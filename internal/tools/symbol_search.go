package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/averycrespi/gopls-mcp/internal/results"
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
	tool := mcp.NewTool("symbol_search",
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
		result := results.SymbolSearchResult{
			Query:   symbol,
			Count:   0,
			Symbols: []results.SymbolSearchResultEntry{},
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	result := results.SymbolSearchResult{
		Query:   symbol,
		Count:   len(symbols),
		Symbols: make([]results.SymbolSearchResultEntry, 0, len(symbols)),
	}

	for _, sym := range symbols {
		entry := results.SymbolSearchResultEntry{
			Name: sym.Name,
			Kind: results.NewSymbolKind(sym.Kind),
			Location: results.SymbolLocation{
				File:      getRelativePath(uriToPath(sym.Location.URI), t.config.WorkspaceRoot),
				Line:      sym.Location.Range.Start.Line + 1,
				Character: sym.Location.Range.Start.Character + 1,
			},
		}

		// Try to enhance the result with hover information
		if hoverInfo, err := t.client.GetHoverInfo(ctx, sym.Location.URI, sym.Location.Range.Start); err == nil && hoverInfo != "" {
			entry.Documentation = hoverInfo
		}

		// Try to enhance the result with source context
		if sourceContext, err := getSymbolContext(sym.Location.URI, sym.Location.Range.Start.Line, sym.Location.Range.Start.Character, 2); err == nil {
			entry.Source = sourceContext
		}

		result.Symbols = append(result.Symbols, entry)
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
