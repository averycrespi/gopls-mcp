package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSymbolAnchor(t *testing.T) {
	anchor := NewSymbolAnchor("test.go", 10, 5)
	expected := "anchor://test.go#10:5"
	assert.Equal(t, expected, anchor.String())
}

func TestSymbolAnchor_Parse(t *testing.T) {
	tests := []struct {
		name          string
		anchor        SymbolAnchor
		expectedFile  string
		expectedLine  int
		expectedChar  int
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid anchor",
			anchor:       "anchor://test.go#10:5",
			expectedFile: "test.go",
			expectedLine: 10,
			expectedChar: 5,
			expectError:  false,
		},
		{
			name:         "valid anchor with path",
			anchor:       "anchor://src/main.go#1:1",
			expectedFile: "src/main.go",
			expectedLine: 1,
			expectedChar: 1,
			expectError:  false,
		},
		{
			name:          "invalid scheme",
			anchor:        "http://test.go#10:5",
			expectError:   true,
			errorContains: "invalid anchor scheme",
		},
		{
			name:          "missing scheme",
			anchor:        "test.go#10:5",
			expectError:   true,
			errorContains: "invalid anchor scheme",
		},
		{
			name:          "no fragment separator",
			anchor:        "anchor://test.go",
			expectError:   true,
			errorContains: "invalid anchor format",
		},
		{
			name:          "empty file",
			anchor:        "anchor://#10:5",
			expectError:   true,
			errorContains: "empty file",
		},
		{
			name:          "invalid coordinates format",
			anchor:        "anchor://test.go#10",
			expectError:   true,
			errorContains: "invalid coordinate format",
		},
		{
			name:          "invalid line number",
			anchor:        "anchor://test.go#abc:5",
			expectError:   true,
			errorContains: "invalid line number",
		},
		{
			name:          "invalid character number",
			anchor:        "anchor://test.go#10:abc",
			expectError:   true,
			errorContains: "invalid character number",
		},
		{
			name:          "zero line number",
			anchor:        "anchor://test.go#0:5",
			expectError:   true,
			errorContains: "line number must be positive (1-indexed)",
		},
		{
			name:          "zero character number",
			anchor:        "anchor://test.go#10:0",
			expectError:   true,
			errorContains: "character number must be positive (1-indexed)",
		},
		{
			name:          "negative line number",
			anchor:        "anchor://test.go#-1:5",
			expectError:   true,
			errorContains: "line number must be positive (1-indexed)",
		},
		{
			name:          "negative character number",
			anchor:        "anchor://test.go#10:-1",
			expectError:   true,
			errorContains: "character number must be positive (1-indexed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, line, char, err := tt.anchor.Parse()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFile, file)
				assert.Equal(t, tt.expectedLine, line)
				assert.Equal(t, tt.expectedChar, char)
			}
		})
	}
}

func TestSymbolAnchor_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		anchor   SymbolAnchor
		expected bool
	}{
		{
			name:     "valid anchor",
			anchor:   "anchor://test.go#10:5",
			expected: true,
		},
		{
			name:     "invalid anchor",
			anchor:   "invalid",
			expected: false,
		},
		{
			name:     "empty anchor",
			anchor:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.anchor.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSymbolAnchor_ToSymbolLocation(t *testing.T) {
	anchor := SymbolAnchor("anchor://test.go#10:5")
	location, err := anchor.ToSymbolLocation()

	assert.NoError(t, err)
	assert.Equal(t, "test.go", location.File)
	assert.Equal(t, 10, location.Line)     // Should remain 1-indexed
	assert.Equal(t, 5, location.Character) // Should remain 1-indexed
}

func TestSymbolAnchor_ToSymbolLocation_Invalid(t *testing.T) {
	anchor := SymbolAnchor("invalid")
	_, err := anchor.ToSymbolLocation()
	assert.Error(t, err)
}

func TestSymbolLocation_ToAnchor(t *testing.T) {
	location := SymbolLocation{
		File:      "test.go",
		Line:      10, // 1-indexed
		Character: 5,  // 1-indexed
	}

	anchor := location.ToAnchor()
	expected := "anchor://test.go#10:5" // Should remain 1-indexed
	assert.Equal(t, expected, anchor.String())
}

func TestRoundTrip(t *testing.T) {
	// Test that converting location -> anchor -> location preserves data
	originalLocation := SymbolLocation{
		File:      "src/main.go",
		Line:      15,
		Character: 8,
	}

	anchor := originalLocation.ToAnchor()
	convertedLocation, err := anchor.ToSymbolLocation()

	assert.NoError(t, err)
	assert.Equal(t, originalLocation, convertedLocation)
}

func TestSymbolAnchor_ToLSPPosition(t *testing.T) {
	anchor := SymbolAnchor("anchor://test.go#10:5")
	file, line, character, err := anchor.ToLSPPosition()

	assert.NoError(t, err)
	assert.Equal(t, "test.go", file)
	assert.Equal(t, 9, line)      // Should be 0-indexed for LSP
	assert.Equal(t, 4, character) // Should be 0-indexed for LSP
}
