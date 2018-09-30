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
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/messages"
	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1"
	//Left "github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1/family_2/W_1toKsub1_not_2toK"
	Left "github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1/family_1/W_1toKsub1_not_2toK"
	//Right "github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1/family_2/W_2toK_not_1toKsub1"
	Right "github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1/family_1/W_2toK_not_1toKsub1"
	Middle "github.com/rhu1/scribble-go-runtime/test/foreach/foreach08/Foreach8/Proto1/W_1toKsub1and2toK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = shm.Dial
var _ = tcp.Dial


// FIXME: tcp broken -- panic: EOF main.go:126
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

	K := 4

	wg := new(sync.WaitGroup)
	wg.Add(K)

	go server_right(wg, K, K)

	for j := 2; j <= K-1; j++ {
		go server_middle(wg, K, j)
	}

	time.Sleep(100 * time.Millisecond)

	go client_left(wg, K, 1)

	wg.Wait()
}

// self = K
func server_right(wg *sync.WaitGroup, K int, self int) *Right.End {
	P1 := Proto1.New()
	R := P1.New_family_1_W_2toK_not_1toKsub1(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN(PORT+self); err != nil {
		panic(err)
	}
	defer ss.Close()
	if err = R.
			W_1toKsub1_not_2toK_Accept(self-1, ss, FORMATTER());
			//W_1toKsub1_not_2toK_Accept(self-1, ss, FORMATTER());  // Target variant (L/M) not constrained, but safe to use either
			err != nil {
		panic(err)
	}
	fmt.Println("Right (" + strconv.Itoa(R.Self) + ") accepted", self-1, "on", PORT+self)
	end := R.Run(runRight)
	wg.Done()
	return &end
}

func runRight(s *Right.Init) Right.End {
	pay := make([]messages.Foo, 1)
	end := s.W_selfplus1sub2_Gather_Foo(pay)
	fmt.Println("Right (" + strconv.Itoa(s.Ept.Self) + ") gathered:", pay)
	return *end
}

func server_middle(wg *sync.WaitGroup, K int, self int) *Middle.End {
	P1 := Proto1.New()
	M := P1.New_W_1toKsub1and2toK(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN(PORT+self); err != nil {
		panic(err)
	}
	defer ss.Close()
	if err = M.W_1toKsub1_not_2toK_Accept(self-1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("Middle (" + strconv.Itoa(M.Self) + ") accepted", self-1, "on", PORT+self)
	if err := M.W_2toK_not_1toKsub1_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("Middle (" + strconv.Itoa(M.Self) + ") connected to", self+1, "on", PORT+self+1)
	end := M.Run(runMiddle)
	wg.Done()
	return &end
}

func runMiddle(s *Middle.Init) Middle.End {
	pay := make([]messages.Foo, 1)
	s2 := s.W_selfplus1sub2_Gather_Foo(pay)
	fmt.Println("Middle (" + strconv.Itoa(s.Ept.Self) + ") gathered:", pay)
	pay = []messages.Foo{messages.Foo{s.Ept.Self}}
	end := s2.W_selfplus2sub1_Scatter_Foo(pay)
	fmt.Println("Middle (" + strconv.Itoa(s.Ept.Self) + ") scattered Foo:", pay)
	return *end
}

func client_left(wg *sync.WaitGroup, K int, self int) *Left.End {
	P1 := Proto1.New()
	L := P1.New_family_1_W_1toKsub1_not_2toK(K, self)  // Endpoint needs n to check self
	if err := L.
			W_1toKsub1and2toK_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER());
			//W_2toK_not_1toKsub1_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER());  // Target variant (M/R) not constrained, but safe to use either
			err != nil {
		panic(err)
	}
	fmt.Println("Left (" + strconv.Itoa(L.Self) + ") connected to", self+1, "on", PORT+self+1)
	end := L.Run(runL)
	wg.Done()
	return &end
}

func runL(s *Left.Init) Left.End {
	pay := []messages.Foo{messages.Foo{s.Ept.Self}}
	end := s.W_selfplus2sub1_Scatter_Foo(pay)
	fmt.Println("Left (" + strconv.Itoa(s.Ept.Self) + ") scattered Foo:", pay)
	return *end
}
