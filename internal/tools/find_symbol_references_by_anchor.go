package tools

import (
	"context"
	"encoding/json"
	"fmt"

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
		mcp.WithDescription("Find all references to a symbol by its anchor in the Go workspace"),
		mcp.WithString("anchor", mcp.Required(), mcp.Description("Symbol anchor in format 'anchor://FILE#LINE:CHAR' (1-indexed coordinates)")),
	)
	return tool
}

// Handle processes the tool request
func (t *FindSymbolReferencesByAnchorTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	anchorStr := mcp.ParseString(req, "anchor", "")
	if anchorStr == "" {
		return mcp.NewToolResultError("anchor parameter is required"), nil
	}

	// Parse and validate the anchor
	anchor := results.SymbolAnchor(anchorStr)
	file, line, character, err := anchor.ToLSPPosition()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid anchor format: %v", err)), nil
	}

	// Convert to absolute path and URI for LSP
	absPath := PathToUri(file, t.config.WorkspaceRoot)
	position := types.Position{
		Line:      line,      // Already converted to 0-indexed by ToLSPPosition
		Character: character, // Already converted to 0-indexed by ToLSPPosition
	}

	// Find references at the specific anchor location
	locations, err := t.client.FindReferences(ctx, absPath, position)
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("Failed to find references for anchor %s: %v", anchorStr, err),
		), nil
	}

	// Convert reference locations to SymbolLocation format
	var references []results.SymbolLocation
	for _, loc := range locations {
		references = append(references, results.SymbolLocation{
			File:      GetRelativePath(UriToPath(loc.URI), t.config.WorkspaceRoot),
			Line:      loc.Range.Start.Line + 1,      // convert to 1-indexed line numbers
			Character: loc.Range.Start.Character + 1, // convert to 1-indexed character numbers
		})
	}

	// Get symbol information at the anchor location for metadata
	hoverInfo, _ := t.client.GetHoverInfo(ctx, absPath, position)

	// Try to get the symbol at this location to determine name and kind
	symbols, symbolErr := t.client.GetDocumentSymbols(ctx, absPath)
	var symbolName string = "Unknown"
	var symbolKind results.SymbolKind = "unknown"

	if symbolErr == nil {
		// Find the symbol that contains this position
		for _, docSym := range symbols {
			if t.containsPosition(docSym, position) {
				symbolName = docSym.Name
				symbolKind = results.NewSymbolKind(docSym.Kind)
				break
			}
		}
	}

	// Create the result with the original anchor location (convert back to 1-indexed for display)
	originalLocation := results.SymbolLocation{
		File:      file,
		Line:      line + 1,      // convert back to 1-indexed for display
		Character: character + 1, // convert back to 1-indexed for display
	}

	result := results.SymbolReferenceResult{
		Name:       symbolName,
		Kind:       symbolKind,
		Location:   originalLocation,
		Anchor:     anchor,
		HoverInfo:  hoverInfo,
		References: references,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tool result into JSON: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// containsPosition checks if a DocumentSymbol contains the given position
func (t *FindSymbolReferencesByAnchorTool) containsPosition(docSym types.DocumentSymbol, pos types.Position) bool {
	// Check if position is within the symbol's range
	start := docSym.Range.Start
	end := docSym.Range.End

	// Position is within range if:
	// - line is after start line, or
	// - line equals start line and character >= start character, and
	// - line is before end line, or
	// - line equals end line and character <= end character
	if pos.Line < start.Line || pos.Line > end.Line {
		return false
	}
	if pos.Line == start.Line && pos.Character < start.Character {
		return false
	}
	if pos.Line == end.Line && pos.Character > end.Character {
		return false
	}

	// Check children first (more specific)
	for _, child := range docSym.Children {
		if t.containsPosition(child, pos) {
			return false // Child contains it, so this parent doesn't
		}
	}

	return true
}
