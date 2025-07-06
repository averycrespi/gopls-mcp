package results

// ListSymbolsInFileToolResult represents the result of the list_symbols_in_file tool
type ListSymbolsInFileToolResult struct {
	FilePath    string       `json:"file_path"`
	Message     string       `json:"message"`
	FileSymbols []FileSymbol `json:"file_symbols,omitempty"`
}

// FileSymbol represents a symbol within a file with hierarchical structure
type FileSymbol struct {
	Name      string         `json:"name"`
	Kind      SymbolKind     `json:"kind"`
	Location  SymbolLocation `json:"location"`
	Anchor    SymbolAnchor   `json:"anchor"`
	HoverInfo string         `json:"hover_info,omitempty"`
	Children  []FileSymbol   `json:"children,omitempty"`
}
