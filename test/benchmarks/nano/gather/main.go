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

	"github.com/rhu1/scribble-go-runtime/test/benchmarks/nano/gather/Gather/Proto1"
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
	wg.Add(ncpu+1)

	servConns := make([]transport.Transport, ncpu)
	cliConns := make([]transport.Transport, ncpu)
	for j := 0; j < ncpu; j++ {
		/*
		port := strconv.Itoa(33333+j)
		servConns[j] = tcp.NewAcceptor(port)	
		cliConns[j] = tcp.NewRequestor(util.LOCALHOST, port)
		/*/
		servConns[j] = shm.NewConnector()	
		cliConns[j] = servConns[j]	
		//*/
	}

	ch := make(chan func())

	serverCode := func() {
		P1 := Proto1.NewProto1()
		serverIni := P1.NewProto1_S_1To1(ncpu, 1)
		/*if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}*/
		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {
			//err := 
			serverIni.Accept(P1.W, i, servConns[i-1])
			/*if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}*/
		}
		s1 := serverIni.Init()

		f := func() {
			mkservmain(ncpu)(s1)
			wg.Done()
		}
		ch <- f
	}
	go serverCode()
	time.Sleep(100 * time.Millisecond)  // Make sure all server sockets are open before client requests

	clientCode := func(i int) func() {
		P1 := Proto1.NewProto1()
		clientIni := P1.NewProto1_W_1Ton(ncpu, i)
		/*if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}*/
		// One connection for each participant in the group
		//err = 
		clientIni.Request(P1.S, 1, cliConns[i-1])
		w1 := clientIni.Init()
		/*if err != nil {
			log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
		}*/

		return func() {
			mkworkermain(ncpu)(w1)
			wg.Done()
		}
	}

	clients := make([]func(), ncpu)
	for i := 1; i <= ncpu; i++ {
		clients[i-1] = clientCode(i)
	}
	srvf := <-ch

	run_startt := time.Now()
	go srvf()
	for i := 1; i <= ncpu; i++ {
		go clients[i-1]()
	}
	wg.Wait()
	run_endt := time.Now()
	fmt.Println(Avg(run_endt.Sub(run_startt), niters))
}

func mkservmain(nw int) func(st1 *Proto1.Proto1_S_1To1_1) {
	var xs []int
	return func(st1 *Proto1.Proto1_S_1To1_1) {
		for i := 0; i < niters; i++ {
			st1 = st1.Recv_W_1Ton_(&xs)
		}
	}
}

func mkworkermain(idx int) func(st1 *Proto1.Proto1_W_1Ton_1) {
	return func(st1 *Proto1.Proto1_W_1Ton_1) {
		for i := 0; i < niters; i++ {
			st1 = st1.Send_S_1To1_([]int{41+idx})
		}
	}
}
