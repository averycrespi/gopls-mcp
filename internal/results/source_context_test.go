package results

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSourceContext(t *testing.T) {
	tests := []struct {
		name          string
		contextStr    string
		highlightLine int
		expected      *SourceContext
	}{
		{
			name: "Simple context with highlight",
			contextStr: `    10: func main() {
>>> 11: 	fmt.Println("Hello")
    12: }`,
			highlightLine: 11,
			expected: &SourceContext{
				Lines: []SourceLine{
					{Number: 10, Content: "func main() {", Highlight: false},
					{Number: 11, Content: "\tfmt.Println(\"Hello\")", Highlight: true},
					{Number: 12, Content: "}", Highlight: false},
				},
			},
		},
		{
			name: "Context without highlight",
			contextStr: `    5: package main
    6: 
    7: import "fmt"`,
			highlightLine: 6,
			expected: &SourceContext{
				Lines: []SourceLine{
					{Number: 5, Content: "package main", Highlight: false},
					{Number: 6, Content: "", Highlight: false},
					{Number: 7, Content: "import \"fmt\"", Highlight: false},
				},
			},
		},
		{
			name: "Mixed content with multiple highlights",
			contextStr: `    8: type Calculator struct {
>>> 9: 	Value float64
    10: }
>>> 11: func NewCalculator() *Calculator {
    12: 	return &Calculator{}`,
			highlightLine: 9,
			expected: &SourceContext{
				Lines: []SourceLine{
					{Number: 8, Content: "type Calculator struct {", Highlight: false},
					{Number: 9, Content: "\tValue float64", Highlight: true},
					{Number: 10, Content: "}", Highlight: false},
					{Number: 11, Content: "func NewCalculator() *Calculator {", Highlight: true},
					{Number: 12, Content: "\treturn &Calculator{}", Highlight: false},
				},
			},
		},
		{
			name:          "Empty context",
			contextStr:    "",
			highlightLine: 1,
			expected: &SourceContext{
				Lines: []SourceLine{},
			},
		},
		{
			name: "Context with empty lines",
			contextStr: `    1: package main

    3: import "fmt"

>>> 5: func main() {`,
			highlightLine: 5,
			expected: &SourceContext{
				Lines: []SourceLine{
					{Number: 1, Content: "package main", Highlight: false},
					{Number: 3, Content: "import \"fmt\"", Highlight: false},
					{Number: 5, Content: "func main() {", Highlight: true},
				},
			},
		},
		{
			name: "Malformed line numbers",
			contextStr: `invalid line format
    5: valid line
>>> bad format: content`,
			highlightLine: 5,
			expected: &SourceContext{
				Lines: []SourceLine{
					{Number: 5, Content: "valid line", Highlight: false},
				},
			},
		},
		{
			name: "Lines with colons in content",
			contextStr: `    1: func test() error {
>>> 2: 	return fmt.Errorf("error: %s", msg)
    3: }`,
			highlightLine: 2,
			expected: &SourceContext{
				Lines: []SourceLine{
					{Number: 1, Content: "func test() error {", Highlight: false},
					{Number: 2, Content: "\treturn fmt.Errorf(\"error: %s\", msg)", Highlight: true},
					{Number: 3, Content: "}", Highlight: false},
				},
			},
		},
		{
			name: "Large line numbers",
			contextStr: `  999: large line number
>>> 1000: highlighted large line
 1001: another large line`,
			highlightLine: 1000,
			expected: &SourceContext{
				Lines: []SourceLine{
					{Number: 999, Content: "large line number", Highlight: false},
					{Number: 1000, Content: "highlighted large line", Highlight: true},
					{Number: 1001, Content: "another large line", Highlight: false},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSourceContext(tt.contextStr, tt.highlightLine)
			
			assert.Equal(t, tt.expected, result, "NewSourceContext should return expected result")
			
			// Additional detailed assertions for better debugging
			assert.Len(t, result.Lines, len(tt.expected.Lines), "Line count should match")
			
			for i, line := range result.Lines {
				if i < len(tt.expected.Lines) {
					expected := tt.expected.Lines[i]
					assert.Equal(t, expected.Number, line.Number, "Line %d: Number should match", i)
					assert.Equal(t, expected.Content, line.Content, "Line %d: Content should match", i)
					assert.Equal(t, expected.Highlight, line.Highlight, "Line %d: Highlight should match", i)
				}
			}
		})
	}
}

func TestSourceLineJSON(t *testing.T) {
	tests := []struct {
		name     string
		line     SourceLine
		wantJSON string
	}{
		{
			name: "Regular line",
			line: SourceLine{
				Number:    10,
				Content:   "func main() {",
				Highlight: false,
			},
			wantJSON: `{"number":10,"content":"func main() {","highlight":false}`,
		},
		{
			name: "Highlighted line",
			line: SourceLine{
				Number:    11,
				Content:   "\tfmt.Println(\"Hello\")",
				Highlight: true,
			},
			wantJSON: `{"number":11,"content":"\tfmt.Println(\"Hello\")","highlight":true}`,
		},
		{
			name: "Empty content line",
			line: SourceLine{
				Number:    5,
				Content:   "",
				Highlight: false,
			},
			wantJSON: `{"number":5,"content":"","highlight":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.line)
			assert.NoError(t, err, "JSON marshaling should not error")
			
			got := string(jsonBytes)
			assert.Equal(t, tt.wantJSON, got, "JSON output should match expected")
		})
	}
}

func TestSourceContextEdgeCases(t *testing.T) {
	t.Run("Whitespace only context", func(t *testing.T) {
		result := NewSourceContext("   \n\t\n   ", 1)
		expected := &SourceContext{Lines: []SourceLine{}}
		
		assert.Equal(t, expected, result, "Should return empty lines for whitespace-only context")
	})
	
	t.Run("Context with tabs and spaces", func(t *testing.T) {
		contextStr := "\t>>> 5: \tfunction with tabs\n    6: \tnormal line"
		result := NewSourceContext(contextStr, 5)
		
		assert.Len(t, result.Lines, 2, "Should parse 2 lines")
		assert.Equal(t, "\tfunction with tabs", result.Lines[0].Content, "Tab should be preserved in content")
		assert.True(t, result.Lines[0].Highlight, "First line should be highlighted")
		assert.False(t, result.Lines[1].Highlight, "Second line should not be highlighted")
	})
}