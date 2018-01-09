package main

import (
	"flag"
	"fmt"
	//"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/test/util"
	"github.com/rhu1/scribble-go-runtime/test/benchmarks/nano/scatter/Scatter/Proto1"
)

const (
	NCPU   = 7
	//NCPU   = 2
	NITERS = 100000
	//NITERS = 1000
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

	//fmt.Println("1")

	 ch := make(chan func() *Proto1.Proto1_S_1To1_End)

	//serverCode := func() (func() *Proto1.Proto1_S_1To1_End) {
	serverCode := func() {

		P1 := Proto1.NewProto1()
		S := P1.NewProto1_S_1To1(ncpu, 1)
		/*if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}*/

		//fmt.Println("S1")

		// One connection for each participant in the group
		for i := 1; i <= ncpu; i++ {

			S.Accept(P1.W, i, util.LOCALHOST, strconv.Itoa(33333+i))

			/*if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}*/
		}

		//fmt.Println("S2")

		s1 := S.Init()

		//fmt.Println("S3")

		//return
		f := func() *Proto1.Proto1_S_1To1_End {
			end := mkservmain(ncpu)(s1)
			wg.Done()
			return end
		}
		ch <- f
	}

	//serverf := serverCode()
	go serverCode()
	time.Sleep(100 * time.Millisecond)

	//fmt.Println("2")

	P1 := Proto1.NewProto1()
	clientCode := func(i int) (func() *Proto1.Proto1_W_1Ton_End) {

		W := P1.NewProto1_W_1Ton(ncpu, i)
		/*if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}*/

		//fmt.Println("W1")

		// One connection for each participant in the group
		W.Connect(P1.S, 1, "127.0.0.1", strconv.Itoa(33333+i))
		w1 := W.Init()
		/*if err != nil {
			log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
		}*/

		//fmt.Println("W2")

		return func() *Proto1.Proto1_W_1Ton_End {
			end := mkworkermain(ncpu)(w1)
			wg.Done()
			return end
		}
	}

	//fmt.Println("3")

	clients := make([]func() *Proto1.Proto1_W_1Ton_End, ncpu)
	for i := 1; i <= ncpu; i++ {
		clients[i-1] = clientCode(i)
	}

	//fmt.Println("4")

	serverf := <-ch	

	//fmt.Println("5")

	run_startt := time.Now()
	go serverf()
	for i := 1; i <= ncpu; i++ {
		go clients[i-1]()
	}
	wg.Wait()
	run_endt := time.Now()
	fmt.Println(Avg(run_endt.Sub(run_startt), niters))
}

func mkservmain(nw int) (func(st1 *Proto1.Proto1_S_1To1_1) *Proto1.Proto1_S_1To1_End) {
	return func(st1 *Proto1.Proto1_S_1To1_1) *Proto1.Proto1_S_1To1_End {
		for i := 0; i < niters; i++ {
			st1 = st1.Send_W_1Ton_(42, splitFn0)
		}
		return nil
	}
}

func splitFn0(x int, i int) int {
	return x + i
}

func mkworkermain(idx int) func(st1 *Proto1.Proto1_W_1Ton_1) *Proto1.Proto1_W_1Ton_End {
	return func(st1 *Proto1.Proto1_W_1Ton_1) *Proto1.Proto1_W_1Ton_End {
		var x int
		for i := 0; i < niters; i++ {
			st1 = st1.Reduce_S_1To1_(&x, foo)
		}
		return nil
	}
}

func foo(xs []int) int {
	return xs[0]	
}
