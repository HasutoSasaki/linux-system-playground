package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"os"
)

func fileIOTask(id int) {
	f, _ := os.CreateTemp("", "io-test")
	defer os.Remove(f.Name())
	defer f.Close()
	f.Write(make([]byte, 1024*1024*1024)) // 1GB書き込み
	f.Sync()
}

func main() {
	runtime.GOMAXPROCS(1) // CPU1個に制限
	
	var wg sync.WaitGroup
	start := time.Now()

	wg.Add(2)
	go func() { defer wg.Done(); fileIOTask(1) }()
	go func() { defer wg.Done(); fileIOTask(2) }()
	
	wg.Wait()
	fmt.Printf("elapsed: %v\n", time.Since(start))
}