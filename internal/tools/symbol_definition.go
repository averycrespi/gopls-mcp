package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// SymbolDefinitionTool handles symbol definition requests
type SymbolDefinitionTool struct {
	client types.Client
	config types.Config
}

// NewSymbolDefinitionTool creates a new symbol definition tool
func NewSymbolDefinitionTool(client types.Client, config types.Config) *SymbolDefinitionTool {
	return &SymbolDefinitionTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *SymbolDefinitionTool) GetTool() mcp.Tool {
	tool := mcp.NewTool(ToolSymbolDefinition,
		mcp.WithDescription("Find the definition of a symbol in Go code"),
		mcp.WithString("symbol", mcp.Required(), mcp.Description("Symbol name to find the definition for")),
	)
	return tool
}

// Handle processes the tool request
func (t *SymbolDefinitionTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbol := mcp.ParseString(req, "symbol", "")
	if symbol == "" {
		return mcp.NewToolResultError("symbol parameter is required"), nil
	}

	symbols, err := t.client.FuzzyFindSymbol(ctx, symbol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search workspace symbols: %v", err)), nil
	}

	if len(symbols) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No symbols found matching '%s'", symbol)), nil
	}

	// For each symbol found, get its definition
	var results []string
	for _, sym := range symbols {
		// Get the definition for this symbol
		definitions, err := t.client.GoToDefinition(ctx, sym.Location.URI, sym.Location.Range.Start)
		if err != nil {
			results = append(results, fmt.Sprintf("Symbol: %s at %s:%d:%d - Error getting definition: %v",
				sym.Name, sym.Location.URI, sym.Location.Range.Start.Line, sym.Location.Range.Start.Character, err))
			continue
		}

		if len(definitions) == 0 {
			results = append(results, fmt.Sprintf("Symbol: %s at %s:%d:%d - No definition found",
				sym.Name, sym.Location.URI, sym.Location.Range.Start.Line, sym.Location.Range.Start.Character))
		} else {
			for _, def := range definitions {
				results = append(results, fmt.Sprintf("Symbol: %s at %s:%d:%d - Definition: %s:%d:%d",
					sym.Name, sym.Location.URI, sym.Location.Range.Start.Line, sym.Location.Range.Start.Character,
					def.URI, def.Range.Start.Line, def.Range.Start.Character))
			}
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d symbol(s) matching '%s':\n- %s",
		len(symbols), symbol, strings.Join(results, "\n- "))), nil
}
