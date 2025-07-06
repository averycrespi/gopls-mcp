package results

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSymbolKind(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected SymbolKind
	}{
		{
			name:     "File symbol kind",
			input:    1,
			expected: SymbolKindFile,
		},
		{
			name:     "Module symbol kind",
			input:    2,
			expected: SymbolKindModule,
		},
		{
			name:     "Function symbol kind",
			input:    12,
			expected: SymbolKindFunction,
		},
		{
			name:     "Struct symbol kind",
			input:    23,
			expected: SymbolKindStruct,
		},
		{
			name:     "Method symbol kind",
			input:    6,
			expected: SymbolKindMethod,
		},
		{
			name:     "Variable symbol kind",
			input:    13,
			expected: SymbolKindVariable,
		},
		{
			name:     "Unknown symbol kind - negative",
			input:    -1,
			expected: SymbolKindUnknown,
		},
		{
			name:     "Unknown symbol kind - zero",
			input:    0,
			expected: SymbolKindUnknown,
		},
		{
			name:     "Unknown symbol kind - out of range",
			input:    999,
			expected: SymbolKindUnknown,
		},
		{
			name:     "TypeParameter symbol kind",
			input:    26,
			expected: SymbolKindTypeParameter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSymbolKind(tt.input)
			assert.Equal(t, tt.expected, result, "NewSymbolKind(%d) should return %v", tt.input, tt.expected)
		})
	}
}

func TestSymbolKindString(t *testing.T) {
	tests := []struct {
		name     string
		kind     SymbolKind
		expected string
	}{
		{
			name:     "Function kind to string",
			kind:     SymbolKindFunction,
			expected: "function",
		},
		{
			name:     "Struct kind to string",
			kind:     SymbolKindStruct,
			expected: "struct",
		},
		{
			name:     "Method kind to string",
			kind:     SymbolKindMethod,
			expected: "method",
		},
		{
			name:     "Unknown kind to string",
			kind:     SymbolKindUnknown,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.kind)
			assert.Equal(t, tt.expected, result, "string(%v) should return %v", tt.kind, tt.expected)
		})
	}
}

func TestSymbolKindMapCompleteness(t *testing.T) {
	// Test that all expected LSP symbol kinds are mapped
	expectedMappings := map[int]SymbolKind{
		1:  SymbolKindFile,
		2:  SymbolKindModule,
		3:  SymbolKindNamespace,
		4:  SymbolKindPackage,
		5:  SymbolKindClass,
		6:  SymbolKindMethod,
		7:  SymbolKindProperty,
		8:  SymbolKindField,
		9:  SymbolKindConstructor,
		10: SymbolKindEnum,
		11: SymbolKindInterface,
		12: SymbolKindFunction,
		13: SymbolKindVariable,
		14: SymbolKindConstant,
		15: SymbolKindString,
		16: SymbolKindNumber,
		17: SymbolKindBoolean,
		18: SymbolKindArray,
		19: SymbolKindObject,
		20: SymbolKindKey,
		21: SymbolKindNull,
		22: SymbolKindEnumMember,
		23: SymbolKindStruct,
		24: SymbolKindEvent,
		25: SymbolKindOperator,
		26: SymbolKindTypeParameter,
	}

	for lspKind, expectedSymbolKind := range expectedMappings {
		result := NewSymbolKind(lspKind)
		assert.Equal(t, expectedSymbolKind, result, "LSP kind %d should map to %v", lspKind, expectedSymbolKind)
	}
}
