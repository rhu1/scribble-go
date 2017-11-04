package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/nickng/scribble-go/runtime/transport/tcp"
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

	conn := make([][]tcp.ConnCfg, ncpu)
	for i := 0; i < ncpu; i++ {
		conn[i] = make([]tcp.ConnCfg, ncpu)
		for j := 0; j < ncpu; j++ {
			conn[i][j] = tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i*ncpu+j))
		}
	}

	serverCode := func(idx int) {

		cnn := make([](*tcp.Conn), ncpu)
		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {
			go func(i int) {
				cnn[i-1] = conn[idx][i-1].Accept().(*tcp.Conn)
			}(i)
		}

		for i := 0; i < ncpu; i++ {
			for cnn[i] == nil {
			}
		}

		for i := 0; i < niters; i++ {
			for _, cn := range cnn {
				cn.Send(42)
			}
		}
		wg.Done()
	}

	for i := 0; i < ncpu; i++ {
		go serverCode(i)
	}
	time.Sleep(100 * time.Millisecond)

	clientCode := func(idx int) {
		var tmp int

		cnn := make([](*tcp.Conn), ncpu)
		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {
			cnn[i-1] = conn[i-1][idx-1].Connect().(*tcp.Conn)
		}

		for i := 0; i < niters; i++ {
			for _, cn := range cnn {
				err := cn.Recv(&tmp)
				if err != nil {
					log.Fatalf("wrong value from server at %d: %s", i, err)
				}
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
	fmt.Println(ncpu, "\t", Avg(run_endt.Sub(run_startt), niters))
}
