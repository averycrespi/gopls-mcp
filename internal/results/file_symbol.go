package results

// FileSymbolResult represents a symbol within a file with hierarchical structure
type FileSymbolResult struct {
	Name      string             `json:"name"`
	Kind      SymbolKind         `json:"kind"`
	Location  SymbolLocation     `json:"location"`
	HoverInfo string             `json:"hover_info,omitempty"`
	Children  []FileSymbolResult `json:"children,omitempty"`
}
