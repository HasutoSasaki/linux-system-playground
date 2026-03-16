package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func ioTask(id int) {
	fmt.Printf("task %d: start\n", id)
	time.Sleep(10 * time.Second) // I/O-bound的なタスク（ただ待つ）
	fmt.Printf("task %d: done\n", id)
}

func main() {
	runtime.GOMAXPROCS(1) // CPU1個に制限
	
	var wg sync.WaitGroup
	start := time.Now()

	wg.Add(2)
	go func() { defer wg.Done(); ioTask(1) }()
	go func() { defer wg.Done(); ioTask(2) }()
	
	wg.Wait()
	fmt.Printf("elapsed: %v\n", time.Since(start))
}