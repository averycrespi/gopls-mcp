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

const (
	definitionContextLines = 3
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

	var symbolResults []results.SymbolDefinitionResult
	for _, sym := range symbols {
		definitions, err := t.client.GoToDefinition(ctx, sym.Location.URI, sym.Location.Range.Start)
		if err != nil || len(definitions) == 0 {
			continue
		}

		// Use the first definition matching the symbol
		def := definitions[0]
		entry := results.SymbolDefinitionResult{
			Name: sym.Name,
			Kind: results.NewSymbolKind(sym.Kind),
			Location: results.SymbolLocation{
				File:      GetRelativePath(UriToPath(def.URI), t.config.WorkspaceRoot),
				Line:      def.Range.Start.Line + 1,      // convert to 1-indexed line numbers
				Character: def.Range.Start.Character + 1, // convert to 1-indexed character numbers
			},
		}


		// Try to enhance with source context
		if file, err := os.Open(UriToPath(def.URI)); err == nil {
			defer file.Close()
			if sourceContext, contextErr := ReadSourceContext(file, def.Range.Start.Line, definitionContextLines); contextErr == nil {
				entry.Source = sourceContext
			}
		}

		symbolResults = append(symbolResults, entry)
	}

	if len(symbolResults) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No results for symbol name: %s", symbolName)), nil
	}

	jsonBytes, err := json.MarshalIndent(symbolResults, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
