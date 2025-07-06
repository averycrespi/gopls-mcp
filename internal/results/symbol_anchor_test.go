package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSymbolAnchor(t *testing.T) {
	anchor := NewSymbolAnchor("test.go", 10, 5)
	expected := "go://test.go#10:5"
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
			anchor:       "go://test.go#10:5",
			expectedFile: "test.go",
			expectedLine: 10,
			expectedChar: 5,
			expectError:  false,
		},
		{
			name:         "valid anchor with path",
			anchor:       "go://src/main.go#1:1",
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
			anchor:        "go://test.go",
			expectError:   true,
			errorContains: "invalid anchor format",
		},
		{
			name:          "empty file",
			anchor:        "go://#10:5",
			expectError:   true,
			errorContains: "empty file",
		},
		{
			name:          "invalid coordinates format",
			anchor:        "go://test.go#10",
			expectError:   true,
			errorContains: "invalid coordinate format",
		},
		{
			name:          "invalid line number",
			anchor:        "go://test.go#abc:5",
			expectError:   true,
			errorContains: "invalid line number",
		},
		{
			name:          "invalid character number",
			anchor:        "go://test.go#10:abc",
			expectError:   true,
			errorContains: "invalid character number",
		},
		{
			name:          "zero line number",
			anchor:        "go://test.go#0:5",
			expectError:   true,
			errorContains: "display line must be positive (starts at 1)",
		},
		{
			name:          "zero character number",
			anchor:        "go://test.go#10:0",
			expectError:   true,
			errorContains: "display character must be positive (starts at 1)",
		},
		{
			name:          "negative line number",
			anchor:        "go://test.go#-1:5",
			expectError:   true,
			errorContains: "display line must be positive (starts at 1)",
		},
		{
			name:          "negative character number",
			anchor:        "go://test.go#10:-1",
			expectError:   true,
			errorContains: "display character must be positive (starts at 1)",
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
			anchor:   "go://test.go#10:5",
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
	anchor := SymbolAnchor("go://test.go#10:5")
	location, err := anchor.ToSymbolLocation()

	assert.NoError(t, err)
	assert.Equal(t, "test.go", location.File)
	assert.Equal(t, 10, location.DisplayLine) // Display line
	assert.Equal(t, 5, location.DisplayChar)  // Display character
}

func TestSymbolAnchor_ToSymbolLocation_Invalid(t *testing.T) {
	anchor := SymbolAnchor("invalid")
	_, err := anchor.ToSymbolLocation()
	assert.Error(t, err)
}

func TestSymbolLocation_ToAnchor(t *testing.T) {
	location := SymbolLocation{
		File:        "test.go",
		DisplayLine: 10, // Display line
		DisplayChar: 5,  // Display character
	}

	anchor := location.ToAnchor()
	expected := "go://test.go#10:5" // Display coordinates
	assert.Equal(t, expected, anchor.String())
}

func TestRoundTrip(t *testing.T) {
	// Test that converting location -> anchor -> location preserves data
	originalLocation := SymbolLocation{
		File:        "src/main.go",
		DisplayLine: 15,
		DisplayChar: 8,
	}

	anchor := originalLocation.ToAnchor()
	convertedLocation, err := anchor.ToSymbolLocation()

	assert.NoError(t, err)
	assert.Equal(t, originalLocation, convertedLocation)
}

func TestSymbolAnchor_ToFilePosition(t *testing.T) {
	anchor := SymbolAnchor("go://test.go#10:5")
	file, position, err := anchor.ToFilePosition()

	assert.NoError(t, err)
	assert.Equal(t, "test.go", file)
	assert.Equal(t, 9, position.Line)      // LSP line (0-indexed)
	assert.Equal(t, 4, position.Character) // LSP character (0-indexed)
}
