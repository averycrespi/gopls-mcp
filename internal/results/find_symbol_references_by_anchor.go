package results

// SymbolReferenceResult represents a symbol reference result
type SymbolReferenceResult struct {
	Name       string           `json:"name"`
	Kind       SymbolKind       `json:"kind"`
	Location   SymbolLocation   `json:"location"`
	Anchor     SymbolAnchor     `json:"anchor"`
	HoverInfo  string           `json:"hover_info,omitempty"`
	References []SymbolLocation `json:"references"`
}
