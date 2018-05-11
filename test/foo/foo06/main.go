//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo06
//$ bin/foo06.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo06/Foo6/Proto1"
	S_1 "github.com/rhu1/scribble-go-runtime/test/foo/foo06/Foo6/Proto1/S_1to1"
	"github.com/rhu1/scribble-go-runtime/test/foo/foo06/Foo6/Proto1/W_1toK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

// Bypass bloody annoying Go "unused import" errors
var _ = strconv.Itoa
var _ = tcp.Dial
var _ = shm.Dial

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K + 1)

	go server(wg, K)

	time.Sleep(100 * time.Millisecond)

	for j := 1; j <= K; j++ {
		go client(wg, K, j)
	}

	wg.Wait()
}

func server(wg *sync.WaitGroup, K int) *S_1.End {
	var err error
	P1 := Proto1.New()
	S := P1.New_S_1to1(K, 1)
	as := make([]*tcp.TcpListener, K)
	//as := make([]*shm.ShmListener, K)
	for j := 1; j <= K; j++ {
		as[j-1], err = tcp.Listen(PORT+j)
		//as[j-1], err = shm.Listen(PORT+j)
		if err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	for j := 1; j <= K; j++ {
		err := S.W_1toK_Accept(j, as[j-1], 
			new(session2.GobFormatter))
			//new(session2.PassByPointer))
		if err != nil {
			panic(err)
		}
	}
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1.Init) S_1.End {
	data := []int{ 2, 3, 5, 7, 11, 13, 17, 19, 23 }
	pay := data[0:s.Ept.K]

	for z := 0; z < 3; z++ {
		s = s.W_1toK_Scatter_A(pay).
		      W_1toK_Scatter_B(pay).
		      W_1toK_Gather_C(pay)
		fmt.Println("S gathered C:", pay)
	}
	s4 := s.W_1toK_Scatter_A(pay).
	        W_1toK_Scatter_D(pay)
	fmt.Println("S scattered D:", pay)

	end := s4.W_1toK_Gather_E(pay)
	fmt.Println("S gathered E:", pay)
	return *end
}

func client(wg *sync.WaitGroup, K int, self int) *W_1toK.End {
	P1 := Proto1.New()
	W := P1.New_W_1toK(K, self)
	err := W.S_1to1_Dial(1, util.LOCALHOST, PORT+self,
			tcp.Dial, new(session2.GobFormatter))
			//shm.Dial, new(session2.PassByPointer))
	if err != nil {
		panic(err)
	}
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1toK.Init) W_1toK.End {
	pay := make([]int, 1)
	var x int
	for {
		w2 := w.S_1to1_Gather_A(pay)

		/*
		select {
		case w3 := <-w2.S_1to1_Recv_B(&x):
			fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received B:", x)
			w = w3.S_1to1_Scatter_C(pay)
		case w4 := <-w2.S_1to1_Recv_D(&x):
			fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received D:", x)
			end := w4.S_1to1_Scatter_E(pay)
			return *end
		}
		/*/
		switch c := w2.S_1to1_Branch().(type) {
		case *W_1toK.B: 
			w3 := c.Recv_B(&x)
			fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received B:", x)
			w = w3.S_1to1_Scatter_C(pay)
		case *W_1toK.D: 
			w4 := c.Recv_D(&x)
			fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received D:", x)
			end := w4.S_1to1_Scatter_E(pay) 
			return *end
		}
		//*/
	}
}
