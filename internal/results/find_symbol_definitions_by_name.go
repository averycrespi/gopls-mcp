package results

// FindSymbolDefinitionsByNameToolResult represents the result of the find symbol definitions by name tool
type FindSymbolDefinitionsByNameToolResult struct {
	Message     string                             `json:"message"`
	Arguments   FindSymbolDefinitionByNameToolArgs `json:"arguments"`
	Definitions []SymbolDefinition                 `json:"definitions,omitempty"`
}

// FindSymbolDefinitionByNameToolArgs represents the arguments for the find symbol definitions by name tool
type FindSymbolDefinitionByNameToolArgs struct {
	SymbolName string `json:"symbol_name"`
}

// SymbolDefinition represents a symbol definition result
type SymbolDefinition struct {
	Name      string         `json:"name"`
	Kind      SymbolKind     `json:"kind"`
	Location  SymbolLocation `json:"location"`
	Anchor    SymbolAnchor   `json:"anchor"`
	HoverInfo string         `json:"hover_info,omitempty"`
}
