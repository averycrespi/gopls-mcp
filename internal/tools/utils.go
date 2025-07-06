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
	ToolSymbolSearch     = "symbol_search"
)

// LSP Symbol Kinds - based on Language Server Protocol specification
// See: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#symbolKind
const (
	SymbolKindFile          = 1
	SymbolKindModule        = 2
	SymbolKindNamespace     = 3
	SymbolKindPackage       = 4
	SymbolKindClass         = 5
	SymbolKindMethod        = 6
	SymbolKindProperty      = 7
	SymbolKindField         = 8
	SymbolKindConstructor   = 9
	SymbolKindEnum          = 10
	SymbolKindInterface     = 11
	SymbolKindFunction      = 12
	SymbolKindVariable      = 13
	SymbolKindConstant      = 14
	SymbolKindString        = 15
	SymbolKindNumber        = 16
	SymbolKindBoolean       = 17
	SymbolKindArray         = 18
	SymbolKindObject        = 19
	SymbolKindKey           = 20
	SymbolKindNull          = 21
	SymbolKindEnumMember    = 22
	SymbolKindStruct        = 23
	SymbolKindEvent         = 24
	SymbolKindOperator      = 25
	SymbolKindTypeParameter = 26
)

// symbolKindToString converts LSP symbol kind number to readable string
func symbolKindToString(kind int) string {
	switch kind {
	case SymbolKindFile:
		return "file"
	case SymbolKindModule:
		return "module"
	case SymbolKindNamespace:
		return "namespace"
	case SymbolKindPackage:
		return "package"
	case SymbolKindClass:
		return "class"
	case SymbolKindMethod:
		return "method"
	case SymbolKindProperty:
		return "property"
	case SymbolKindField:
		return "field"
	case SymbolKindConstructor:
		return "constructor"
	case SymbolKindEnum:
		return "enum"
	case SymbolKindInterface:
		return "interface"
	case SymbolKindFunction:
		return "function"
	case SymbolKindVariable:
		return "variable"
	case SymbolKindConstant:
		return "constant"
	case SymbolKindString:
		return "string"
	case SymbolKindNumber:
		return "number"
	case SymbolKindBoolean:
		return "boolean"
	case SymbolKindArray:
		return "array"
	case SymbolKindObject:
		return "object"
	case SymbolKindKey:
		return "key"
	case SymbolKindNull:
		return "null"
	case SymbolKindEnumMember:
		return "enum_member"
	case SymbolKindStruct:
		return "struct"
	case SymbolKindEvent:
		return "event"
	case SymbolKindOperator:
		return "operator"
	case SymbolKindTypeParameter:
		return "type_parameter"
	default:
		return fmt.Sprintf("unknown(%d)", kind)
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
