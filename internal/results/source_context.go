package results

// SourceContext represents source code context around a symbol
type SourceContext struct {
	Lines []SourceLine `json:"lines"`
}

// SourceLine represents a line of source code
type SourceLine struct {
	Number    int    `json:"number"`
	Content   string `json:"content"`
	Highlight bool   `json:"highlight"`
}
