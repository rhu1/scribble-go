package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"../gather"

	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/shm"
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

	conn := make([]transport.Transport, ncpu)
	for i := 0; i < ncpu; i++ {
		conn[i] = shm.NewBufferedConnection(100)
	}

	serverIni, err := gather.NewServer(1, 1, ncpu)
	if err != nil {
		log.Fatalf("cannot create server endpoint: %s", err)
	}
	// One connection for each participant in the group
	for i := 1; i <= ncpu; i++ {
		err := session.Accept(serverIni, gather.Worker, i, conn[i-1])
		if err != nil {
			log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
		}
	}
	serverMain := mkservmain(ncpu)

	time.Sleep(100 * time.Millisecond)

	serverCode := func() {
		serverIni.Run(serverMain)
		wg.Done()
	}

	clientCode := func(i int) func() {
		clientIni, err := gather.NewWorker(i, ncpu, 1)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group
		err = session.Connect(clientIni, gather.Server, 1, conn[i-1])
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
	go serverCode()
	for i := 0; i < ncpu; i++ {
		go clients[i]()
	}
	wg.Wait()
	run_endt := time.Now()
	// fmt.Println(ncpu, "\t", Avg(run_endt.Sub(run_startt), niters))
	fmt.Println(Avg(run_endt.Sub(run_startt), niters))
}

func mkservmain(nw int) func(st1 *gather.Server_1To1_1) *gather.Server_1To1_End {
	return func(st1 *gather.Server_1To1_1) *gather.Server_1To1_End {
		for i := 0; i < niters; i++ {
			_, st1 = st1.RecvAll()
		}
		return nil
	}
}

func mkworkermain(idx int) func(st1 *gather.Worker_1Ton_1) *gather.Worker_1Ton_End {
	return func(st1 *gather.Worker_1Ton_1) *gather.Worker_1Ton_End {
		for i := 0; i < niters; i++ {
			st1 = st1.SendAll(41 + idx)
		}
		return nil
	}
}
