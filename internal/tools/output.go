package tools

import (
	"fmt"
	"strings"
)

// SymbolSearchResult represents the JSON structure for symbol search results
type SymbolSearchResult struct {
	Query   string                    `json:"query"`
	Count   int                       `json:"count"`
	Symbols []SymbolSearchResultEntry `json:"symbols"`
}

// SymbolSearchResultEntry represents a single symbol in the search results
type SymbolSearchResultEntry struct {
	Name          string         `json:"name"`
	Kind          string         `json:"kind"`
	Location      SymbolLocation `json:"location"`
	Documentation string         `json:"documentation,omitempty"`
	Source        *SourceContext `json:"source,omitempty"`
}

// SymbolLocation represents the location of a symbol
type SymbolLocation struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

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

// SymbolDefinitionResult represents the JSON structure for symbol definition results
type SymbolDefinitionResult struct {
	Query   string                        `json:"query"`
	Count   int                           `json:"count"`
	Symbols []SymbolDefinitionResultEntry `json:"symbols"`
}

// SymbolDefinitionResultEntry represents a single symbol definition in the results
type SymbolDefinitionResultEntry struct {
	Name        string                 `json:"name"`
	Kind        string                 `json:"kind"`
	Location    SymbolLocation         `json:"location"`
	Definitions []SymbolDefinitionInfo `json:"definitions"`
}

// SymbolDefinitionInfo represents information about a symbol definition
type SymbolDefinitionInfo struct {
	Location      SymbolLocation `json:"location"`
	Documentation string         `json:"documentation,omitempty"`
	Source        *SourceContext `json:"source,omitempty"`
}

// parseSourceContext parses a source context string into structured SourceContext
func parseSourceContext(contextStr string, highlightLine int) *SourceContext {
	lines := strings.Split(strings.TrimSpace(contextStr), "\n")
	sourceLines := make([]SourceLine, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse line format: ">>> 11: content" or "    11: content"
		isHighlight := strings.HasPrefix(line, ">>>")
		var lineNumber int
		var content string

		if isHighlight {
			// Format: ">>> 11: content"
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) >= 2 {
				fmt.Sscanf(parts[0], ">>> %d", &lineNumber)
				content = parts[1]
			}
		} else {
			// Format: "    11: content"
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) >= 2 {
				fmt.Sscanf(parts[0], "    %d", &lineNumber)
				content = parts[1]
			}
		}

		if lineNumber > 0 {
			sourceLines = append(sourceLines, SourceLine{
				Number:    lineNumber,
				Content:   content,
				Highlight: isHighlight,
			})
		}
	}

	return &SourceContext{Lines: sourceLines}
}
