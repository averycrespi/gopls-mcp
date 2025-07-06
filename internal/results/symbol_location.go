package results

// SymbolLocation represents the location of a symbol
type SymbolLocation struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}
