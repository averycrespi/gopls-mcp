package results

// RenameSymbolByAnchorToolResult represents the result of renaming a symbol
type RenameSymbolByAnchorToolResult struct {
	Message   string                       `json:"message"`
	Arguments RenameSymbolByAnchorToolArgs `json:"arguments"`
	FileEdits []FileEdit                   `json:"file_edits,omitempty"`
}

// RenameSymbolByAnchorToolArgs represents the input arguments for the rename symbol tool
type RenameSymbolByAnchorToolArgs struct {
	SymbolAnchor string `json:"symbol_anchor"`
	NewName      string `json:"new_name"`
}

// FileEdit represents name changes in a single file
type FileEdit struct {
	File    string       `json:"file"`
	Changes []NameChange `json:"changes"`
}

// NameChange represents a name change from old to new
type NameChange struct {
	OldName string `json:"old_name"` // The original name being replaced
	NewName string `json:"new_name"` // The new name replacing the old one
}
