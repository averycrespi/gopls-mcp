package main

import "fmt"

// Example function for testing gopls-mcp functionality
func ExampleFunction(message string) string {
	return fmt.Sprintf("Hello, %s!", message)
}

// ExampleStruct demonstrates struct operations
type ExampleStruct struct {
	Name  string
	Value int
}

// Method on ExampleStruct
func (e *ExampleStruct) GetInfo() string {
	return fmt.Sprintf("Name: %s, Value: %d", e.Name, e.Value)
}

func main() {
	result := ExampleFunction("World")
	fmt.Println(result)
	
	example := &ExampleStruct{
		Name:  "Test",
		Value: 42,
	}
	
	fmt.Println(example.GetInfo())
}