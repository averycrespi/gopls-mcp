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
	Hover(ctx context.Context, uri string, position Position) (string, error)
	GetCompletion(ctx context.Context, uri string, position Position) ([]CompletionItem, error)
	FuzzyFindSymbol(ctx context.Context, query string) ([]SymbolInformation, error)
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

// Diagnostic represents a diagnostic message
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
	Source   string `json:"source,omitempty"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label  string `json:"label"`
	Kind   int    `json:"kind"`
	Detail string `json:"detail,omitempty"`
}
