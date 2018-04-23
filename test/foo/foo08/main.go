//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo8
//$ bin/foo8.exe

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

	"github.com/rhu1/scribble-go-runtime/test/foo/foo8/Foo8/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

// Bypass bloody annoying Go "unused import" errors
var _ = strconv.Itoa
var _ = tcp.NewAcceptor
var _ = shm.NewConnector

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	n := 3

	wg := new(sync.WaitGroup)
	wg.Add(n + 1)

	as := make([]transport.Transport, n)
	for i := 1; i <= n; i++ {
		as[i-1] = tcp.NewAcceptor(strconv.Itoa(PORT+i))
		//as[i-1] = shm.NewConnector()
	}
	go serverCode(wg, n, as)

	time.Sleep(100 * time.Millisecond)

	conn := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+1))
	go W_1_Code(wg, n, conn)

	for i := 2; i <= n; i++ {
		conn = tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+i))
		//conn := as[i-1]
		go W_2Ton_Code(wg, n, i, conn)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, n int, conns []transport.Transport) *Proto1.Proto1_S_1To1_End {
	P1 := Proto1.NewProto1()

	S := P1.NewProto1_S_1To1(n, 1)
	for i := 1; i <= n; i++ {
		S.Accept(P1.W, i, conns[i-1])
	}
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	//var bs []byte
	s2 := s1.Split_W_1To1_a(1234, util.Copy)
	end = s2.Split_W_1Ton_b(5678, util.Copy)
	//fmt.Println("S received:", bs)

	wg.Done()
	return end
}

func W_1_Code(wg *sync.WaitGroup, n int, conn transport.Transport) *Proto1.Proto1_W_1To1and1Ton_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1To1and1Ton(1, 1)
	W.Request(P1.S, 1, conn)
	w1 := W.Init()
	var end *Proto1.Proto1_W_1To1and1Ton_End

	var x int
	w2 := w1.Reduce_S_1To1_a(&x, util.UnaryReduce)
	fmt.Println("W" + strconv.Itoa(1) + ":", x)
	end = w2.Reduce_S_1To1_b(&x, util.UnaryReduce)
	fmt.Println("W" + strconv.Itoa(1) + ":", x)

	wg.Done()
	return end
}

func W_2Ton_Code(wg *sync.WaitGroup, n int, self int, conn transport.Transport) *Proto1.Proto1_W_1Ton_not_1To1_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1Ton_not_1To1(1, self)
	W.Request(P1.S, 1, conn)
	w1 := W.Init()
	var end *Proto1.Proto1_W_1Ton_not_1To1_End

	var x int
	end = w1.Reduce_S_1To1_b(&x, util.UnaryReduce)
	fmt.Println("W" + strconv.Itoa(self) + ":", x)

	wg.Done()
	return end
}
