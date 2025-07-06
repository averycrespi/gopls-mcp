package results

import (
	"fmt"
	"strings"
)

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

// NewSourceContext parses a source context string into structured SourceContext
func NewSourceContext(contextStr string, highlightLine int) *SourceContext {
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
			// Format: "    11: content" - parse more flexibly
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) >= 2 {
				// Extract just the number part by trimming whitespace
				numberPart := strings.TrimSpace(parts[0])
				fmt.Sscanf(numberPart, "%d", &lineNumber)
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
