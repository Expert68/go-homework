package main

import "fmt"

// 实现一个函数，接收一个整数切片的指针，将切片中的每个元素乘以2。
func mul2(a *[]int) []int {
	for i, v := range *a {
		(*a)[i] = v * 2
	}
	return *a
}

func main() {
	a := []int{1, 2, 3}
	fmt.Println(mul2(&a))
}
