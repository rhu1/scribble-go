package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
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
	wg.Add(ncpu + 1)

	serverCode := func() func() {

		cnn := make([](*tcp.Conn), ncpu)
		cnnMu := new(sync.RWMutex)
		cnnWg := new(sync.WaitGroup)
		cnnWg.Add(ncpu)
		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {
			go func(i int) {
				conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
				c := conn.Accept().(*tcp.Conn)
				cnnMu.Lock()
				cnn[i-1] = c
				cnnMu.Unlock()
				cnnWg.Done()
			}(i)
		}

		payload := make([]int, ncpu)
		for i := 0; i < ncpu; i++ {
			payload[i] = 42 + i
		}

		return func() {
			cnnWg.Wait()

			for i := 0; i < niters; i++ {
				cnnMu.RLock()
				for j, v := range payload {
					cnn[j].Send(v)
				}
				cnnMu.RUnlock()
			}
			wg.Done()
		}
	}

	serverf := serverCode()
	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) func() {
		var tmp int

		conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
		cnn := conn.Connect()

		return func() {
			for i := 0; i < niters; i++ {

				err := cnn.Recv(&tmp)
				if err != nil {
					log.Fatalf("wrong value from server at %d: %s", i, err)
				}
			}
			wg.Done()
		}
	}

	clients := make([]func(), ncpu)
	for i := 1; i <= ncpu; i++ {
		clients[i-1] = clientCode(i)
	}

	run_startt := time.Now()
	go serverf()
	for i := 1; i <= ncpu; i++ {
		go clients[i-1]()
	}
	wg.Wait()
	run_endt := time.Now()
	fmt.Println(Avg(run_endt.Sub(run_startt), niters))
}
