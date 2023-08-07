/*
this class contains methods which is equivalent to javascript functions which we called
higher order functions like map,fliter,reduce check link https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/filter
*/
package hoc

import (
	"golang.org/x/exp/constraints"
)

// this can be expensive as we need to create a new array
func GoMap[T constraints.Ordered](values []T, Func func(value T) T) []T {
	var result []T
	for _, value := range values {
		result = append(result, Func(value))
	}
	return result
}

// this can be expensive as we need to create a new array
func GoFilter[T constraints.Ordered](values []T, Func func(value T) bool) []T {
	var result []T
	for _, value := range values {
		if Func(value) {
			result = append(result, value)
		}
	}
	return result
}
func goReduce() {
}
