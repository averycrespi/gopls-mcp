package results

// SymbolLocation represents the location of a symbol.
// Unlike types.Location, it contains a file (not a URI) and is 1-indexed (not 0-indexed).
type SymbolLocation struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// ToAnchor creates a SymbolAnchor from this location (coordinates remain 1-indexed)
func (sl SymbolLocation) ToAnchor() SymbolAnchor {
	return NewSymbolAnchor(sl.File, sl.Line, sl.Character)
}
