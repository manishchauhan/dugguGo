package util

type Other struct {
	Houseno string
	Salary  []int
}
type User struct {
	Name   string
	Age    int
	_Other Other
}

func Addnumber(num1 int, num2 int) int {
	sum := num1 + num2
	return sum
}
func minus(num1 int, num2 int) int {
	result := 0
	if num1 > num2 {
		result = num1 - num2
	} else {
		result = num2 - num1
	}
	return result
}
func multiple(num1 int, num2 int) int {
	return num1 * num2
}

func MakeSumArray() int {
	var user User
	user.Name = "manish chauhan"
	user.Age = 22
	user._Other.Houseno = "house no 1557"
	user._Other.Salary = []int{1, 2, 3, 4, 5}
	user._Other.Salary = append(user._Other.Salary, 6, 7, 8)
	sum := 0
	for i := 0; i < len(user._Other.Salary); i++ {
		sum += user._Other.Salary[i]
	}
	return sum
}
