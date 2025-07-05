package tools

// Tool name prefix for all MCP tools
const ToolPrefix = "gopls."

// Tool names
const (
	ToolGoToDefinition = ToolPrefix + "go_to_definition"
	ToolFindReferences = ToolPrefix + "find_references"
	ToolHoverInfo      = ToolPrefix + "hover_info"
	ToolGetCompletion  = ToolPrefix + "get_completion"
	ToolFormatCode     = ToolPrefix + "format_code"
	ToolRenameSymbol   = ToolPrefix + "rename_symbol"
)
