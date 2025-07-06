package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	referenceContextLines = 0 // disable extra context for now
)

// SymbolReferencesTool handles symbol-references requests
type SymbolReferencesTool struct {
	client types.Client
	config types.Config
}

// NewSymbolReferencesTool creates a new symbol-references tool
func NewSymbolReferencesTool(client types.Client, config types.Config) *SymbolReferencesTool {
	return &SymbolReferencesTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *SymbolReferencesTool) GetTool() mcp.Tool {
	tool := mcp.NewTool("symbol_references",
		mcp.WithDescription("Find all references to a symbol in Go code"),
		mcp.WithString("symbol_name", mcp.Required(), mcp.Description("Symbol name to find references for")),
	)
	return tool
}

// Handle processes the tool request
func (t *SymbolReferencesTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbolName := mcp.ParseString(req, "symbol_name", "")
	if symbolName == "" {
		return mcp.NewToolResultError("symbol_name parameter is required"), nil
	}

	symbols, err := t.client.FuzzyFindSymbol(ctx, symbolName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search workspace symbols: %v", err)), nil
	}

	var symbolResults []results.SymbolReferenceResult
	for _, sym := range symbols {
		// Only return results for the exact symbol name, case insensitive
		if !strings.EqualFold(sym.Name, symbolName) {
			continue
		}

		locations, err := t.client.FindReferences(ctx, sym.Location.URI, sym.Location.Range.Start)
		if err != nil {
			continue
		}

		var references []results.SymbolLocation
		for _, loc := range locations {
			references = append(references, results.SymbolLocation{
				File:      GetRelativePath(UriToPath(loc.URI), t.config.WorkspaceRoot),
				Line:      loc.Range.Start.Line + 1,      // convert to 1-indexed line numbers
				Character: loc.Range.Start.Character + 1, // convert to 1-indexed character numbers
			})
		}

		entry := results.SymbolReferenceResult{
			Name: sym.Name,
			Kind: results.NewSymbolKind(sym.Kind),
			Location: results.SymbolLocation{
				File:      GetRelativePath(UriToPath(sym.Location.URI), t.config.WorkspaceRoot),
				Line:      sym.Location.Range.Start.Line + 1,      // convert to 1-indexed line numbers
				Character: sym.Location.Range.Start.Character + 1, // convert to 1-indexed character numbers
			},
			References: references,
		}

		// Try to enhance with source context
		if file, err := os.Open(UriToPath(sym.Location.URI)); err == nil {
			defer file.Close()
			if sourceContext, contextErr := ReadSourceContext(file, sym.Location.Range.Start.Line, referenceContextLines); contextErr == nil {
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
