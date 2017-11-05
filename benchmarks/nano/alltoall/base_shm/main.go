package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	NCPU   = 7
	NITERS = 100000
)

func Avg(d time.Duration, v int) float64 {
	return float64(d.Nanoseconds()) / float64(v)
}

var ncpu, niters int

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.IntVar(&ncpu, "ncpu", NCPU, "GOMAXPROCS")
	flag.IntVar(&niters, "niters", NITERS, "ITERS")
	flag.Parse()
	wg := new(sync.WaitGroup)
	wg.Add(2 * ncpu)

	conn := make([][]chan int, ncpu)
	for i := 0; i < ncpu; i++ {
		conn[i] = make([]chan int, ncpu)
		for j := 0; j < ncpu; j++ {
			conn[i][j] = make(chan int, 100)
		}
	}

	serverCode := func(idx int) {

		for i := 0; i < niters; i++ {
			for j := 0; j < ncpu; j++ {
				conn[idx][j] <- (42*j + idx)
			}
		}

		wg.Done()
	}

	for i := 0; i < ncpu; i++ {
		go serverCode(i)
	}
	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) {

		for j := 0; j < niters; j++ {
			for k := 0; k < ncpu; k++ {
				<-conn[k][i-1]
			}
		}
		wg.Done()
	}

	run_startt := time.Now()
	for i := 1; i <= ncpu; i++ {
		go clientCode(i)
	}
	wg.Wait()
	run_endt := time.Now()
	fmt.Println(Avg(run_endt.Sub(run_startt), niters))
}
