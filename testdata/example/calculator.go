package main

import "fmt"

// Calculator performs basic arithmetic operations
type Calculator struct {
	Value float64
}

// NewCalculator creates a new Calculator with the given initial value
func NewCalculator(value float64) *Calculator {
	return &Calculator{Value: value}
}

// Add adds the given number to the calculator's value
func (c *Calculator) Add(x float64) float64 {
	c.Value += x
	return c.Value
}

// Subtract subtracts the given number from the calculator's value
func (c *Calculator) Subtract(x float64) float64 {
	c.Value -= x
	return c.Value
}

// Multiply multiplies the calculator's value by the given number
func (c *Calculator) Multiply(x float64) float64 {
	c.Value *= x
	return c.Value
}

// Divide divides the calculator's value by the given number
func (c *Calculator) Divide(x float64) (float64, error) {
	if x == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	c.Value /= x
	return c.Value, nil
}

// Reset resets the calculator's value to zero
func (c *Calculator) Reset() {
	c.Value = 0
}

// String returns a string representation of the calculator
func (c *Calculator) String() string {
	return fmt.Sprintf("Calculator{Value: %.2f}", c.Value)
}
