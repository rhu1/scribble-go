//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo12
//$ bin/foo12.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo12/Foo12/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/foo/foo12/Foo12/Proto2"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = strconv.Atoi

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	wg := new(sync.WaitGroup)
	wg.Add(2)
	
	wgSW := new(sync.WaitGroup)
	wgSW.Add(2)

	// Sets up shared memory connection between A and B.
	connAB := shm.NewConnector()
	go runA(connAB, wg, wgSW)

	time.Sleep(200 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	go runB(connAB, wg, wgSW)

	wg.Wait()
	wgSW.Wait()
}

func runB(conn shm.ConnCfg, wg *sync.WaitGroup, wgSW *sync.WaitGroup) (*Proto1.Proto1_B_1To1_End, *Proto2.Proto2_W_1To1_End) {
	P1 := Proto1.NewProto1()

	B := P1.NewProto1_B_1To1(1)
	//conn := tcp.NewAcceptor(strconv.Itoa(PORT))  // FIXME: check shm for deleg
	// FIXME: check shm for deleg
	/*if err != nil {
		log.Fatalf("failed to create connection to W %d: %v", i, err)
	}*/
	B.Accept(P1.A, 1, conn)
	b1 := B.Init()
	var endB *Proto1.Proto1_B_1To1_End
	var endW *Proto2.Proto2_W_1To1_End

	fmt.Println("B: initialised")

	var w1 *Proto2.Proto2_W_1To1_1
	endB = b1.Reduce_A_1To1_a(w1)
	fmt.Println("B: received W")

	var x int
	endW = w1.Reduce_S_1To1_b(&x, util.UnaryReduce)

	fmt.Println("W: ", x)

	wg.Done()
	wgSW.Done()
	return endB, endW
}

func runA(conn shm.ConnCfg, wg *sync.WaitGroup, wgSW *sync.WaitGroup) *Proto1.Proto1_A_1To1_End {
	// Sets up connection between W and S.
	connWS := shm.NewConnector()

	// connWS (endpoint S) is passed to S spawned here
	// At this point connWS is dangling - not connected to W nor S.
	go runS(connWS, wgSW)

	P2 := Proto2.NewProto2()

	W := P2.NewProto2_W_1To1(1)
	W.Request(P2.S, 1, connWS)
	var w1 *Proto2.Proto2_W_1To1_1 = W.Init()

	fmt.Println("W: initialised")

	P1 := Proto1.NewProto1()

	A := P1.NewProto1_A_1To1(1)
	A.Request(P1.B, 1, conn)
	/*if err != nil {
		log.Fatalf("failed to create connection to Auctioneer: %v", err)
	}*/
	var a1 *Proto1.Proto1_A_1To1_1 = A.Init()

	fmt.Println("A: initialised")

	fmt.Println("A: sending")
	end := a1.Split_B_1To1_a(w1, func(w1 *Proto2.Proto2_W_1To1_1, i int) *Proto2.Proto2_W_1To1_1 { return w1 })
	fmt.Println("A: sent W")

	wg.Done()
	return end
}

func runS(conn shm.ConnCfg, wgSW *sync.WaitGroup) *Proto2.Proto2_S_1To1_End {
	P2 := Proto2.NewProto2()

	S := P2.NewProto2_S_1To1(1)
	/*if err != nil {
		log.Fatalf("failed to create connection to W %d: %v", i, err)
	}*/
	S.Accept(P2.W, 1, conn)
	s1 := S.Init()
	var endS *Proto2.Proto2_S_1To1_End

	fmt.Println("S: initialised")

	x := 123
	endS = s1.Split_W_1To1_b(x, util.Copy)
	fmt.Println("S: sent", x)

	wgSW.Done()
	return endS
}
