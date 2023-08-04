package main

import (
	"fmt"

	"github.com/manishchauhan/dugguGo/util/tree"
)

// Function to add two numbers and return their sum
func addNumbers(a, b int) int {
	return a + b
}
func resolveTree() {
	t := tree.Tree{}

	t.Insert(50)
	t.Insert(30)
	t.Insert(20)
	t.Insert(40)
	t.Insert(70)
	t.Insert(60)
	t.Insert(80)

	t.InOrderTraversal()
}

func main() {
	num1 := 10
	num2 := 20

	// Call the addNumbers function with num1 and num2 as arguments
	result := addNumbers(num1, num2)
	resolveTree()
	fmt.Printf("The sum of %d and %d is %d\n", num1, num2, result)
}
