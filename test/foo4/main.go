//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo4
//$ bin/foo4.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/test/foo4/Foo4/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	n := 3

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
		S.Accept(P1.W, i, util.LOCALHOST, strconv.Itoa(PORT+i))
	}
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	s2 := s1.Send_W_1Ton_a(1234, util.Copy)

	var x int
	if 1 < 2 {
		end = s2.Send_W_1Ton_b(1234, util.Copy).Reduce_W_1Ton_c(&x, util.Sum)
		fmt.Println("S got c:", x)
	} else {
		end = s2.Send_W_1Ton_d(5678, util.Copy)
	}

	wg.Done()
	return end
}

func clientCode(wg *sync.WaitGroup, n int, self int) *Proto1.Proto1_W_1Ton_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1Ton(n, self)
	W.Connect(P1.S, 1, util.LOCALHOST, strconv.Itoa(PORT+self))
	w1 := W.Init()
	var end *Proto1.Proto1_W_1Ton_End

	var y int
	w2 := w1.Reduce_S_1To1_a(&y, util.Sum)

	var x int
	select {
	case w3 := <-w2.Recv_S_1To1_b(&x):
		fmt.Println("W got b:", self, x)
		end = w3.Send_S_1To1_c(5678, util.Copy)
	case end = <-w2.Recv_S_1To1_d(&x):
		fmt.Println("W got d:", self, x)
	}

	wg.Done()
	return end
}
