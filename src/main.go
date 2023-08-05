package main

import (
	"fmt"

	"github.com/manishchauhan/dugguGo/util/hoc"
)

func main() {

	values := []string{"manish", "sachin", "manish", "sachin", "deepak"}

	// Example usage of goMap with a custom function.
	squaredValues := hoc.GoFilter[string](values, func(value string) bool {
		return value == "deepak"
	})
	fmt.Println("Squared values:", squaredValues)
}
