package main

import (
	"fmt"
	"sync"
)

//题目 ：实现一个带有缓冲的通道，生产者协程向通道中发送100个整数，消费者协程从通道中接收这些整数并打印。
//考察点 ：通道的缓冲机制。

func producer(ch chan<- int) {
	for i := 1; i <= 100; i++ {
		ch <- i
	}
	close(ch)
}

func consumer(ch <-chan int) {
	for i := range ch {
		fmt.Println(i)
	}
}

func main() {
	ch := make(chan int, 3)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() { defer wg.Done(); producer(ch) }()
	go func() { defer wg.Done(); consumer(ch) }()
	wg.Wait()
}
