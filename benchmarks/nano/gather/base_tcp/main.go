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

	serverCode := func() {

		cnn := make([](*tcp.Conn), ncpu)
		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {
			go func(i int) {
				conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
				cnn[i-1] = conn.Accept().(*tcp.Conn)
			}(i)
		}

		for i := 0; i < ncpu; i++ {
			for cnn[i] == nil {
			}
		}

		var tmp int

		for i := 0; i < niters; i++ {
			for _, cn := range cnn {
				cn.Recv(&tmp)
			}
		}
		wg.Done()
	}

	go serverCode()
	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) {
		tmp := 41 + i

		conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
		cnn := conn.Connect()

		for i := 0; i < niters; i++ {

			err := cnn.Send(tmp)
			if err != nil {
				log.Fatalf("wrong value from server at %d: %s", i, err)
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
