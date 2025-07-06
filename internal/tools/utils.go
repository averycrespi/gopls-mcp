package tools

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/averycrespi/gopls-mcp/pkg/types"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool names
const (
	ToolFindReferences = "find_references"
	ToolGetCompletion  = "get_completion"
)

// PathToUri converts a file path to a file URI
func PathToUri(filePath string, workspaceRoot string) string {
	if strings.HasPrefix(filePath, "file://") {
		return filePath
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workspaceRoot, filePath)
	}

	return "file://" + filePath
}

// UriToPath converts a file URI to a local file path
func UriToPath(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}

// GetPosition extracts position from MCP request
func GetPosition(req mcp.CallToolRequest) (types.Position, error) {
	line := mcp.ParseFloat64(req, "line", 0)
	character := mcp.ParseFloat64(req, "character", 0)

	return types.Position{
		Line:      int(line),
		Character: int(character),
	}, nil
}

// GetRelativePath converts an absolute path to a relative path from the workspace root
func GetRelativePath(absolutePath, workspaceRoot string) string {
	if rel, err := filepath.Rel(workspaceRoot, absolutePath); err == nil {
		return rel
	}
	return filepath.Base(absolutePath)
}

// ReadSourceLines reads specific source lines
func ReadSourceLines(reader io.Reader, startLine, endLine int, highlightLine int) ([]results.SourceLine, error) {
	var sourceLines []results.SourceLine
	scanner := bufio.NewScanner(reader)
	currentLine := 0

	for scanner.Scan() {
		if currentLine >= startLine && currentLine <= endLine {
			lineNum := currentLine + 1 // convert to 1-based line numbers
			isHighlight := currentLine == highlightLine

			sourceLines = append(sourceLines, results.SourceLine{
				Number:    lineNum,
				Content:   scanner.Text(),
				Highlight: isHighlight,
			})
		}
		currentLine++
		if currentLine > endLine {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan source lines: %w", err)
	}

	return sourceLines, nil
}

// ReadSourceContext reads source code around a location and returns structured SourceContext
func ReadSourceContext(reader io.Reader, startLine, contextLines int) (*results.SourceContext, error) {
	contextStart := max(0, startLine-contextLines)
	contextEnd := startLine + contextLines

	sourceLines, err := ReadSourceLines(reader, contextStart, contextEnd, startLine)
	if err != nil {
		return nil, fmt.Errorf("failed to read source lines: %w", err)
	}

	return &results.SourceContext{Lines: sourceLines}, nil
}
