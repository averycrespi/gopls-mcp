package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/averycrespi/gopls-mcp/internal/results"
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
	tool := mcp.NewTool("symbol_definition",
		mcp.WithDescription("Find the definition of a symbol in Go code"),
		mcp.WithString("symbol_name", mcp.Required(), mcp.Description("Symbol name to find the definition for")),
	)
	return tool
}

// Handle processes the tool request
func (t *SymbolDefinitionTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbolName := mcp.ParseString(req, "symbol_name", "")
	if symbolName == "" {
		return mcp.NewToolResultError("symbol_name parameter is required"), nil
	}

	symbols, err := t.client.FuzzyFindSymbol(ctx, symbolName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search workspace symbols: %v", err)), nil
	}

	if len(symbols) == 0 {
		result := results.SymbolDefinitionResult{
			Query:   symbolName,
			Count:   0,
			Symbols: []results.SymbolDefinitionResultEntry{},
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	// Build JSON result
	result := results.SymbolDefinitionResult{
		Query:   symbolName,
		Count:   len(symbols),
		Symbols: make([]results.SymbolDefinitionResultEntry, 0, len(symbols)),
	}

	for _, sym := range symbols {
		entry := results.SymbolDefinitionResultEntry{
			Name: sym.Name,
			Kind: results.NewSymbolKind(sym.Kind),
			Location: results.SymbolLocation{
				File:      GetRelativePath(UriToPath(sym.Location.URI), t.config.WorkspaceRoot),
				Line:      sym.Location.Range.Start.Line + 1,
				Character: sym.Location.Range.Start.Character + 1,
			},
			Definitions: make([]results.SymbolDefinitionInfo, 0),
		}

		// Get the definition for this symbol
		definitions, err := t.client.GoToDefinition(ctx, sym.Location.URI, sym.Location.Range.Start)
		if err != nil {
			// Add error as documentation if we can't get definitions
			entry.Definitions = append(entry.Definitions, results.SymbolDefinitionInfo{
				Location:      entry.Location,
				Documentation: fmt.Sprintf("Error getting definition: %v", err),
			})
		} else {
			for _, def := range definitions {
				defInfo := results.SymbolDefinitionInfo{
					Location: results.SymbolLocation{
						File:      GetRelativePath(UriToPath(def.URI), t.config.WorkspaceRoot),
						Line:      def.Range.Start.Line + 1,
						Character: def.Range.Start.Character + 1,
					},
				}

				// Try to enhance the definition with hover information
				if hoverInfo, hoverErr := t.client.GetHoverInfo(ctx, def.URI, def.Range.Start); hoverErr == nil && hoverInfo != "" {
					defInfo.Documentation = hoverInfo
				}

				// Try to enhance the definition with source context
				if file, err := os.Open(UriToPath(def.URI)); err == nil {
					defer file.Close()
					if sourceContext, contextErr := ReadSourceContext(file, def.Range.Start.Line, 3); contextErr == nil {
						defInfo.Source = sourceContext
					}
				}

				entry.Definitions = append(entry.Definitions, defInfo)
			}
		}

		result.Symbols = append(result.Symbols, entry)
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
