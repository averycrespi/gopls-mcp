package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

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
		mcp.WithDescription("Find all references to a symbol by its anchor in the Go workspace, returning a list of symbol references"),
		mcp.WithString(
			"symbol_anchor",
			mcp.Required(),
			mcp.Description("Symbol anchor, which is included in tool responses. Don't try to parse or generate this yourself."),
		),
	)
	return tool
}

// Handle processes the tool request
func (t *FindSymbolReferencesByAnchorTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	anchorStr := mcp.ParseString(req, "symbol_anchor", "")
	if anchorStr == "" {
		slog.Debug("MCP tool called with missing symbol_anchor parameter", "tool", "find_symbol_references_by_anchor")
		return mcp.NewToolResultError("symbol_anchor parameter is required"), nil
	}

	slog.Debug("MCP tool called", "tool", "find_symbol_references_by_anchor", "symbol_anchor", anchorStr)

	// Parse and validate the anchor
	anchor := results.SymbolAnchor(anchorStr)
	file, position, err := anchor.ToFilePosition()
	if err != nil {
		slog.Debug("Invalid anchor format",
			"tool", "find_symbol_references_by_anchor",
			"symbol_anchor", anchorStr,
			"error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Invalid anchor format: %v", err)), nil
	}

	slog.Debug("Parsed symbol anchor",
		"tool", "find_symbol_references_by_anchor",
		"symbol_anchor", anchorStr,
		"file", file,
		"line", position.Line,
		"character", position.Character)

	uri := PathToUri(file, t.config.WorkspaceRoot)
	refLocations, err := t.client.FindReferences(ctx, uri, position)
	if err != nil {
		slog.Error("Failed to find references",
			"tool", "find_symbol_references_by_anchor",
			"symbol_anchor", anchorStr,
			"uri", uri,
			"error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Failed to find references for anchor %s: %v", anchorStr, err),
		), nil
	}

	slog.Debug("Found references from LSP",
		"tool", "find_symbol_references_by_anchor",
		"symbol_anchor", anchorStr,
		"reference_count", len(refLocations))

	toolResult := results.FindSymbolReferencesByAnchorToolResult{
		Arguments: results.FindSymbolReferencesByAnchorToolArgs{
			SymbolAnchor: anchorStr,
		},
		References: make([]results.SymbolReference, 0),
	}

	for _, refLoc := range refLocations {
		symbolLoc := results.SymbolLocation{
			File:        GetRelativePath(UriToPath(refLoc.URI), t.config.WorkspaceRoot),
			DisplayLine: refLoc.Range.Start.Line + 1,      // Convert LSP coordinates to display line
			DisplayChar: refLoc.Range.Start.Character + 1, // Convert LSP coordinates to display character
		}
		toolResult.References = append(toolResult.References, results.SymbolReference{
			Location: symbolLoc,
			Anchor:   symbolLoc.ToAnchor(),
		})
	}

	if len(toolResult.References) == 0 {
		toolResult.Message = "No references found for the symbol anchor. " +
			"This could mean that the symbol has no references, or that your symbol anchor is out of date. " +
			"You can try getting a fresh symbol anchor from another tool."
		slog.Debug("No references found",
			"tool", "find_symbol_references_by_anchor",
			"symbol_anchor", anchorStr)
	} else {
		toolResult.Message = fmt.Sprintf("Found %d references for the symbol anchor.", len(toolResult.References))
		slog.Debug("Found symbol references",
			"tool", "find_symbol_references_by_anchor",
			"symbol_anchor", anchorStr,
			"reference_count", len(toolResult.References))
	}

	jsonBytes, err := json.MarshalIndent(toolResult, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal tool result",
			"tool", "find_symbol_references_by_anchor",
			"symbol_anchor", anchorStr,
			"error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tool result into JSON: %v", err)), nil
	}

	slog.Debug("MCP tool completed successfully",
		"tool", "find_symbol_references_by_anchor",
		"symbol_anchor", anchorStr,
		"reference_count", len(toolResult.References))

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
