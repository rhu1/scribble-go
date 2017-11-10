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
var ncpu1, ncpu2 int

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.IntVar(&ncpu, "ncpu", NCPU, "GOMAXPROCS")
	flag.IntVar(&niters, "niters", NITERS, "ITERS")
	flag.Parse()

	ncpu1 = ncpu / 2
	ncpu2 = ncpu - ncpu1
	if ncpu2 == 0 {
		ncpu2 = 1
	}
	wg := new(sync.WaitGroup)
	wg.Add(ncpu)

	conn := make([][]tcp.ConnCfg, ncpu1)
	for i := 0; i < ncpu1; i++ {
		conn[i] = make([]tcp.ConnCfg, ncpu2)
		for j := 0; j < ncpu2; j++ {
			conn[i][j] = tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i*(ncpu2)+j))
		}
	}

	serverCode := func(idx int) func() {

		cnn := make([](*tcp.Conn), ncpu2)
		cnnMu := new(sync.RWMutex)
		cnnWg := new(sync.WaitGroup)
		cnnWg.Add(ncpu2)
		// One connection for each participant in the group
		for i := 1; i <= ncpu2; i++ {
			go func(i int) {
				c := conn[idx][i-1].Accept().(*tcp.Conn)
				cnnMu.Lock()
				cnn[i-1] = c
				cnnMu.Unlock()
				cnnWg.Done()
			}(i)
		}

		return func() {
			cnnWg.Wait()

			for i := 0; i < niters; i++ {
				cnnMu.RLock()
				for _, cn := range cnn {
					cn.Send(42)
				}
				cnnMu.RUnlock()
			}
			wg.Done()
		}
	}

	servers := make([]func(), ncpu1)
	for i := 0; i < ncpu1; i++ {
		servers[i] = serverCode(i)
	}

	for i := 1; i <= 1000000; i++ {
		// time waster instead of time.Sleep
	}

	clientCode := func(idx int) func() {
		var tmp int

		cnn := make([](*tcp.Conn), ncpu1)
		cnnMu := new(sync.RWMutex)
		cnnWg := new(sync.WaitGroup)
		cnnWg.Add(ncpu1)
		// One connection for each participant in the group
		for i := 1; i <= ncpu1; i++ {
			c := conn[i-1][idx].Connect().(*tcp.Conn)
			cnnMu.Lock()
			cnn[i-1] = c
			cnnMu.Unlock()
			cnnWg.Done()
		}

		return func() {
			cnnWg.Wait()

			for i := 0; i < niters; i++ {
				cnnMu.RLock()
				for _, cn := range cnn {
					err := cn.Recv(&tmp)
					if err != nil {
						log.Fatalf("wrong value from server at %d: %s", i, err)
					}
				}
				cnnMu.RUnlock()
			}
			wg.Done()
		}
	}

	clients := make([]func(), ncpu2)
	for i := 0; i < ncpu2; i++ {
		clients[i] = clientCode(i)
	}

	run_startt := time.Now()
	for i := 1; i <= ncpu1; i++ {
		go servers[i-1]()
	}
	for i := 1; i <= ncpu2; i++ {
		go clients[i-1]()
	}
	wg.Wait()
	run_endt := time.Now()
	fmt.Println(Avg(run_endt.Sub(run_startt), niters))
}
