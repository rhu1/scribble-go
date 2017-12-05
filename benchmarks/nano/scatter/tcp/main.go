package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/nickng/scribble-go-runtime/benchmarks/nano/scatter/scatter"

	"github.com/nickng/scribble-go-runtime/runtime/session"
	"github.com/nickng/scribble-go-runtime/runtime/transport/tcp"
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
		serverIni, err := scatter.NewServer(1, 1, ncpu)
		if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}
		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {
			conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
			err := session.Accept(serverIni, scatter.Worker, i, conn)
			if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}
		}

		serverMain := mkservmain(ncpu)
		return func() {
			serverIni.Run(serverMain)
			wg.Done()
		}
	}

	serverf := serverCode()
	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) func() {
		clientIni, err := scatter.NewWorker(i, ncpu, 1)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group
		conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
		err = session.Connect(clientIni, scatter.Server, 1, conn)
		if err != nil {
			log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
		}

		clientMain := mkworkermain(i)
		return func() {
			clientIni.Run(clientMain)
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

func mkservmain(nw int) func(st1 *scatter.Server_1To1_1) *scatter.Server_1To1_End {
	payload := make([]int, nw)
	for i := 0; i < nw; i++ {
		payload[i] = 42 + i
	}
	return func(st1 *scatter.Server_1To1_1) *scatter.Server_1To1_End {
		for i := 0; i < niters; i++ {
			st1 = st1.SendAll(payload)
		}
		return nil
	}
}

func mkworkermain(idx int) func(st1 *scatter.Worker_1Ton_1) *scatter.Worker_1Ton_End {
	return func(st1 *scatter.Worker_1Ton_1) *scatter.Worker_1Ton_End {
		for i := 0; i < niters; i++ {
			_, st1 = st1.RecvAll()
		}
		return nil
	}
}
