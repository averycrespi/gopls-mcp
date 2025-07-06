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

const (
	// DefaultDefinitionsLimit is the default maximum number of symbol definitions to return
	DefaultDefinitionsLimit = 50
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
		mcp.WithDescription("Find the definitions of a symbol by name in the Go workspace, returning a list of symbol definitions"),
		mcp.WithString("symbol_name", mcp.Required(), mcp.Description("Symbol name to find the definitions for, with fuzzy matching")),
		mcp.WithNumber("limit", mcp.Description(fmt.Sprintf("Maximum number of symbol definitions to return (default: %d)", DefaultDefinitionsLimit))),
		mcp.WithBoolean("include_hover", mcp.Description("Whether to include hover information for symbols (default: false)")),
	)
	return tool
}

// Handle processes the tool request
func (t *FindSymbolDefinitionsByNameTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbolName := mcp.ParseString(req, "symbol_name", "")
	if symbolName == "" {
		slog.Debug("MCP tool called with missing symbol_name parameter", "tool", "find_symbol_definitions_by_name")
		return mcp.NewToolResultError("symbol_name parameter is required"), nil
	}

	limit := mcp.ParseInt(req, "limit", DefaultDefinitionsLimit)
	if limit <= 0 {
		limit = DefaultDefinitionsLimit
	}

	includeHover := mcp.ParseBoolean(req, "include_hover", false)

	slog.Debug("MCP tool called", "tool", "find_symbol_definitions_by_name", "symbol_name", symbolName, "limit", limit, "include_hover", includeHover)

	symbols, err := t.client.FuzzyFindSymbol(ctx, symbolName)
	if err != nil {
		slog.Error("Failed to search workspace symbols",
			"tool", "find_symbol_definitions_by_name",
			"symbol_name", symbolName,
			"error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Failed to search Go workspace symbols for symbol name: %s: %v", symbolName, err),
		), nil
	}

	slog.Debug("Found workspace symbols",
		"tool", "find_symbol_definitions_by_name",
		"symbol_name", symbolName,
		"symbol_count", len(symbols))

	toolResult := results.FindSymbolDefinitionsByNameToolResult{
		Arguments: results.FindSymbolDefinitionByNameToolArgs{
			SymbolName:   symbolName,
			Limit:        limit,
			IncludeHover: includeHover,
		},
		Definitions: make([]results.SymbolDefinition, 0),
	}
	for _, sym := range symbols {
		defLocations, err := t.client.GoToDefinition(ctx, sym.Location.URI, sym.Location.Range.Start)
		if err != nil {
			// Skip definition errors; we'll handle the empty result case later
			continue
		}

		for _, loc := range defLocations {
			location := results.SymbolLocation{
				File:        GetRelativePath(UriToPath(loc.URI), t.config.WorkspaceRoot),
				DisplayLine: loc.Range.Start.Line + 1,      // Convert LSP coordinates to display line
				DisplayChar: loc.Range.Start.Character + 1, // Convert LSP coordinates to display character
			}
			entry := results.SymbolDefinition{
				Name:     sym.Name,
				Kind:     results.NewSymbolKind(sym.Kind),
				Location: location,
				Anchor:   location.ToAnchor(),
			}

			// Try to enhance with hover information if requested
			if includeHover {
				if hoverInfo, err := t.client.GetHoverInfo(ctx, loc.URI, loc.Range.Start); err == nil && hoverInfo != "" {
					entry.HoverInfo = hoverInfo
				}
			}

			toolResult.Definitions = append(toolResult.Definitions, entry)

			// Apply limit to prevent token overflow
			if len(toolResult.Definitions) >= limit {
				break
			}
		}

		// Break outer loop if limit reached
		if len(toolResult.Definitions) >= limit {
			break
		}
	}

	if len(toolResult.Definitions) == 0 {
		toolResult.Message = "No symbol definitions found in the Go workspace. " +
			"This could mean that the symbol name is incorrect, or that the symbol is not defined in the workspace."
		slog.Debug("No definitions found",
			"tool", "find_symbol_definitions_by_name",
			"symbol_name", symbolName)
	} else {
		toolResult.Message = fmt.Sprintf("Found %d symbol definitions in the Go workspace.", len(toolResult.Definitions))
		slog.Debug("Found symbol definitions",
			"tool", "find_symbol_definitions_by_name",
			"symbol_name", symbolName,
			"definition_count", len(toolResult.Definitions))
	}

	jsonBytes, err := json.Marshal(toolResult)
	if err != nil {
		slog.Error("Failed to marshal tool result",
			"tool", "find_symbol_definitions_by_name",
			"symbol_name", symbolName,
			"error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tool result into JSON: %v", err)), nil
	}

	slog.Debug("MCP tool completed successfully",
		"tool", "find_symbol_definitions_by_name",
		"symbol_name", symbolName,
		"definition_count", len(toolResult.Definitions),
		"response_size_bytes", len(jsonBytes))

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
