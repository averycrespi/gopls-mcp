package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
)

// FindSymbolReferencesByAnchorTool handles find symbol references by anchor requests
type FindSymbolReferencesByAnchorTool struct {
	client types.Client
	config types.Config
}

// NewFindSymbolReferencesByAnchorTool creates a new find symbol references by anchor tool
func NewFindSymbolReferencesByAnchorTool(client types.Client, config types.Config) *FindSymbolReferencesByAnchorTool {
	return &FindSymbolReferencesByAnchorTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *FindSymbolReferencesByAnchorTool) GetTool() mcp.Tool {
	tool := mcp.NewTool("find_symbol_references_by_anchor",
		mcp.WithDescription("Find all references to a symbol by anchor name in the Go workspace"),
		mcp.WithString("symbol_name", mcp.Required(), mcp.Description("Symbol name to find references for (case-insensitive matching)")),
	)
	return tool
}

// Handle processes the tool request
func (t *FindSymbolReferencesByAnchorTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbolName := mcp.ParseString(req, "symbol_name", "")
	if symbolName == "" {
		return mcp.NewToolResultError("symbol_name parameter is required"), nil
	}

	symbols, err := t.client.FuzzyFindSymbol(ctx, symbolName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search Go workspace symbols for anchor: %s: %v", symbolName, err)), nil
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

		// Try to enhance with hover information
		if hoverInfo, err := t.client.GetHoverInfo(ctx, sym.Location.URI, sym.Location.Range.Start); err == nil && hoverInfo != "" {
			entry.HoverInfo = hoverInfo
		}

		symbolResults = append(symbolResults, entry)
	}

	if len(symbolResults) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No references found for symbol anchor: %s", symbolName)), nil
	}

	jsonBytes, err := json.MarshalIndent(symbolResults, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tool result into JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
