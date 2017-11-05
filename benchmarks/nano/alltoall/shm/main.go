package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"../alltoall"

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
	wg.Add(2 * ncpu)

	conn := make([][]transport.Transport, ncpu)
	for i := 0; i < ncpu; i++ {
		conn[i] = make([]transport.Transport, ncpu)
		for j := 0; j < ncpu; j++ {
			conn[i][j] = shm.NewBufferedConnection(100)
		}
	}

	serverCode := func(idx int) {
		serverIni, err := alltoall.NewServer(idx, ncpu, ncpu)
		if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}
		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {
			err := session.Accept(serverIni, alltoall.Worker, i, conn[idx-1][i-1])
			if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}
		}

		serverMain := mkservmain(idx, ncpu)
		serverIni.Run(serverMain)
		wg.Done()
	}

	for i := 1; i <= ncpu; i++ {
		go serverCode(i)
	}
	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) {
		clientIni, err := alltoall.NewWorker(i, ncpu, ncpu)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group

		for j := 1; j <= ncpu; j++ {
			err = session.Connect(clientIni, alltoall.Server, j, conn[j-1][i-1])
			if err != nil {
				log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
			}
		}

		clientMain := mkworkermain(i)
		clientIni.Run(clientMain)
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

func mkservmain(idx, nw int) func(st1 *alltoall.Server_1Ton_1) *alltoall.Server_1Ton_End {
	payload := make([]int, ncpu)
	for i := 0; i < ncpu; i++ {
		payload[i] = idx*42 + i
	}
	return func(st1 *alltoall.Server_1Ton_1) *alltoall.Server_1Ton_End {
		for i := 0; i < niters; i++ {
			st1 = st1.SendAll(payload)
		}
		return nil
	}
}

func mkworkermain(idx int) func(st1 *alltoall.Worker_1Ton_1) *alltoall.Worker_1Ton_End {
	return func(st1 *alltoall.Worker_1Ton_1) *alltoall.Worker_1Ton_End {
		for i := 0; i < niters; i++ {
			_, st1 = st1.RecvAll()
		}
		return nil
	}
}
