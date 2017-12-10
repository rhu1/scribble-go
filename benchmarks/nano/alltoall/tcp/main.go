package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/nickng/scribble-go-runtime/benchmarks/nano/alltoall/alltoall"

	"github.com/nickng/scribble-go-runtime/runtime/session"
	"github.com/nickng/scribble-go-runtime/runtime/transport"
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
var cpu1, cpu2 int

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.IntVar(&ncpu, "ncpu", NCPU, "GOMAXPROCS")
	flag.IntVar(&niters, "niters", NITERS, "ITERS")
	flag.Parse()

	cpu1 = ncpu / 2
	cpu2 = ncpu - cpu1
	if cpu2 == 0 {
		cpu2 = 1
	}
	wg := new(sync.WaitGroup)
	wg.Add(ncpu)

	conn := make([][]transport.Transport, cpu1)
	for i := 0; i < cpu1; i++ {
		conn[i] = make([]transport.Transport, cpu2)
		for j := 0; j < cpu2; j++ {
			conn[i][j] = tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i*cpu2+j))
		}
	}

	serverCode := func(idx int) func() {
		serverIni, err := alltoall.NewServer(idx, cpu1, cpu2)
		if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}
		// One connection for each participant in the group
		for i := 1; i <= cpu2; i++ {
			err := session.Accept(serverIni, alltoall.Worker, i, conn[idx-1][i-1])
			if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}
		}

		serverMain := mkservmain(idx, ncpu)
		return func() {
			serverIni.Run(serverMain)
			wg.Done()
		}
	}

	servers := make([]func(), cpu1)
	for i := 1; i <= cpu1; i++ {
		servers[i-1] = serverCode(i)
	}

	for i := 1; i <= 1000000; i++ {
		// time waster instead of time.Sleep
	}

	clientCode := func(i int) func() {
		clientIni, err := alltoall.NewWorker(i, cpu2, cpu1)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group

		for j := 1; j <= cpu1; j++ {
			err = session.Connect(clientIni, alltoall.Server, j, conn[j-1][i-1])
			if err != nil {
				log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
			}
		}

		clientMain := mkworkermain(i)
		return func() {
			clientIni.Run(clientMain)
			wg.Done()
		}
	}

	clients := make([]func(), cpu2)
	for i := 1; i <= cpu2; i++ {
		clients[i-1] = clientCode(i)
	}

	run_startt := time.Now()
	for i := 1; i <= cpu1; i++ {
		go servers[i-1]()
	}
	for i := 1; i <= cpu2; i++ {
		go clients[i-1]()
	}
	wg.Wait()
	run_endt := time.Now()
	fmt.Println(Avg(run_endt.Sub(run_startt), niters))
}

func mkservmain(idx, nw int) func(st1 *alltoall.Server_1Ton_1) *alltoall.Server_1Ton_End {
	payload := make([]int, cpu2)
	for i := 0; i < cpu2; i++ {
		payload[i] = idx*42 + i
	}
	return func(st1 *alltoall.Server_1Ton_1) *alltoall.Server_1Ton_End {
		for i := 0; i < niters; i++ {
			fmt.Println("Sent payload ", payload, " from ", idx)
			st1 = st1.SendAll(payload)
		}
		return nil
	}
}

func mkworkermain(idx int) func(st1 *alltoall.Worker_1Ton_1) *alltoall.Worker_1Ton_End {
	return func(st1 *alltoall.Worker_1Ton_1) *alltoall.Worker_1Ton_End {
		var v []int
		for i := 0; i < niters; i++ {
			v, st1 = st1.RecvAll()
			fmt.Println("Received payload ", v, "at ", idx)
		}
		return nil
	}
}
