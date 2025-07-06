package results

// SymbolKind represents the type of a symbol as an enum
type SymbolKind string

const (
	SymbolKindFile          SymbolKind = "file"
	SymbolKindModule        SymbolKind = "module"
	SymbolKindNamespace     SymbolKind = "namespace"
	SymbolKindPackage       SymbolKind = "package"
	SymbolKindClass         SymbolKind = "class"
	SymbolKindMethod        SymbolKind = "method"
	SymbolKindProperty      SymbolKind = "property"
	SymbolKindField         SymbolKind = "field"
	SymbolKindConstructor   SymbolKind = "constructor"
	SymbolKindEnum          SymbolKind = "enum"
	SymbolKindInterface     SymbolKind = "interface"
	SymbolKindFunction      SymbolKind = "function"
	SymbolKindVariable      SymbolKind = "variable"
	SymbolKindConstant      SymbolKind = "constant"
	SymbolKindString        SymbolKind = "string"
	SymbolKindNumber        SymbolKind = "number"
	SymbolKindBoolean       SymbolKind = "boolean"
	SymbolKindArray         SymbolKind = "array"
	SymbolKindObject        SymbolKind = "object"
	SymbolKindKey           SymbolKind = "key"
	SymbolKindNull          SymbolKind = "null"
	SymbolKindEnumMember    SymbolKind = "enum_member"
	SymbolKindStruct        SymbolKind = "struct"
	SymbolKindEvent         SymbolKind = "event"
	SymbolKindOperator      SymbolKind = "operator"
	SymbolKindTypeParameter SymbolKind = "type_parameter"
	SymbolKindUnknown       SymbolKind = "unknown"
)

// See: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#symbolKind
var symbolKindMap = map[int]SymbolKind{
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
	27: SymbolKindUnknown,
}

// NewSymbolKind returns the SymbolKind for a given LSP symbol kind
func NewSymbolKind(kind int) SymbolKind {
	symbolKind, ok := symbolKindMap[kind]
	if !ok {
		return SymbolKindUnknown
	}
	return symbolKind
}
