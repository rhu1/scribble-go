//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foreach/foreach07
//$ bin/foreach07.exe

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

	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach07/messages"
	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach07/Foreach7/Proto1"
	S_1 "github.com/rhu1/scribble-go-runtime/test/foreach/foreach07/Foreach7/Proto1/S_1to1"
	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach07/Foreach7/Proto1/W_1toKsub1"
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


const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K-1 + 1)

	go serverCode(wg, K)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for j := 1; j <= K-1; j++ {
		go clientCode(wg, K, j)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, K int) *S_1.End {
	var err error
	P1 := Proto1.New()
	S := P1.New_S_1to1(K, 1)
	as := make([]transport2.ScribListener, K-1)
	for j := 1; j <= K-1; j++ {
		as[j-1], err = LISTEN(PORT+j)
		if err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	for j := 1; j <= K-1; j++ {
		err := S.W_1toKsub1_Accept(j, as[j-1], FORMATTER())
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println("S ready to run")
	end := S.Run(runS)
	wg.Done()
	return &end
}

func runS(s *S_1.Init) S_1.End {
	return *s.Foreach(nested)
}

func nested(s *S_1.Init_11) S_1.End {
	data := []messages.Foo { messages.Foo{2}, messages.Foo{3}, messages.Foo{5}, messages.Foo{7}, messages.Foo{11}, messages.Foo{13} }
	K := s.Ept.K  // Good API? -- generate param values as direct fields? (instead of generic map)
	pay := data[0:K-1]
	end := s.W_ItoI_Scatter_Foo(pay)
	fmt.Println("S scattered A:", pay)
	return *end
}

func clientCode(wg *sync.WaitGroup, K int, self int) *W_1toKsub1.End {
	P1 := Proto1.New()
	W := P1.New_W_1toKsub1(K, self)  // Endpoint needs n to check self
	err := W.S_1to1_Dial(1, util.LOCALHOST, PORT+self, DIAL, FORMATTER())
	if err != nil {
		panic(err)
	}
	//fmt.Println("W(" + strconv.Itoa(W.Self) + ") ready to run")
	end := W.Run(runW)
	wg.Done()
	return &end
}

func runW(w *W_1toKsub1.Init) W_1toKsub1.End {
	pay := make([]messages.Foo, 1)
	end := w.S_1to1_Gather_Foo(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered:", pay)
	return *end
}
