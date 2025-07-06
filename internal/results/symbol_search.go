package results

// SymbolSearchResult represents the JSON structure for symbol search results
type SymbolSearchResult struct {
	Query   string                    `json:"query"`
	Count   int                       `json:"count"`
	Symbols []SymbolSearchResultEntry `json:"symbols"`
}

// SymbolSearchResultEntry represents a single symbol in the search results
type SymbolSearchResultEntry struct {
	Name          string         `json:"name"`
	Kind          SymbolKind     `json:"kind"`
	Location      SymbolLocation `json:"location"`
	Documentation string         `json:"documentation,omitempty"`
	Source        *SourceContext `json:"source,omitempty"`
}
