//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo5
//$ bin/foo5.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/test/util"
	"github.com/rhu1/scribble-go-runtime/test/foo5/Foo5/Proto1"
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
		S.Accept(P1.W, i, util.LOCALHOST, strconv.Itoa(PORT+i))
	}
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	for i := 0; i < 3; i++ {
		s1 = s1.Send_W_1Ton_a(1234, util.Copy)
	}
	end = s1.Send_W_1Ton_b(5678, util.Copy)

	wg.Done()
	return end
}

func clientCode(wg *sync.WaitGroup, n int, self int) *Proto1.Proto1_W_1Ton_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1Ton(n, self)
	W.Connect(P1.S, 1, util.LOCALHOST, strconv.Itoa(PORT+self))
	w1 := W.Init()
	var end *Proto1.Proto1_W_1Ton_End

	for b := true; b; {
		var x int
		select {
		case w1 = <-w1.Recv_S_1To1_a(&x):
			fmt.Println("W got a:", self, x)
		case end = <-w1.Recv_S_1To1_b(&x):
			fmt.Println("W got b:", self, x)
			b = false
		}
	}

	wg.Done()
	return end
}
