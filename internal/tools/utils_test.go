package tools

import (
	"path/filepath"
	"testing"

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

