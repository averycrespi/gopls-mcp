package results

// FindSymbolReferencesByAnchorToolResult represents the result of the find symbol references by anchor tool
type FindSymbolReferencesByAnchorToolResult struct {
	Message    string                               `json:"message"`
	Arguments  FindSymbolReferencesByAnchorToolArgs `json:"arguments"`
	References []SymbolReference                    `json:"references,omitempty"`
}

// FindSymbolReferencesByAnchorToolArgs represents the arguments for the find symbol references by anchor tool
type FindSymbolReferencesByAnchorToolArgs struct {
	SymbolAnchor string `json:"symbol_anchor"`
	Limit        int    `json:"limit,omitempty"`
}

// SymbolReference represents a symbol reference
type SymbolReference struct {
	Location SymbolLocation `json:"location"`
	Anchor   SymbolAnchor   `json:"anchor"`
}
