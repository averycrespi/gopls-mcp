package results

// FindSymbolDefinitionsByNameToolResult represents the result of the find symbol definitions by name tool
type FindSymbolDefinitionsByNameToolResult struct {
	Name    string                   `json:"name"`
	Message string                   `json:"message"`
	Results []SymbolDefinitionResult `json:"results,omitempty"`
}

// SymbolDefinitionResult represents a symbol definition result
type SymbolDefinitionResult struct {
	Name      string         `json:"name"`
	Kind      SymbolKind     `json:"kind"`
	Location  SymbolLocation `json:"location"`
	HoverInfo string         `json:"hover_info,omitempty"`
}
