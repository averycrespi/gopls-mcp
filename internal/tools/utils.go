package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/averycrespi/gopls-mcp/pkg/types"
	"github.com/mark3labs/mcp-go/mcp"
)

// Tool names
const (
	ToolSymbolDefinition = "symbol_definition"
	ToolFindReferences   = "find_references"
	ToolHoverInfo        = "hover_info"
	ToolGetCompletion    = "get_completion"
)

// LSP symbol kinds, based on protocol specification
// See: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#symbolKind
const (
	LSPSymbolKindFile          = 1
	LSPSymbolKindModule        = 2
	LSPSymbolKindNamespace     = 3
	LSPSymbolKindPackage       = 4
	LSPSymbolKindClass         = 5
	LSPSymbolKindMethod        = 6
	LSPSymbolKindProperty      = 7
	LSPSymbolKindField         = 8
	LSPSymbolKindConstructor   = 9
	LSPSymbolKindEnum          = 10
	LSPSymbolKindInterface     = 11
	LSPSymbolKindFunction      = 12
	LSPSymbolKindVariable      = 13
	LSPSymbolKindConstant      = 14
	LSPSymbolKindString        = 15
	LSPSymbolKindNumber        = 16
	LSPSymbolKindBoolean       = 17
	LSPSymbolKindArray         = 18
	LSPSymbolKindObject        = 19
	LSPSymbolKindKey           = 20
	LSPSymbolKindNull          = 21
	LSPSymbolKindEnumMember    = 22
	LSPSymbolKindStruct        = 23
	LSPSymbolKindEvent         = 24
	LSPSymbolKindOperator      = 25
	LSPSymbolKindTypeParameter = 26
)

// symbolKindToEnum converts LSP symbol kind number to SymbolKind enum
func symbolKindToEnum(kind int) SymbolKind {
	switch kind {
	case LSPSymbolKindFile:
		return SymbolKindFile
	case LSPSymbolKindModule:
		return SymbolKindModule
	case LSPSymbolKindNamespace:
		return SymbolKindNamespace
	case LSPSymbolKindPackage:
		return SymbolKindPackage
	case LSPSymbolKindClass:
		return SymbolKindClass
	case LSPSymbolKindMethod:
		return SymbolKindMethod
	case LSPSymbolKindProperty:
		return SymbolKindProperty
	case LSPSymbolKindField:
		return SymbolKindField
	case LSPSymbolKindConstructor:
		return SymbolKindConstructor
	case LSPSymbolKindEnum:
		return SymbolKindEnum
	case LSPSymbolKindInterface:
		return SymbolKindInterface
	case LSPSymbolKindFunction:
		return SymbolKindFunction
	case LSPSymbolKindVariable:
		return SymbolKindVariable
	case LSPSymbolKindConstant:
		return SymbolKindConstant
	case LSPSymbolKindString:
		return SymbolKindString
	case LSPSymbolKindNumber:
		return SymbolKindNumber
	case LSPSymbolKindBoolean:
		return SymbolKindBoolean
	case LSPSymbolKindArray:
		return SymbolKindArray
	case LSPSymbolKindObject:
		return SymbolKindObject
	case LSPSymbolKindKey:
		return SymbolKindKey
	case LSPSymbolKindNull:
		return SymbolKindNull
	case LSPSymbolKindEnumMember:
		return SymbolKindEnumMember
	case LSPSymbolKindStruct:
		return SymbolKindStruct
	case LSPSymbolKindEvent:
		return SymbolKindEvent
	case LSPSymbolKindOperator:
		return SymbolKindOperator
	case LSPSymbolKindTypeParameter:
		return SymbolKindTypeParameter
	default:
		return SymbolKindUnknown
	}
}

// getFileURI converts a file path to a file URI
func getFileURI(filePath string, workspaceRoot string) string {
	if strings.HasPrefix(filePath, "file://") {
		return filePath
	}

	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workspaceRoot, filePath)
	}

	return "file://" + filePath
}

// getPosition extracts position from MCP request
func getPosition(req mcp.CallToolRequest) (types.Position, error) {
	line := mcp.ParseFloat64(req, "line", 0)
	character := mcp.ParseFloat64(req, "character", 0)

	return types.Position{
		Line:      int(line),
		Character: int(character),
	}, nil
}

// uriToPath converts a file URI to a local file path
func uriToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		return strings.TrimPrefix(uri, "file://")
	}
	return uri
}

// getRelativePath converts absolute path to relative path from workspace root
func getRelativePath(absolutePath, workspaceRoot string) string {
	if rel, err := filepath.Rel(workspaceRoot, absolutePath); err == nil {
		return rel
	}
	return filepath.Base(absolutePath)
}

// readSourceLines reads specific lines from a source file
func readSourceLines(filePath string, startLine, endLine int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	currentLine := 0

	for scanner.Scan() {
		if currentLine >= startLine && currentLine <= endLine {
			lines = append(lines, scanner.Text())
		}
		currentLine++
		if currentLine > endLine {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// getSymbolContext reads source code around a symbol location
func getSymbolContext(uri string, startLine, startChar int, contextLines int) (string, error) {
	filePath := uriToPath(uri)

	// Read lines around the symbol (with context)
	contextStart := startLine - contextLines
	if contextStart < 0 {
		contextStart = 0
	}
	contextEnd := startLine + contextLines

	lines, err := readSourceLines(filePath, contextStart, contextEnd)
	if err != nil {
		return "", err
	}

	// Format with line numbers and highlight the target line
	var result strings.Builder
	for i, line := range lines {
		lineNum := contextStart + i + 1 // 1-based line numbers
		if contextStart+i == startLine {
			result.WriteString(fmt.Sprintf(">>> %d: %s\n", lineNum, line))
		} else {
			result.WriteString(fmt.Sprintf("    %d: %s\n", lineNum, line))
		}
	}

	return result.String(), nil
}
