package results

// SymbolDefinitionResult represents a symbol definition result
type SymbolDefinitionResult struct {
	Name      string         `json:"name"`
	Kind      SymbolKind     `json:"kind"`
	Location  SymbolLocation `json:"location"`
	HoverInfo string         `json:"hover_info,omitempty"`
}
