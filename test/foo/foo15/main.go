//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo15
//$ bin/foo15.exe

package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo15/Foo15/Proto1"
	S "github.com/rhu1/scribble-go-runtime/test/foo/foo15/Foo15/Proto1/S_1to1"
	W1K "github.com/rhu1/scribble-go-runtime/test/foo/foo15/Foo15/Proto1/W_1toK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = shm.Dial
var _ = tcp.Dial


//*
var LISTEN = tcp.Listen
var DIAL = tcp.Dial
var FORMATTER = func() *session2.GobFormatter { return new(session2.GobFormatter) } 
/*/
var LISTEN = shm.Listen
var DIAL = shm.Dial
var FORMATTER = func() *session2.PassByPointer { return new(session2.PassByPointer) } 
//*/


const PORT = 33333


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K + 1)

	go serverCode(wg, K)

	time.Sleep(100 * time.Millisecond)

	for j := 1; j <= K; j++ {
		go clientCode(wg, K, j)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, K int) *S.End {
	var err error
	P1 := Proto1.New()
	S := P1.New_S_1to1(K, 1)
	as := make([]transport2.ScribListener, K)
	for j := 1; j <= K; j++ {
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
	fmt.Println("S ready to run")
	end := S.Run(runS)
	wg.Done()
	return &end
}

var seed = rand.NewSource(time.Now().UnixNano())
var rnd = rand.New(seed)

func runS(s *S.Init) S.End {
	var end *S.End
	if rnd.Intn(2) < 1 {
		end = s.W_1toK_Scatter_Foo()
		fmt.Println("S scattered Foo:")
	} else{
		data := []int{ 2, 3, 5, 7, 11, 13, 17, 19, 23 }
		pay := data[0:s.Ept.K]
		end = s.W_1toK_Scatter_Bar(pay)
		fmt.Println("S scattered Bar:")
	}
	return *end
}

func clientCode(wg *sync.WaitGroup, K int, self int) *W1K.End {
	P1 := Proto1.New()
	W := P1.New_W_1toK(K, self)  // Endpoint needs n to check self
	err := W.S_1to1_Dial(1, util.LOCALHOST, PORT+self, DIAL, FORMATTER())
	if err != nil {
		panic(err)
	}
	fmt.Println("W(" + strconv.Itoa(W.Self) + ") ready to run")
	end := W.Run(runW)
	wg.Done()
	return &end
}

func runW(w *W1K.Init) W1K.End {
	var end *W1K.End
	switch c := w.S_1_Branch().(type) {
	case *W1K.Foo:
		end = c.Recv_Foo()
	  fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received Foo:")
	case *W1K.Bar:
		var x int
    end = c.Recv_Bar(&x)
		fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") received  Bar:", x)
	}
	return *end
}
