package main

import (
	/*"log"
	"time"*/
	"fmt"
	"sync"
	"strconv"

	"github.com/rhu1/scribble-go-runtime/test/foo4/Foo4/Proto1"
)


func main() {
	n := 3

	wg := new(sync.WaitGroup)
	wg.Add(n+1)

	serverCode := func() {
		P1 := Proto1.NewProto1()

		serverIni := P1.NewProto1_S_1To1(1, n)
		for i := 1; i <= n; i++ {
			serverIni.Accept(P1.W, i, "127.0.0.1", strconv.Itoa(33333+i))  // FIXME: ensure ports open before clients request?
		}

		var s1 *Proto1.Proto1_S_1To1_1 = serverIni.Init()

		s2 := s1.Send_W_1Ton_a(1234, func(data int, i int) int { return data })

		sum := func(xs []int) int {
			res := 0
			for j := 0; j < len(xs); j++ {
				res = res + xs[j]	
			}
			return res
		}

		var x int
		if (1 < 2) {
			s2.Send_W_1Ton_b(1234, func(data int, i int) int { return data }).Reduce_W_1Ton_c(&x, sum)
			fmt.Println("S got c:", x)
		} else {
			s2.Send_W_1Ton_d(5678, func(data int, i int) int { return data })//.Recv_W_1Ton_b(&x, sum)
		}

		wg.Done()
	}

	go serverCode()

	clientCode := func(i int) {
		P1 := Proto1.NewProto1()

		clientIni := P1.NewProto1_W_1Ton(i, 1)
		clientIni.Connect(P1.S, 1, "127.0.0.1", strconv.Itoa(33333+i))

		var c_i *Proto1.Proto1_W_1Ton_1 = clientIni.Init()

		var y int
		c2 := c_i.Reduce_S_1To1_a(&y, func(data []int) int { return data[0] })

		var 
x int	
		select {
		case c3 := <-c2.Recv_S_1To1_b(&x):
			fmt.Println("W got b:", i, x)
			c3.Send_S_1To1_c(5678, func(data int, i int) int { return data })
		case <-c2.Recv_S_1To1_d(&x):
			fmt.Println("W got d:", i, x)
		}

		wg.Done()
	}

	for i := 1; i <= n; i++ {
		go clientCode(i)
	}

	wg.Wait()
}

/*func main() {
	wg := new(sync.WaitGroup)
	wg.Add(11)

	serverCode := func() {
		serverIni, err := scatter.NewServer(1, 1, 10)
		if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}
		// One connection for each participant in the group
		for i := 1; i <= 10; i++ {
			err := serverIni.Accept(scatter.Worker, i, "127.0.0.1", strconv.Itoa(33333+i))
			if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}
		}

		serverMain := mkservmain(10)
		serverIni.Run(serverMain)  // serverMain: init -> end
		wg.Done()
	}

	go serverCode()

	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) {
		clientIni, err := scatter.NewWorker(i, 10, 1)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group
		err = clientIni.Connect(scatter.Server, 1, "127.0.0.1", strconv.Itoa(33333+i))
		if err != nil {
			log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
		}

		clientMain := mkworkermain(i)
		clientIni.Run(clientMain)
		wg.Done()
	}

	for i := 1; i <= 10; i++ {
		go clientCode(i)
	}

	wg.Wait()
}

func mkservmain(nw int) func(st1 *scatter.Server_1To1_1) *scatter.Server_1To1_End {
	payload := make([]int, nw)
	for i := 0; i < nw; i++ {
		payload[i] = 42 + i
	}
	return func(st1 *scatter.Server_1To1_1) *scatter.Server_1To1_End {
		return st1.SendAll(payload)
	}
}

func mkworkermain(idx int) func(st1 *scatter.Worker_1Ton_1) *scatter.Worker_1Ton_End {
	return func(st1 *scatter.Worker_1Ton_1) *scatter.Worker_1Ton_End {
		r, st := st1.RecvAll()
		fmt.Println("Received payload at worker ", idx, "\t: ", r[0])
		return st
	}
}*/
