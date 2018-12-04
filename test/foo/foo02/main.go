//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo02
//$ bin/foo02.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo02/Foo2/Proto1"
	S_1 "github.com/rhu1/scribble-go-runtime/test/foo/foo02/Foo2/Proto1/S_1to1"
	"github.com/rhu1/scribble-go-runtime/test/foo/foo02/Foo2/Proto1/W_1toK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = shm.Dial
var _ = tcp.Dial


/*
var LISTEN = tcp.Listen
var DIAL = tcp.Dial
var FORMATTER = func() *session2.GobFormatter { return new(session2.GobFormatter) } 
/*/
var LISTEN = shm.Listen
var DIAL = shm.Dial
var FORMATTER = func() *session2.PassByPointer { return new(session2.PassByPointer) } 
//*/


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
	P1 := Proto1.New()
	S := P1.New_S_1to1(K, 1)
	as := make([]transport2.ScribListener, K)
	for j := 1; j <= K; j++ {
		var err error
		as[j-1], err = LISTEN(PORT+j)
		if err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	for j := 1; j <= K; j++ {
		err := S.W_1toK_Accept(j, as[j-1], FORMATTER())
		if err != nil {
			panic(err)
		}
	}
	end := S.Run(runS)
	wg.Done()
	return &end
}

func runS(s *S_1.Init) S_1.End {
	data := []int{ 2, 3, 5, 7, 11 }
	pay := data[0:s.Ept.K]
	s2 := s.W_1toK_Scatter_A(pay)
	fmt.Println("S scattered:", pay)
	end := s2.W_1toK_Gather_B(pay)
	fmt.Println("S gathered:", pay)
	return *end
}

func client(wg *sync.WaitGroup, K int, self int) *W_1toK.End {
	P1 := Proto1.New()
	W := P1.New_W_1toK(K, self)
	err := W.S_1to1_Dial(1, util.LOCALHOST, PORT+self, DIAL, FORMATTER())
	if err != nil {
		panic(err)
	}
	end := W.Run(runW)
	wg.Done()
	return &end
}

func runW(w *W_1toK.Init) W_1toK.End {
	pay := make([]int, 1)
	w2 := w.S_1_Gather_A(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered: ", pay)
	end := w2.S_1_Scatter_B(pay[0:1])
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") scattered: ", pay)
	return *end
}
