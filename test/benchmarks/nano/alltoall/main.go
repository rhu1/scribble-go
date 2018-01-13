package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/benchmarks/nano/alltoall/AllToAll/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = log.Fatal
var _ = strconv.Itoa
var _ = shm.NewConnector
var _ = tcp.NewAcceptor
var _ = util.Copy

const (
	NCPU   = 7
	NITERS = 100000
	/*NCPU   = 4
	NITERS = 100*/
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

	servConns := make([][]transport.Transport, cpu1)
	cliConns := make([][]transport.Transport, cpu1)
	for j := 0; j < cpu1; j++ {
		servConns[j] = make([]transport.Transport, cpu2)	
		cliConns[j] = make([]transport.Transport, cpu2)	
		for k := 0; k < cpu2; k++ {
			//*
			port := strconv.Itoa(33333+j*cpu2+k)
			servConns[j][k] = tcp.NewAcceptor(port)	
			cliConns[j][k] = tcp.NewRequestor(util.LOCALHOST, port)
			/*/
			servConns[j][k] = shm.NewConnector()	
			cliConns[j][k] = servConns[j][k]	
			//*/
		}
	}

	chs := make([]chan func(), cpu1)

	serverCode := func(idx int) {
		P1 := Proto1.NewProto1()
		serverIni := P1.NewProto1_S_1Tom(cpu1, cpu2, idx)
		/*if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}*/
		// One connection for each participant in the group
		for i := 1; i <= cpu2; i++ {
			//err := 
			serverIni.Accept(P1.W, i, servConns[idx-1][i-1])
			/*if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}*/
		}
		s1 := serverIni.Init()

		f := func() {
			mkservmain(idx)(s1)
			wg.Done()
		}
		chs[idx-1] <- f
	}

	for i := 1; i <= cpu1; i++ {
		chs[i-1] = make(chan func())
		go serverCode(i)
	}
	/*for i := 1; i <= 1000000; i++ {
		// time waster instead of time.Sleep
	}*/
	time.Sleep(200 * time.Millisecond)  // Make sure all server sockets are open before client requests

	clientCode := func(i int) func() {
		P1 := Proto1.NewProto1()
		clientIni := P1.NewProto1_W_1Ton(cpu1, cpu2, i)
		/*if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}*/
		// One connection for each participant in the group
		for j := 1; j <= cpu1; j++ {
			clientIni.Request(P1.S, j, cliConns[j-1][i-1])
			/*if err != nil {
				log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
			}*/
		}
		w1 := clientIni.Init()

		return func() {
			mkworkermain(i)(w1)
			wg.Done()
		}
	}

	clients := make([]func(), cpu2)
	for i := 1; i <= cpu2; i++ {
		clients[i-1] = clientCode(i)
	}

	servers := make([]func(), cpu1)
	for i := 1; i <= cpu1; i++ {
		servers[i-1] = <-chs[i-1]
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

func mkservmain(idx int) func(st1 *Proto1.Proto1_S_1Tom_1) {
	payload := make([]int, cpu2)
	for i := 0; i < cpu2; i++ {
		payload[i] = idx*42 + i
	}
	return func(st1 *Proto1.Proto1_S_1Tom_1) {
		for i := 0; i < niters; i++ {
			//fmt.Println("Sent payload ", payload, " from ", idx)
			st1 = st1.Send_W_1Ton_(payload)
		}
	}
}

func mkworkermain(idx int) func(st1 *Proto1.Proto1_W_1Ton_1) {
	return func(st1 *Proto1.Proto1_W_1Ton_1) {
		var v []int
		for i := 0; i < niters; i++ {
			st1 = st1.Recv_S_1Tom_(&v)
			//fmt.Println("Received payload ", v, " at ", idx)
		}
	}
}
