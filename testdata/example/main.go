package main

import (
	"fmt"
	"log"
)

func main() {
	// Create a new calculator
	calc := NewCalculator(10.0)
	fmt.Printf("Initial calculator: %s\n", calc.String())

	// Test basic operations
	result := calc.Add(5.0)
	fmt.Printf("After adding 5: %.2f\n", result)

	result = calc.Multiply(2.0)
	fmt.Printf("After multiplying by 2: %.2f\n", result)

	quotient, err := calc.Divide(3.0)
	if err != nil {
		log.Printf("Error dividing: %v", err)
	} else {
		fmt.Printf("After dividing by 3: %.2f\n", quotient)
	}

	// Test processor interface
	processor := NewBasicProcessor(Addition)
	sum, err := processor.Process(10.0, 20.0)
	if err != nil {
		log.Printf("Error processing: %v", err)
	} else {
		fmt.Printf("Processor result: %.2f\n", sum)
	}

	// Test utility functions
	utils := &MathUtils{}
	absValue := utils.Abs(-15.5)
	fmt.Printf("Absolute value of -15.5: %.2f\n", absValue)

	maxValue := utils.Max(10.5, 7.3)
	fmt.Printf("Max of 10.5 and 7.3: %.2f\n", maxValue)

	// Test global utility functions
	factResult := Factorial(5)
	fmt.Printf("Factorial of 5: %d\n", factResult)

	fibResult := Fibonacci(8)
	fmt.Printf("8th Fibonacci number: %d\n", fibResult)

	fmt.Printf("Is 42 even? %t\n", IsEven(42))
	fmt.Printf("Is 43 odd? %t\n", IsOdd(43))
}
