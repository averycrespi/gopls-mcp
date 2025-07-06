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

