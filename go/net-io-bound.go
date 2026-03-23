package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func startServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/delay", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Write([]byte("ok"))
	})
	srv := &http.Server{Addr: ":18080", Handler: mux}
	go srv.ListenAndServe()
	time.Sleep(100 * time.Millisecond)
	return srv
}

func fetch(id int) {
	resp, err := http.Get("http://localhost:18080/delay")
	if err != nil {
		fmt.Printf("task %d: error: %v\n", id, err)
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

func main() {
	procs, _ := strconv.Atoi(os.Args[1])
	numTasks, _ := strconv.Atoi(os.Args[2])
	runtime.GOMAXPROCS(procs)

	srv := startServer()
	defer srv.Close()

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fetch(id)
		}(i)
	}

	wg.Wait()
	fmt.Printf("GOMAXPROCS=%d  goroutines=%d  elapsed=%v\n", procs, numTasks, time.Since(start))
}
