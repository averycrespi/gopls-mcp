package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// FindSymbolDefinitionsByNameTool handles find symbol definitions by name requests
type FindSymbolDefinitionsByNameTool struct {
	client types.Client
	config types.Config
}

// NewFindSymbolDefinitionsByNameTool creates a new find symbol definitions by name tool
func NewFindSymbolDefinitionsByNameTool(client types.Client, config types.Config) *FindSymbolDefinitionsByNameTool {
	return &FindSymbolDefinitionsByNameTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *FindSymbolDefinitionsByNameTool) GetTool() mcp.Tool {
	tool := mcp.NewTool("find_symbol_definitions_by_name",
		mcp.WithDescription("Find the definition of a symbol by name in Go code, returning a list of symbol definitions"),
		mcp.WithString("symbol_name", mcp.Required(), mcp.Description("Symbol name to find the definition for, with fuzzy matching")),
	)
	return tool
}

// Handle processes the tool request
func (t *FindSymbolDefinitionsByNameTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbolName := mcp.ParseString(req, "symbol_name", "")
	if symbolName == "" {
		return mcp.NewToolResultError("symbol_name parameter is required"), nil
	}

	symbols, err := t.client.FuzzyFindSymbol(ctx, symbolName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search workspace symbols: %v", err)), nil
	}

	toolResult := results.FindSymbolDefinitionsByNameToolResult{
		Name:    symbolName,
		Results: make([]results.SymbolDefinitionResult, 0),
	}
	for _, sym := range symbols {
		definitions, err := t.client.GoToDefinition(ctx, sym.Location.URI, sym.Location.Range.Start)
		if err != nil {
			// Skip definition errors; we'll handle the empty result case later.
			continue
		}

		for _, def := range definitions {
			entry := results.SymbolDefinitionResult{
				Name: sym.Name,
				Kind: results.NewSymbolKind(sym.Kind),
				Location: results.SymbolLocation{
					File:      GetRelativePath(UriToPath(def.URI), t.config.WorkspaceRoot),
					Line:      def.Range.Start.Line + 1,      // convert to 1-indexed line numbers
					Character: def.Range.Start.Character + 1, // convert to 1-indexed character numbers
				},
			}

			// Try to enhance with hover information
			if hoverInfo, err := t.client.GetHoverInfo(ctx, def.URI, def.Range.Start); err == nil && hoverInfo != "" {
				entry.HoverInfo = hoverInfo
			}

			toolResult.Results = append(toolResult.Results, entry)
		}
	}

	if len(toolResult.Results) == 0 {
		toolResult.Message = "No symbol definitions found. " +
			"This could mean that the symbol name is incorrect, or that the symbol is not defined in the workspace."
	} else {
		toolResult.Message = fmt.Sprintf("Found %d symbol definitions.", len(toolResult.Results))
	}

	jsonBytes, err := json.MarshalIndent(toolResult, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
