package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

//题目 ：使用原子操作（ sync/atomic 包）实现一个无锁的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
//考察点 ：原子操作、并发数据安全。

func main() {
	counter := atomic.Int64{}
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				counter.Add(1)
			}
		}()
	}
	wg.Wait()
	fmt.Println("counter:", counter.Load())
}
