package results

// ListSymbolsInFileToolResult represents the result of the list_symbols_in_file tool
type ListSymbolsInFileToolResult struct {
	FilePath string             `json:"file_path"`
	Message  string             `json:"message"`
	Results  []FileSymbolResult `json:"results,omitempty"`
}

// FileSymbolResult represents a symbol within a file with hierarchical structure
type FileSymbolResult struct {
	Name      string             `json:"name"`
	Kind      SymbolKind         `json:"kind"`
	Location  SymbolLocation     `json:"location"`
	HoverInfo string             `json:"hover_info,omitempty"`
	Children  []FileSymbolResult `json:"children,omitempty"`
}
