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

// FileEdit represents edits to a single file
type FileEdit struct {
	File  string `json:"file"`
	Edits []Edit `json:"edits"`
}

// Edit represents a single text edit
type Edit struct {
	StartLine      int    `json:"start_line"`      // Display line (1-indexed)
	StartCharacter int    `json:"start_character"` // Display character (1-indexed)
	EndLine        int    `json:"end_line"`        // Display line (1-indexed)
	EndCharacter   int    `json:"end_character"`   // Display character (1-indexed)
	OldText        string `json:"old_text"`        // The text being replaced
	NewText        string `json:"new_text"`        // The replacement text
}
