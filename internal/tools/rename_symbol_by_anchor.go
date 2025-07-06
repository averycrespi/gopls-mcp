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

// RenameSymbolByAnchorTool handles rename symbol by anchor requests
type RenameSymbolByAnchorTool struct {
	client types.Client
	config types.Config
}

// NewRenameSymbolByAnchorTool creates a new rename symbol by anchor tool
func NewRenameSymbolByAnchorTool(client types.Client, config types.Config) *RenameSymbolByAnchorTool {
	return &RenameSymbolByAnchorTool{
		client: client,
		config: config,
	}
}

// GetTool returns the MCP tool definition
func (t *RenameSymbolByAnchorTool) GetTool() mcp.Tool {
	tool := mcp.NewTool("rename_symbol_by_anchor",
		mcp.WithDescription("Rename a symbol by its anchor in the Go workspace, returning a list of file edits"),
		mcp.WithString(
			"symbol_anchor",
			mcp.Required(),
			mcp.Description("Symbol anchor, which is included in tool responses. Don't try to parse or generate this yourself."),
		),
		mcp.WithString(
			"new_name",
			mcp.Required(),
			mcp.Description("New name for the symbol"),
		),
	)
	return tool
}

// Handle processes the tool request
func (t *RenameSymbolByAnchorTool) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	anchorStr := mcp.ParseString(req, "symbol_anchor", "")
	if anchorStr == "" {
		slog.Debug("MCP tool called with missing symbol_anchor parameter", "tool", "rename_symbol_by_anchor")
		return mcp.NewToolResultError("symbol_anchor parameter is required"), nil
	}

	newName := mcp.ParseString(req, "new_name", "")
	if newName == "" {
		slog.Debug("MCP tool called with missing new_name parameter", "tool", "rename_symbol_by_anchor")
		return mcp.NewToolResultError("new_name parameter is required"), nil
	}

	if !IsValidGoIdentifier(newName) {
		slog.Debug("Invalid Go identifier provided",
			"tool", "rename_symbol_by_anchor",
			"new_name", newName)
		return mcp.NewToolResultError(fmt.Sprintf("'%s' is not a valid Go identifier", newName)), nil
	}

	slog.Debug("MCP tool called",
		"tool", "rename_symbol_by_anchor",
		"symbol_anchor", anchorStr,
		"new_name", newName)

	anchor := results.SymbolAnchor(anchorStr)
	file, position, err := anchor.ToFilePosition()
	if err != nil {
		slog.Debug("Invalid anchor format",
			"tool", "rename_symbol_by_anchor",
			"symbol_anchor", anchorStr,
			"error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Invalid anchor format: %v", err)), nil
	}

	slog.Debug("Parsed symbol anchor",
		"tool", "rename_symbol_by_anchor",
		"symbol_anchor", anchorStr,
		"file", file,
		"line", position.Line,
		"character", position.Character)

	uri := PathToUri(file, t.config.WorkspaceRoot)
	prepareResult, err := t.client.PrepareRename(ctx, uri, position)
	if err != nil {
		slog.Debug("Failed to prepare rename",
			"tool", "rename_symbol_by_anchor",
			"symbol_anchor", anchorStr,
			"uri", uri,
			"error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Cannot rename at anchor %s: %v", anchorStr, err),
		), nil
	}

	slog.Debug("Rename prepared",
		"tool", "rename_symbol_by_anchor",
		"symbol_anchor", anchorStr,
		"range", prepareResult.Range,
		"placeholder", prepareResult.Placeholder)

	workspaceEdit, err := t.client.RenameSymbol(ctx, uri, position, newName)
	if err != nil {
		slog.Error("Failed to rename symbol",
			"tool", "rename_symbol_by_anchor",
			"symbol_anchor", anchorStr,
			"new_name", newName,
			"uri", uri,
			"error", err)
		return mcp.NewToolResultError(
			fmt.Sprintf("Failed to rename symbol at anchor %s: %v", anchorStr, err),
		), nil
	}

	// Calculate total affected files from both formats
	affectedFiles := 0
	if workspaceEdit.Changes != nil {
		affectedFiles += len(workspaceEdit.Changes)
	}
	if workspaceEdit.DocumentChanges != nil {
		affectedFiles += len(workspaceEdit.DocumentChanges)
	}

	slog.Debug("Symbol renamed",
		"tool", "rename_symbol_by_anchor",
		"symbol_anchor", anchorStr,
		"new_name", newName,
		"affected_files", affectedFiles)

	toolResult := results.RenameSymbolByAnchorToolResult{
		Arguments: results.RenameSymbolByAnchorToolArgs{
			SymbolAnchor: anchorStr,
			NewName:      newName,
		},
		FileEdits: make([]results.FileEdit, 0),
	}

	// Collect all name changes by file, deduplicating as we go
	fileNameChanges := make(map[string]map[string]string) // file -> oldName -> newName

	// Process Changes format (legacy)
	for fileURI, textEdits := range workspaceEdit.Changes {
		filePath := GetRelativePath(UriToPath(fileURI), t.config.WorkspaceRoot)
		if fileNameChanges[filePath] == nil {
			fileNameChanges[filePath] = make(map[string]string)
		}

		for _, textEdit := range textEdits {
			oldText := prepareResult.Placeholder
			if oldText == "" {
				// Fallback: try to extract old text from the range if possible
				oldText = textEdit.NewText // This is not ideal, but better than empty
			}
			fileNameChanges[filePath][oldText] = textEdit.NewText
		}
	}

	// Process DocumentChanges format (modern)
	for _, docEdit := range workspaceEdit.DocumentChanges {
		filePath := GetRelativePath(UriToPath(docEdit.TextDocument.URI), t.config.WorkspaceRoot)
		if fileNameChanges[filePath] == nil {
			fileNameChanges[filePath] = make(map[string]string)
		}

		for _, textEdit := range docEdit.Edits {
			oldText := prepareResult.Placeholder
			if oldText == "" {
				// Fallback: try to extract old text from the range if possible
				oldText = textEdit.NewText // This is not ideal, but better than empty
			}
			fileNameChanges[filePath][oldText] = textEdit.NewText
		}
	}

	// Convert to FileEdit structure
	for filePath, nameChanges := range fileNameChanges {
		fileEdit := results.FileEdit{
			File:    filePath,
			Changes: make([]results.NameChange, 0, len(nameChanges)),
		}

		for oldName, newName := range nameChanges {
			// Only add if old and new names are different
			if oldName != newName {
				fileEdit.Changes = append(fileEdit.Changes, results.NameChange{
					OldName: oldName,
					NewName: newName,
				})
			}
		}

		// Only add file edit if there are actual changes
		if len(fileEdit.Changes) > 0 {
			toolResult.FileEdits = append(toolResult.FileEdits, fileEdit)
		}
	}

	if len(toolResult.FileEdits) == 0 {
		toolResult.Message = "No changes needed to rename symbol. The symbol may already have this name."
		slog.Debug("No changes needed",
			"tool", "rename_symbol_by_anchor",
			"symbol_anchor", anchorStr,
			"new_name", newName)
	} else {
		totalChanges := 0
		for _, fe := range toolResult.FileEdits {
			totalChanges += len(fe.Changes)
		}
		toolResult.Message = fmt.Sprintf("Successfully renamed symbol with %d name changes across %d files.",
			totalChanges, len(toolResult.FileEdits))
		slog.Debug("Rename completed",
			"tool", "rename_symbol_by_anchor",
			"symbol_anchor", anchorStr,
			"new_name", newName,
			"file_count", len(toolResult.FileEdits),
			"change_count", totalChanges)
	}

	jsonBytes, err := json.MarshalIndent(toolResult, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal tool result",
			"tool", "rename_symbol_by_anchor",
			"symbol_anchor", anchorStr,
			"error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tool result into JSON: %v", err)), nil
	}

	slog.Debug("MCP tool completed successfully",
		"tool", "rename_symbol_by_anchor",
		"symbol_anchor", anchorStr,
		"new_name", newName,
		"file_count", len(toolResult.FileEdits))

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
