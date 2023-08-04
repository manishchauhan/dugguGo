package main

import "fmt"
func addNumber(num1 int, num2 int) int {
	sum := num1 + num2
	return sum
}

func main() {
	fmt.Println("Hello, World!", addNumber(10, 22))
}
