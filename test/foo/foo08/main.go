//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo08
//$ bin/foo08.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo08/Foo8/Proto1"
	S_1    "github.com/rhu1/scribble-go-runtime/test/foo/foo08/Foo8/Proto1/S_1to1"
	W_1    "github.com/rhu1/scribble-go-runtime/test/foo/foo08/Foo8/Proto1/W_1to1and1toK"
	W_2toK "github.com/rhu1/scribble-go-runtime/test/foo/foo08/Foo8/Proto1/W_1toK_not_1to1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

// Bypass bloody annoying Go "unused import" errors
var _ = strconv.Itoa
var _ = tcp.NewAcceptor
var _ = shm.NewConnector

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K + 1)

	go server(wg, K)

	time.Sleep(100 * time.Millisecond)

	go W1(wg, K, 1)
	for j := 2; j <= K; j++ {
		go W2K(wg, K, j)
	}

	wg.Wait()
}

func server(wg *sync.WaitGroup, K int) *S_1.End {
	P1 := Proto1.New()
	S := P1.New_S_1to1(K, 1)
	as := make([]tcp.ConnCfg, K)
	for j := 1; j <= K; j++ {
		as[j-1] = tcp.NewAcceptor(strconv.Itoa(PORT+j))
	}
	S.W_1to1and1toK_Accept(1, as[0])
	for j := 2; j <= K; j++ {
		S.W_1toK_not_1to1_Accept(j, as[j-1])
	}
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1.Init) S_1.End {
	data := []int{ 2, 3, 5, 7, 11, 13, 17, 19, 23, 29 }
	end := s.W_1to1_Scatter_A(data[0:1]).
	         W_1toK_Scatter_B(data[1:s.Ept.K+1])
	return *end
}

func W1(wg *sync.WaitGroup, K int, self int) *W_1.End {
	P1 := Proto1.New()
	W := P1.New_W_1to1and1toK(K, self)
	req := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+self))
	W.S_1to1_Dial(1, req)
	end := W.Run(runW1)
	wg.Done()
	return end
}

func runW1(w *W_1.Init) W_1.End {
	pay := make([]int, 1)
	w2 := w.S_1to1_Gather_A(pay)
	fmt.Println("W(" + strconv.Itoa(1) + ") gathered:", pay)
	end := w2.S_1to1_Gather_B(pay)
	fmt.Println("W(" + strconv.Itoa(1) + ") gathered:", pay)
	return *end
}

func W2K(wg *sync.WaitGroup, K int, self int) *W_2toK.End {
	P1 := Proto1.New()
	W := P1.New_W_1toK_not_1to1(K, self)
	req := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+self))
	W.S_1to1_Dial(1, req)
	end := W.Run(runW2K)
	wg.Done()
	return end
}

func runW2K(w *W_2toK.Init) W_2toK.End {
	pay := make([]int, 1)
	end := w.S_1to1_Gather_B(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered:", pay)
	return *end
}
