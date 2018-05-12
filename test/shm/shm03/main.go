//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/shm/shm03
//$ bin/shm03.exe

package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/shm/shm03/messages"
	"github.com/rhu1/scribble-go-runtime/test/shm/shm03/Shm3/Proto1"
	S_1  "github.com/rhu1/scribble-go-runtime/test/shm/shm03/Shm3/Proto1/S_1to1"
	W_1K "github.com/rhu1/scribble-go-runtime/test/shm/shm03/Shm3/Proto1/W_1toK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = shm.Dial
var _ = tcp.Dial

const PORT = 8888

//*
var LISTEN = tcp.Listen
var DIAL = tcp.Dial
var FORMATTER = func() *session2.GobFormatter { return new(session2.GobFormatter) } 
/*/
var LISTEN = shm.Listen
var DIAL = shm.Dial
var FORMATTER = func() *session2.PassByPointer { return new(session2.PassByPointer) } 
//*/

func init() {
	var foo messages.Foo
	gob.Register(&foo)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K + 1)

	go serverCode(wg, K)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for i := 1; i <= K; i++ {
		go clientCode(wg, K, i)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, K int) *S_1.End {
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
	//fmt.Println("S ready to run")
	end := S.Run(runS)
	wg.Done()
	return end
}


func runS(s *S_1.Init) S_1.End {
	data := []messages.Foo{messages.Foo(2), messages.Foo(3), messages.Foo(5)}
	end := s.W_1toK_Scatter_Foo(data)
	fmt.Println("S scattered:", data)
	return *end
}

func clientCode(wg *sync.WaitGroup, K int, self int) *W_1K.End {
	P1 := Proto1.New()
	W := P1.New_W_1toK(K, self)
	err := W.S_1to1_Dial(1, util.LOCALHOST, PORT + self, DIAL, FORMATTER())
	if err != nil {
		panic(err)
	}
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1K.Init) W_1K.End {
	data := make([]messages.Foo, 1)
	end := w.S_1to1_Gather_Foo(data)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered:", data)
	return *end
}
