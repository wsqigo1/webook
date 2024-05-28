package main

import "fmt"

func plusOne(digits []int) []int {
	for i := len(digits) - 1; i >= 0; i-- {
		num := (digits[i] + 1) % 10
		digits[i] = num
		if num != 0 {
			return digits
		}
	}
	return append([]int{1}, digits...)
}

func main() {
	fmt.Println(plusOne([]int{9}))
}
