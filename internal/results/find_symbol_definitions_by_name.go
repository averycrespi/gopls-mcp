package results

// FindSymbolDefinitionsByNameToolResult represents the result of the find symbol definitions by name tool
type FindSymbolDefinitionsByNameToolResult struct {
	SymbolName  string             `json:"symbol_name"`
	Message     string             `json:"message"`
	Definitions []SymbolDefinition `json:"definitions,omitempty"`
}

// SymbolDefinition represents a symbol definition result
type SymbolDefinition struct {
	Name      string         `json:"name"`
	Kind      SymbolKind     `json:"kind"`
	Location  SymbolLocation `json:"location"`
	HoverInfo string         `json:"hover_info,omitempty"`
}
