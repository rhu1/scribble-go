package main

import (
	"fmt"
	"sync"
	"strconv"
	"time"

	"github.com/rhu1/scribble-go-runtime/test/foo6/Foo6/Proto1"
)

func mysum(xs []int) int {
	res := 0
	for j := 0; j < len(xs); j++ {
		res = res + xs[j]	
	}
	return res
}

func mydup(data int, i int) int {
	return data
}

func main() {
	n := 1

	wg := new(sync.WaitGroup)
	wg.Add(n+1)

	serverCode := func() *Proto1.Proto1_S_1To1_End {
		P1 := Proto1.NewProto1()

		serverIni := P1.NewProto1_S_1To1(1, n)
		for i := 1; i <= n; i++ {
			serverIni.Accept(P1.W, i, "127.0.0.1", strconv.Itoa(33333+i))  // FIXME: ensure ports open before clients request?
		}

		var s1 *Proto1.Proto1_S_1To1_1 = serverIni.Init()

		var xs []int
		var x int
		for z := 0; z < 3; z++ {
			s2 := s1.Send_W_1Ton_a(1, mydup)
			s1 = s2.Send_W_1Ton_b(2, mydup).Recv_W_1Ton_c(&xs)
			fmt.Println("S got c:", xs)
		}
		s4 := s1.Send_W_1Ton_a(1, mydup)
		s5 := s4.Send_W_1Ton_d(4, mydup)
		fmt.Println("S sent d:")
		end := s5.Reduce_W_1Ton_e(&x, mysum)
		fmt.Println("S got e:", x)

		wg.Done()
		return end
	}

	go serverCode()

	clientCode := func(i int) *Proto1.Proto1_W_1Ton_End {
		P1 := Proto1.NewProto1()

		clientIni := P1.NewProto1_W_1Ton(i, 1)
		clientIni.Connect(P1.S, 1, "127.0.0.1", strconv.Itoa(33333+i))

		var c1 *Proto1.Proto1_W_1Ton_1 = clientIni.Init()
		var end *Proto1.Proto1_W_1Ton_End

		var xs[] int	
		var x int
		for b := true; b; {
			c2 := c1.Recv_S_1To1_a(&xs)
			select {
			case c3 := <-c2.Recv_S_1To1_b(&x):
				fmt.Println("W got b:", i, x)
				c1 = c3.Send_S_1To1_c(3, mydup)
			case c5 := <-c2.Recv_S_1To1_d(&x):
				fmt.Println("W got d:", i, x)
				end = c5.Send_S_1To1_e(5, mydup)
				fmt.Println("W sent e:", i)
				b = false
			}
		}
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("W end:", i)

		wg.Done()
		return end
	}

	for i := 1; i <= n; i++ {
		go clientCode(i)
	}

	wg.Wait()
}

