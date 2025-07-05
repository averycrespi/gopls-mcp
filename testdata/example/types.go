package main

import "fmt"

// Operation represents a mathematical operation
type Operation int

const (
	// Addition represents the addition operation
	Addition Operation = iota
	// Subtraction represents the subtraction operation
	Subtraction
	// Multiplication represents the multiplication operation
	Multiplication
	// Division represents the division operation
	Division
)

// String returns the string representation of an operation
func (op Operation) String() string {
	switch op {
	case Addition:
		return "+"
	case Subtraction:
		return "-"
	case Multiplication:
		return "*"
	case Division:
		return "/"
	default:
		return "unknown"
	}
}

// Processor defines an interface for processing numbers
type Processor interface {
	Process(x, y float64) (float64, error)
}

// BasicProcessor implements the Processor interface
type BasicProcessor struct {
	operation Operation
}

// NewBasicProcessor creates a new BasicProcessor with the given operation
func NewBasicProcessor(op Operation) *BasicProcessor {
	return &BasicProcessor{operation: op}
}

// Process processes two numbers using the configured operation
func (bp *BasicProcessor) Process(x, y float64) (float64, error) {
	switch bp.operation {
	case Addition:
		return x + y, nil
	case Subtraction:
		return x - y, nil
	case Multiplication:
		return x * y, nil
	case Division:
		if y == 0 {
			return 0, ErrDivisionByZero
		}
		return x / y, nil
	default:
		return 0, ErrUnknownOperation
	}
}

// Custom error types
var (
	ErrDivisionByZero   = fmt.Errorf("division by zero")
	ErrUnknownOperation = fmt.Errorf("unknown operation")
)
