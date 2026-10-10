package main

import (
	"fmt"
	"sync"
	"time"
)

//题目 ：设计一个任务调度器，接收一组任务（可以用函数表示），并使用协程并发执行这些任务，同时统计每个任务的执行时间。
//考察点 ：协程原理、并发任务调度

func addto100() int {
	var sum int
	for i := 1; i <= 100; i++ {
		sum += i
	}
	return sum
}

func main() {
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := time.Now()
			sum := addto100()
			elapsed := time.Since(start)
			fmt.Printf("任务 %d 执行完成，结果=%d，耗时=%v\n", i+1, sum, elapsed)
		}(i)
	}
	wg.Wait()
}
