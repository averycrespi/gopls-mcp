package results

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	anchorScheme = "anchor"
)

// SymbolAnchor represents the encoding of the fixed position of a symbol in a file
type SymbolAnchor string

// NewSymbolAnchor creates a new SymbolAnchor from a file, line number (1-indexed), and character (1-indexed)
func NewSymbolAnchor(file string, line int, character int) SymbolAnchor {
	return SymbolAnchor(fmt.Sprintf("%s://%s#%d:%d", anchorScheme, file, line, character))
}

// IsValid checks if the anchor has a valid format
func (a SymbolAnchor) IsValid() bool {
	_, _, _, err := a.Parse()
	return err == nil
}

// String returns the string representation of the anchor
func (a SymbolAnchor) String() string {
	return string(a)
}

// ToSymbolLocation converts the anchor to a SymbolLocation (already 1-indexed)
func (a SymbolAnchor) ToSymbolLocation() (SymbolLocation, error) {
	file, line, character, err := a.Parse()
	if err != nil {
		return SymbolLocation{}, err
	}
	return SymbolLocation{
		File:      file,
		Line:      line,      // Already 1-indexed
		Character: character, // Already 1-indexed
	}, nil
}

// ToLSPPosition converts the anchor to LSP position (0-indexed for internal use)
func (a SymbolAnchor) ToLSPPosition() (file string, line int, character int, err error) {
	file, line1, character1, err := a.Parse()
	if err != nil {
		return "", 0, 0, err
	}
	return file, line1 - 1, character1 - 1, nil // Convert to 0-indexed for LSP
}

// Parse parses a SymbolAnchor into a file, line number (1-indexed), and character (1-indexed)
func (a SymbolAnchor) Parse() (file string, line int, character int, err error) {
	anchorStr := string(a)

	// Check scheme
	if !strings.HasPrefix(anchorStr, anchorScheme+"://") {
		return "", 0, 0, fmt.Errorf("invalid anchor scheme, expected '%s://', got: %s", anchorScheme, anchorStr)
	}

	// Remove scheme
	rest := anchorStr[len(anchorScheme)+3:] // +3 for "://"

	// Split on # to separate file from coordinates
	parts := strings.SplitN(rest, "#", 2)
	if len(parts) != 2 {
		return "", 0, 0, fmt.Errorf("invalid anchor format, expected 'anchor://FILE#LINE:CHAR', got: %s", anchorStr)
	}

	file = parts[0]
	if file == "" {
		return "", 0, 0, fmt.Errorf("empty file in anchor: %s", anchorStr)
	}

	// Parse coordinates (LINE:CHAR)
	coords := parts[1]
	coordParts := strings.SplitN(coords, ":", 2)
	if len(coordParts) != 2 {
		return "", 0, 0, fmt.Errorf("invalid coordinate format, expected 'LINE:CHAR', got: %s", coords)
	}

	line, err = strconv.Atoi(coordParts[0])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid line number '%s': %v", coordParts[0], err)
	}

	character, err = strconv.Atoi(coordParts[1])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid character number '%s': %v", coordParts[1], err)
	}

	if line < 1 {
		return "", 0, 0, fmt.Errorf("line number must be positive (1-indexed): %d", line)
	}

	if character < 1 {
		return "", 0, 0, fmt.Errorf("character number must be positive (1-indexed): %d", character)
	}

	return file, line, character, nil
}
