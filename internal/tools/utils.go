package tools

import (
	"path/filepath"
	"strings"
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

// GetRelativePath converts an absolute path to a relative path from the workspace root
func GetRelativePath(absolutePath, workspaceRoot string) string {
	if rel, err := filepath.Rel(workspaceRoot, absolutePath); err == nil {
		return rel
	}
	return filepath.Base(absolutePath)
}

// IsValidGoIdentifier checks if a string is a valid Go identifier
func IsValidGoIdentifier(name string) bool {
	if name == "" {
		return false
	}

	// Check if it's a Go keyword
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	if keywords[name] {
		return false
	}

	// Must start with letter or underscore
	if !isLetter(rune(name[0])) && name[0] != '_' {
		return false
	}

	// Rest must be letters, digits, or underscores
	for _, r := range name[1:] {
		if !isLetter(r) && !isDigit(r) && r != '_' {
			return false
		}
	}

	return true
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
