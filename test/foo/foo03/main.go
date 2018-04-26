//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo03
//$ bin/foo03.exe

package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo03/Foo3/Proto1"
	S_1 "github.com/rhu1/scribble-go-runtime/test/foo/foo03/Foo3/Proto1/S_1to1"
	"github.com/rhu1/scribble-go-runtime/test/foo/foo03/Foo3/Proto1/W_1toK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

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
	for j := 1; j <= K; j++ {
		as[j-1], err = tcp.Listen(PORT+j)
		if err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	for j := 1; j <= K; j++ {
		if err = S.W_1toK_Accept(j, as[j-1], new(session2.GobFormatter)); err != nil {
			panic(err)
		}
	}
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1.Init) S_1.End {
	seed := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(seed)

	var end *S_1.End
	data := []int{ 2, 3, 5, 7, 11, 13, 17, 19, 23 }
	if rnd.Intn(2) < 1 {
		pay := data[0:s.Ept.K]
		end = s.W_1toK_Scatter_A(pay)
		fmt.Println("S scattered A:", pay)
	} else {
		pay := data[s.Ept.K:s.Ept.K+s.Ept.K]
		end = s.W_1toK_Scatter_B(pay)
		fmt.Println("S scattered B:", pay)
	}
	return *end
}

func client(wg *sync.WaitGroup, K int, self int) *W_1toK.End {
	P1 := Proto1.New()
	W := P1.New_W_1toK(K, self)
	err := W.S_1to1_Dial(1, util.LOCALHOST, PORT+self, tcp.Dial, new(session2.GobFormatter))
	if err != nil {
		panic(err)
	}
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1toK.Init) W_1toK.End {
	var end *W_1toK.End
	var x int

	/*/
	select {
	case end = <-w.S_1to1_Recv_A(&x):
		fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received A:", x)
	case end = <-w.S_1to1_Recv_B(&x):
		fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received B:", x)
	}
	/*/
	switch c := w.S_1to1_Branch().(type) {
	case *W_1toK.A: 
		end = c.Recv_A(&x)
		fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received A:", x)
	case *W_1toK.B: 
		end = c.Recv_B(&x)
		fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received B:", x)
	}
	//*/

	return *end
}
