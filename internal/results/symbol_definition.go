package results

// SymbolDefinitionResult represents the JSON structure for symbol definition results
type SymbolDefinitionResult struct {
	Query   string                        `json:"query"`
	Count   int                           `json:"count"`
	Symbols []SymbolDefinitionResultEntry `json:"symbols"`
}

// SymbolDefinitionResultEntry represents a single symbol definition in the results
type SymbolDefinitionResultEntry struct {
	Name        string                 `json:"name"`
	Kind        SymbolKind             `json:"kind"`
	Location    SymbolLocation         `json:"location"`
	Definitions []SymbolDefinitionInfo `json:"definitions"`
}

// SymbolDefinitionInfo represents information about a symbol definition
type SymbolDefinitionInfo struct {
	Location      SymbolLocation `json:"location"`
	Documentation string         `json:"documentation,omitempty"`
	Source        *SourceContext `json:"source,omitempty"`
}
