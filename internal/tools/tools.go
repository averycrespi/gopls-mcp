package tools

import (
	"path/filepath"
	"strings"

	"github.com/averycrespi/gopls-mcp/pkg/types"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool names
const (
	ToolGoToDefinition = "go_to_definition"
	ToolFindReferences = "find_references"
	ToolHoverInfo      = "hover_info"
	ToolGetCompletion  = "get_completion"
	ToolFormatCode     = "format_code"
	ToolRenameSymbol   = "rename_symbol"
)

// getFileURI converts a file path to a file URI
func getFileURI(filePath string, workspaceRoot string) string {
	if strings.HasPrefix(filePath, "file://") {
		return filePath
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workspaceRoot, filePath)
	}

	return "file://" + filePath
}

// getPosition extracts position from MCP request
func getPosition(req mcp.CallToolRequest) (types.Position, error) {
	line := mcp.ParseFloat64(req, "line", 0)
	character := mcp.ParseFloat64(req, "character", 0)

	return types.Position{
		Line:      int(line),
		Character: int(character),
	}, nil
}
