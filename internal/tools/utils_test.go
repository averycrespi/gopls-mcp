package tools

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/averycrespi/gopls-mcp/pkg/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestPathToUri(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		workspaceRoot string
		expected      string
	}{
		{
			name:          "Absolute path",
			filePath:      "/home/user/project/main.go",
			workspaceRoot: "/home/user/project",
			expected:      "file:///home/user/project/main.go",
		},
		{
			name:          "Relative path",
			filePath:      "src/main.go",
			workspaceRoot: "/home/user/project",
			expected:      "file:///home/user/project/src/main.go",
		},
		{
			name:          "Already a URI",
			filePath:      "file:///home/user/project/main.go",
			workspaceRoot: "/home/user/project",
			expected:      "file:///home/user/project/main.go",
		},
		{
			name:          "Current directory relative",
			filePath:      "./main.go",
			workspaceRoot: "/home/user/project",
			expected:      "file:///home/user/project/main.go",
		},
		{
			name:          "Parent directory relative",
			filePath:      "../main.go",
			workspaceRoot: "/home/user/project",
			expected:      "file:///home/user/main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathToUri(tt.filePath, tt.workspaceRoot)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUriToPath(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "Standard file URI",
			uri:      "file:///home/user/project/main.go",
			expected: "/home/user/project/main.go",
		},
		{
			name:     "Windows file URI",
			uri:      "file:///C:/Users/user/project/main.go",
			expected: "/C:/Users/user/project/main.go",
		},
		{
			name:     "Already a path",
			uri:      "/home/user/project/main.go",
			expected: "/home/user/project/main.go",
		},
		{
			name:     "Empty URI",
			uri:      "",
			expected: "",
		},
		{
			name:     "Just file scheme",
			uri:      "file://",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UriToPath(tt.uri)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPosition(t *testing.T) {
	tests := []struct {
		name        string
		arguments   map[string]interface{}
		expected    types.Position
		expectError bool
	}{
		{
			name: "Valid position",
			arguments: map[string]interface{}{
				"line":      float64(10),
				"character": float64(5),
			},
			expected: types.Position{Line: 10, Character: 5},
		},
		{
			name: "Zero position",
			arguments: map[string]interface{}{
				"line":      float64(0),
				"character": float64(0),
			},
			expected: types.Position{Line: 0, Character: 0},
		},
		{
			name:      "Missing arguments defaults to zero",
			arguments: map[string]interface{}{},
			expected:  types.Position{Line: 0, Character: 0},
		},
		{
			name: "Partial arguments",
			arguments: map[string]interface{}{
				"line": float64(5),
			},
			expected: types.Position{Line: 5, Character: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request with the arguments
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.arguments
			
			result, err := GetPosition(request)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetRelativePath(t *testing.T) {
	tests := []struct {
		name          string
		absolutePath  string
		workspaceRoot string
		expected      string
	}{
		{
			name:          "File in workspace root",
			absolutePath:  "/home/user/project/main.go",
			workspaceRoot: "/home/user/project",
			expected:      "main.go",
		},
		{
			name:          "File in subdirectory",
			absolutePath:  "/home/user/project/src/utils/helper.go",
			workspaceRoot: "/home/user/project",
			expected:      "src/utils/helper.go",
		},
		{
			name:          "File outside workspace (returns basename)",
			absolutePath:  "/other/path/file.go", 
			workspaceRoot: "/home/user/project",
			expected:      "../../../other/path/file.go", // Returns relative path when possible
		},
		{
			name:          "Invalid path (returns basename)",
			absolutePath:  "invalid:path",
			workspaceRoot: "/home/user/project",
			expected:      "invalid:path", // Returns basename when relative path fails
		},
		{
			name:          "Same as workspace root",
			absolutePath:  "/home/user/project",
			workspaceRoot: "/home/user/project",
			expected:      ".",
		},
		{
			name:          "Empty paths",
			absolutePath:  "",
			workspaceRoot: "",
			expected:      ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRelativePath(tt.absolutePath, tt.workspaceRoot)
			
			// Normalize path separators for cross-platform compatibility
			expected := filepath.FromSlash(tt.expected)
			assert.Equal(t, expected, result)
		})
	}
}

func TestReadSourceLines(t *testing.T) {
	sourceCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	fmt.Println("Line 2")
	fmt.Println("Line 3")
}
`

	tests := []struct {
		name          string
		startLine     int
		endLine       int
		highlightLine int
		expected      []results.SourceLine
	}{
		{
			name:          "Read lines 0-2 with highlight on line 1",
			startLine:     0,
			endLine:       2,
			highlightLine: 1,
			expected: []results.SourceLine{
				{Number: 1, Content: "package main", Highlight: false},
				{Number: 2, Content: "", Highlight: true},
				{Number: 3, Content: "import \"fmt\"", Highlight: false},
			},
		},
		{
			name:          "Read single line with highlight",
			startLine:     4,
			endLine:       4,
			highlightLine: 4,
			expected: []results.SourceLine{
				{Number: 5, Content: "func main() {", Highlight: true},
			},
		},
		{
			name:          "Read lines without highlight",
			startLine:     5,
			endLine:       6,
			highlightLine: -1,
			expected: []results.SourceLine{
				{Number: 6, Content: "\tfmt.Println(\"Hello, World!\")", Highlight: false},
				{Number: 7, Content: "\tfmt.Println(\"Line 2\")", Highlight: false},
			},
		},
		{
			name:          "Read beyond end of content",
			startLine:     10,
			endLine:       15,
			highlightLine: 12,
			expected:      nil, // Returns nil slice when no lines found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(sourceCode)
			result, err := ReadSourceLines(reader, tt.startLine, tt.endLine, tt.highlightLine)
			
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadSourceLines_Error(t *testing.T) {
	// Test with a reader that will cause scanner error
	errorReader := &errorReader{}
	
	_, err := ReadSourceLines(errorReader, 0, 5, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan source lines")
}

func TestReadSourceContext(t *testing.T) {
	sourceCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	fmt.Println("Line 2")
	fmt.Println("Line 3")
}
`

	tests := []struct {
		name         string
		startLine    int
		contextLines int
		expected     *results.SourceContext
	}{
		{
			name:         "Context around line 5 with 1 line context",
			startLine:    5,
			contextLines: 1,
			expected: &results.SourceContext{
				Lines: []results.SourceLine{
					{Number: 5, Content: "func main() {", Highlight: false},
					{Number: 6, Content: "\tfmt.Println(\"Hello, World!\")", Highlight: true},
					{Number: 7, Content: "\tfmt.Println(\"Line 2\")", Highlight: false},
				},
			},
		},
		{
			name:         "Context around line 0 with 2 lines context",
			startLine:    0,
			contextLines: 2,
			expected: &results.SourceContext{
				Lines: []results.SourceLine{
					{Number: 1, Content: "package main", Highlight: true},
					{Number: 2, Content: "", Highlight: false},
					{Number: 3, Content: "import \"fmt\"", Highlight: false},
				},
			},
		},
		{
			name:         "Context at end of file",
			startLine:    8,
			contextLines: 1,
			expected: &results.SourceContext{
				Lines: []results.SourceLine{
					{Number: 8, Content: "\tfmt.Println(\"Line 3\")", Highlight: false},
					{Number: 9, Content: "}", Highlight: true},
				},
			},
		},
		{
			name:         "No context lines",
			startLine:    2,
			contextLines: 0,
			expected: &results.SourceContext{
				Lines: []results.SourceLine{
					{Number: 3, Content: "import \"fmt\"", Highlight: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(sourceCode)
			result, err := ReadSourceContext(reader, tt.startLine, tt.contextLines)
			
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadSourceContext_Error(t *testing.T) {
	// Test with a reader that will cause scanner error
	errorReader := &errorReader{}
	
	_, err := ReadSourceContext(errorReader, 5, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read source lines")
}

// errorReader is a helper type that always returns an error when reading
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}