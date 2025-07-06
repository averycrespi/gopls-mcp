package results

// ListSymbolsInFileToolResult represents the result of the list_symbols_in_file tool
type ListSymbolsInFileToolResult struct {
	Message     string                    `json:"message"`
	Arguments   ListSymbolsInFileToolArgs `json:"arguments"`
	FileSymbols []FileSymbol              `json:"file_symbols,omitempty"`
}

// ListSymbolsInFileToolArgs represents the arguments for the list symbols in file tool
type ListSymbolsInFileToolArgs struct {
	FilePath string `json:"file_path"`
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
