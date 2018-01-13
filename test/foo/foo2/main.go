//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo2
//$ bin/foo2.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo2/Foo2/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	n := 2

	wg := new(sync.WaitGroup)
	wg.Add(n + 1)

	go serverCode(wg, n)

	time.Sleep(100 * time.Millisecond)

	for i := 1; i <= n; i++ {
		go clientCode(wg, n, i)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, n int) *Proto1.Proto1_S_1To1_End {
	P1 := Proto1.NewProto1()

	S := P1.NewProto1_S_1To1(n, 1)
	for i := 1; i <= n; i++ {
		conn := tcp.NewAcceptor(strconv.Itoa(PORT+i))
		S.Accept(P1.W, i, conn)
	}
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	s2 := s1.Split_W_1Ton_a(1234, util.Copy)

	var xs []int
	end = s2.Recv_W_1Ton_b(&xs)
	fmt.Println("S Received:", xs)

	wg.Done()
	return end
}

func clientCode(wg *sync.WaitGroup, n int, self int) *Proto1.Proto1_W_1Ton_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1Ton(n, self)
	conn := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+self))
	W.Request(P1.S, 1, conn)
	var w1 *Proto1.Proto1_W_1Ton_1 = W.Init()

	var x int
	w2 := w1.Reduce_S_1To1_a(&x, util.Sum)
	fmt.Println("W Received: ", self, x)

	end := w2.Split_S_1To1_b(self*100, util.Copy)

	wg.Done()
	return end
}
