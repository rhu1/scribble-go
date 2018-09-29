//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foreach/foreach08
//$ bin/foreach08.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	//"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/messages"
	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1"
	Left "github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1/family_2/W_1toKsub1_not_2toK"
	Right "github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1/family_2/W_2toK_not_1toKsub1"
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

	K := 2

	wg := new(sync.WaitGroup)
	wg.Add(K)

	go serverCode(wg, K, 2)

	time.Sleep(100 * time.Millisecond)

	//for j := 1; j <= K; j++ {
		go clientCode(wg, K, 1)
	//}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, K int, self int) *Right.End {
	//var err error
	P1 := Proto1.New()
	R := P1.New_family_2_W_2toK_not_1toKsub1(K, self)
	//as := make([]transport2.ScribListener, K-1)
	//for j := 1; j <= K; j++ {
		ss, err := LISTEN(PORT+self)
		if err != nil {
			panic(err)
		}
		defer ss.Close()
	//}
	//for j := 1; j <= K; j++ {
		err = R.W_1toKsub1_not_2toK_Accept(self-1, ss, FORMATTER())
		if err != nil {
			panic(err)
		}
	//}*/
	//fmt.Println("S ready to run")
	end := R.Run(runS)
	wg.Done()
	return &end
}

func runS(s *Right.Init) Right.End {
	pay := make([]messages.Foo, 1)
	end := s.W_selfplus1sub2_Gather_Foo(pay)
	fmt.Println("W(" + strconv.Itoa(s.Ept.Self) + ") gathered:", pay)
	return *end
}

func clientCode(wg *sync.WaitGroup, K int, self int) *Left.End {
	P1 := Proto1.New()
	L := P1.New_family_2_W_1toKsub1_not_2toK(K, self)  // Endpoint needs n to check self
	err := L.W_2toK_not_1toKsub1_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER())
	if err != nil {
		panic(err)
	}
	//fmt.Println("W(" + strconv.Itoa(W.Self) + ") ready to run")
	end := L.Run(runL)
	wg.Done()
	return &end
}

func runL(s *Left.Init) Left.End {
	pay := []messages.Foo{messages.Foo{s.Ept.Self}}
	end := s.W_selfplus2sub1_Scatter_Foo(pay)
	fmt.Println("S scattered A:", pay)
	return *end
}
