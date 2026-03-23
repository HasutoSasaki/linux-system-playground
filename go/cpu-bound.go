package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
    "os"
    "runtime/trace"
)

func cpuTask(id int) {
    fmt.Printf("task %d: start\n", id)
    sum := 0
    for i := 0; i < 2_000_000_000; i++ {
        sum += i
    }
    fmt.Printf("task %d: done (sum=%d)\n", id, sum)
}

func main() {
    runtime.GOMAXPROCS(1) // ← 1と2で切り替えて計測

    f, _ := os.Create("trace.out") // トレース結果の出力先ファイルを作成
    trace.Start(f) // トレース記録開始
    defer trace.Stop() // トレース記録終了
    
    var wg sync.WaitGroup
    start := time.Now()

    wg.Add(2)
    go func() { defer wg.Done(); cpuTask(1) }()
    go func() { defer wg.Done(); cpuTask(2) }()
    wg.Wait()
    fmt.Printf("elapsed: %v\n", time.Since(start))
}