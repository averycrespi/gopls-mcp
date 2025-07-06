package types

import (
	"context"
)

// Client defines the LSP client interface
type Client interface {
	Start(ctx context.Context, workspaceRoot string) error
	Stop(ctx context.Context) error

	GoToDefinition(ctx context.Context, uri string, position Position) ([]Location, error)
	FindReferences(ctx context.Context, uri string, position Position) ([]Location, error)
	GetHoverInfo(ctx context.Context, uri string, position Position) (string, error)
	FuzzyFindSymbol(ctx context.Context, query string) ([]SymbolInformation, error)
	GetDocumentSymbols(ctx context.Context, uri string) ([]DocumentSymbol, error)
}

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a text document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// SymbolInformation represents information about a symbol
type SymbolInformation struct {
	Name     string   `json:"name"`
	Kind     int      `json:"kind"`
	Location Location `json:"location"`
}

// DocumentSymbol represents a symbol within a document with hierarchical structure
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}
