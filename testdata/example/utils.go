package main

import (
	"math"
)

// MathUtils provides utility mathematical functions
type MathUtils struct{}

// Abs returns the absolute value of x
func (mu *MathUtils) Abs(x float64) float64 {
	return math.Abs(x)
}

// Max returns the maximum of two numbers
func (mu *MathUtils) Max(x, y float64) float64 {
	return math.Max(x, y)
}

// Min returns the minimum of two numbers
func (mu *MathUtils) Min(x, y float64) float64 {
	return math.Min(x, y)
}

// Round rounds a number to the nearest integer
func (mu *MathUtils) Round(x float64) float64 {
	return math.Round(x)
}

// Global utility functions

// IsEven checks if a number is even
func IsEven(n int) bool {
	return n%2 == 0
}

// IsOdd checks if a number is odd
func IsOdd(n int) bool {
	return n%2 != 0
}

// Factorial calculates the factorial of n
func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Factorial(n-1)
}

// Fibonacci returns the nth Fibonacci number
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}
