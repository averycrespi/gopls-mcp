package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// GoToDefinitionTool handles go-to-definition requests
type GoToDefinitionTool struct {
	client types.Client
	config types.Config
}

// NewGoToDefinitionTool creates a new go-to-definition tool
func NewGoToDefinitionTool(client types.Client, config types.Config) *GoToDefinitionTool {
	return &GoToDefinitionTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *GoToDefinitionTool) GetTool() mcp.Tool {
	tool := mcp.NewTool(ToolGoToDefinition,
		mcp.WithDescription("Find the definition of a symbol in Go code"),
		mcp.WithString("symbol", mcp.Required(), mcp.Description("Symbol name to find the definition for")),
	)
	return tool
}

// Handle processes the tool request
func (t *GoToDefinitionTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbol := mcp.ParseString(req, "symbol", "")
	if symbol == "" {
		return mcp.NewToolResultError("symbol parameter is required"), nil
	}

	// First, search for symbols matching the query
	symbols, err := t.client.FindSymbol(ctx, symbol)
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
