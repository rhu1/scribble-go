//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo9
//$ bin/foo9.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo9/Foo9/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

// Bypass bloody annoying Go "unused import" errors
var _ = strconv.Itoa
var _ = tcp.NewAcceptor
var _ = shm.NewConnector

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	//*
	port := strconv.Itoa(PORT)
	servConn := tcp.NewAcceptor(port)
	cliConn := tcp.NewRequestor(util.LOCALHOST, port)
	/*/
	servConn := shm.NewConnector()
	cliConn := servConn
	//*/

	go W_1(wg, servConn)

	time.Sleep(100 * time.Millisecond)

	go W_2(wg, cliConn)

	wg.Wait()
}

func W_1(wg *sync.WaitGroup, conn transport.Transport) *Proto1.Proto1_W_1To1_not_2To2_End {
	P1 := Proto1.NewProto1()

	W1 := P1.NewProto1_W_1To1_not_2To2(1)
	W1.Accept(P1.W, 2, conn)
	s1 := W1.Init()
	var end *Proto1.Proto1_W_1To1_not_2To2_End

	var x []int
	s2 := s1.Split_W_2To2_a(1234, util.Copy)
	end = s2.Recv_W_2To2_b(&x)
	fmt.Println("W1:", x)

	wg.Done()
	return end
}

func W_2(wg *sync.WaitGroup, conn transport.Transport) *Proto1.Proto1_W_2To2_not_1To1_End {
	P1 := Proto1.NewProto1()

	W2 := P1.NewProto1_W_2To2_not_1To1(2)
	W2.Request(P1.W, 1, conn)
	s1 := W2.Init()
	var end *Proto1.Proto1_W_2To2_not_1To1_End

	var x []int
	s2 := s1.Recv_W_1To1_a(&x)
	fmt.Println("W2:", x)
	end = s2.Send_W_1To1_b([]int{x[0]+1})  // FIXME: W1: [0 1]

	wg.Done()
	return end
}
