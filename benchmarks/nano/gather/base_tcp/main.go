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
	wg.Add(ncpu + 1)

	serverCode := func() func() {

		cnn := make([](*tcp.Conn), ncpu)
		rwm := new(sync.RWMutex)
		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {
			go func(i int) {
				conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
				rwm.Lock()
				cnn[i-1] = conn.Accept().(*tcp.Conn)
				rwm.Unlock()
			}(i)
		}

		return func() {
			var tmp int
			for i := 0; i < ncpu; i++ {
				rwm.RLock()
				for cnn[i] == nil {
				}
				rwm.RUnlock()
			}

			for i := 0; i < niters; i++ {
				for _, cn := range cnn {
					rwm.RLock()
					cn.Recv(&tmp)
					rwm.RUnlock()
				}
			}
			wg.Done()
		}
	}

	srvmain := serverCode()
	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) func() {
		tmp := 41 + i

		conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
		cnn := conn.Connect()

		return func() {
			for i := 0; i < niters; i++ {

				err := cnn.Send(tmp)
				if err != nil {
					log.Fatalf("wrong value from server at %d: %s", i, err)
				}
			}
			wg.Done()
		}
	}

	clients := make([](func()), ncpu)
	for i := 1; i <= ncpu; i++ {
		clients[i-1] = clientCode(i)
	}

	run_startt := time.Now()
	go srvmain()
	for i := 1; i <= ncpu; i++ {
		go clients[i-1]()
	}
	wg.Wait()
	run_endt := time.Now()
	fmt.Println(Avg(run_endt.Sub(run_startt), niters))
}
