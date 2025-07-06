package results

// SymbolReferenceResult represents a symbol reference result
type SymbolReferenceResult struct {
	Name          string           `json:"name"`
	Kind          SymbolKind       `json:"kind"`
	Location      SymbolLocation   `json:"location"`
	Documentation string           `json:"documentation,omitempty"`
	Source        *SourceContext   `json:"source,omitempty"`
	References    []SymbolLocation `json:"references"`
}