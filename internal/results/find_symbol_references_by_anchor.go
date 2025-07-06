package results

// FindSymbolReferencesByAnchorToolResult represents the result of the find symbol references by anchor tool
type FindSymbolReferencesByAnchorToolResult struct {
	SymbolAnchor string            `json:"symbol_anchor"`
	Message      string            `json:"message"`
	References   []SymbolReference `json:"references"`
}

// SymbolReference represents a symbol reference
type SymbolReference struct {
	Location SymbolLocation `json:"location"`
	Anchor   SymbolAnchor   `json:"anchor"`
}
