package results

// SymbolLocation represents the location of a symbol.
// Unlike types.Location, it contains a file (not a URI) and uses display coordinates (not LSP coordinates).
// Display coordinates start at line 1, character 1 (matching editor display).
type SymbolLocation struct {
	File        string `json:"file"`      // Relative file path
	DisplayLine int    `json:"line"`      // Display line (starts at 1)
	DisplayChar int    `json:"character"` // Display character (starts at 1)
}

// ToAnchor creates a SymbolAnchor from this location using display coordinates
func (sl SymbolLocation) ToAnchor() SymbolAnchor {
	return NewSymbolAnchor(sl.File, sl.DisplayLine, sl.DisplayChar)
}
