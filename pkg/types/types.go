package types

import (
	"context"
	"encoding/json"
)

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

// Config represents the configuration for the gopls-mcp server
type Config struct {
	GoplsPath     string `json:"gopls_path,omitempty"`
	WorkspaceRoot string `json:"workspace_root"`
	LogLevel      string `json:"log_level,omitempty"`
}

// Client defines the interface for LSP client operations
type Client interface {
	Initialize(ctx context.Context, rootURI string) error
	GoToDefinition(ctx context.Context, uri string, position Position) ([]Location, error)
	FindReferences(ctx context.Context, uri string, position Position) ([]Location, error)
	Hover(ctx context.Context, uri string, position Position) (string, error)
	GetDiagnostics(ctx context.Context, uri string) ([]Diagnostic, error)
	GetCompletion(ctx context.Context, uri string, position Position) ([]CompletionItem, error)
	FormatDocument(ctx context.Context, uri string) ([]json.RawMessage, error)
	RenameSymbol(ctx context.Context, uri string, position Position, newName string) (map[string][]json.RawMessage, error)
	Shutdown(ctx context.Context) error
}

// Transport defines the interface for the transport layer
type Transport interface {
	Listen()
	IsClosed() bool
	Close()
	SendRequest(method string, params any) (json.RawMessage, error)
	SendNotification(method string, params any) error
}
