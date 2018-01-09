//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo6
//$ bin/foo6.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/test/foo6/Foo6/Proto1"
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
		S.Accept(P1.W, i, util.LOCALHOST, strconv.Itoa(PORT+i))
	}
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	var xs []int
	for z := 0; z < 3; z++ {
		s2 := s1.Send_W_1Ton_a(1, util.Copy)
		s1 = s2.Send_W_1Ton_b(2, util.Copy).Recv_W_1Ton_c(&xs)
		fmt.Println("S got c:", xs)
	}
	s4 := s1.Send_W_1Ton_a(1, util.Copy)
	s5 := s4.Send_W_1Ton_d(4, util.Copy)
	fmt.Println("S sent d:")

	end = s5.Recv_W_1Ton_e(&xs)
	fmt.Println("S got e:", xs)

	wg.Done()
	return end
}

func clientCode(wg *sync.WaitGroup, n int, self int) *Proto1.Proto1_W_1Ton_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1Ton(1, self)
	W.Connect(P1.S, 1, "127.0.0.1", strconv.Itoa(PORT+self))
	w1 := W.Init()
	var end *Proto1.Proto1_W_1Ton_End

	var xs []int
	var x int
	for b := true; b; {
		w2 := w1.Recv_S_1To1_a(&xs)
		select {
		case w3 := <-w2.Recv_S_1To1_b(&x):
			fmt.Println("W got b:", self, x)
			w1 = w3.Send_S_1To1_c(3, util.Copy)
		case w5 := <-w2.Recv_S_1To1_d(&x):
			fmt.Println("W got d:", self, x)
			end = w5.Send_S_1To1_e(5, util.Copy)
			fmt.Println("W sent e:", self)
			b = false
		}
	}
	fmt.Println("W end:", self)

	wg.Done()
	return end
}
