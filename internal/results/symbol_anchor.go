package results

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/averycrespi/gopls-mcp/pkg/types"
)

const (
	anchorScheme = "go"
)

// SymbolAnchor represents the encoding of the fixed position of a symbol in a file
type SymbolAnchor string

// NewSymbolAnchor creates a new SymbolAnchor from a file, display line, and display character
func NewSymbolAnchor(file string, displayLine int, displayChar int) SymbolAnchor {
	return SymbolAnchor(fmt.Sprintf("%s://%s#%d:%d", anchorScheme, file, displayLine, displayChar))
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

// ToSymbolLocation converts the anchor to a SymbolLocation using display coordinates
func (a SymbolAnchor) ToSymbolLocation() (SymbolLocation, error) {
	file, displayLine, displayChar, err := a.Parse()
	if err != nil {
		return SymbolLocation{}, err
	}
	return SymbolLocation{
		File:        file,
		DisplayLine: displayLine, // Display line
		DisplayChar: displayChar, // Display character
	}, nil
}

// ToFilePosition converts the anchor to a file path and LSP position (0-indexed for internal use)
func (a SymbolAnchor) ToFilePosition() (file string, position types.Position, err error) {
	file, displayLine, displayChar, err := a.Parse()
	if err != nil {
		return "", types.Position{}, err
	}
	position = types.Position{
		Line:      displayLine - 1, // Convert display coordinates to LSP coordinates
		Character: displayChar - 1, // Convert display coordinates to LSP coordinates
	}
	return file, position, nil
}

// Parse parses a SymbolAnchor into a file, display line, and display character
func (a SymbolAnchor) Parse() (file string, displayLine int, displayChar int, err error) {
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
		return "", 0, 0, fmt.Errorf("invalid anchor format, expected 'go://FILE#LINE:CHAR', got: %s", anchorStr)
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

	displayLine, err = strconv.Atoi(coordParts[0])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid line number '%s': %v", coordParts[0], err)
	}

	displayChar, err = strconv.Atoi(coordParts[1])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid character number '%s': %v", coordParts[1], err)
	}

	if displayLine < 1 {
		return "", 0, 0, fmt.Errorf("display line must be positive (starts at 1): %d", displayLine)
	}

	if displayChar < 1 {
		return "", 0, 0, fmt.Errorf("display character must be positive (starts at 1): %d", displayChar)
	}

	return file, displayLine, displayChar, nil
}
