//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/dotapi/dotapi02
//$ bin/dotapi02.exe

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

	"github.com/rhu1/scribble-go-runtime/test/dotapi/dotapi03/messages"
	"github.com/rhu1/scribble-go-runtime/test/dotapi/dotapi03/DotApi3/Proto1"
	Head "github.com/rhu1/scribble-go-runtime/test/dotapi/dotapi03/DotApi3/Proto1/family_1/W_1toKsub1_not_2toK"
	Middle "github.com/rhu1/scribble-go-runtime/test/dotapi/dotapi03/DotApi3/Proto1/family_1/W_1toKsub1and2toK"
	Tail "github.com/rhu1/scribble-go-runtime/test/dotapi/dotapi03/DotApi3/Proto1/family_1/W_2toK_not_1toKsub1"
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

/*func init() {
	var foo messages.Foo
	gob.Register(&foo)
}*/


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K)

	go Server_tail(wg, K, K)

	for j := 2; j < K; j++ {
		go Server_middle(wg, K, j)
	}

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	go Client_head(wg, K, 1)

	wg.Wait()
}

func Server_tail(wg *sync.WaitGroup, K int, self int) *Tail.End {
	P1 := Proto1.New()
	R := P1.New_family_1_W_2toK_not_1toKsub1(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN(PORT+self); err != nil {
		panic(err)
	}
	defer ss.Close()
	if err = R.
			W_1toKsub1and2toK_Accept(self-1, ss, FORMATTER());
			//W_1toKsub1_not_2toK_Accept(self-1, ss, FORMATTER());  // Target variant (L/M) not constrained, but safe to use either
			err != nil {
		panic(err)
	}
	fmt.Println("Tail (" + strconv.Itoa(R.Self) + ") accepted", self-1, "on", PORT+self)
	end := R.Run(runTail)
	wg.Done()
	return &end
}

func runTail(s *Tail.Init) Tail.End {
	/*data := []int { 2, 3, 5, 7, 11, 13 }
	K := s.Ept.K  // Good API? -- generate param values as direct fields? (instead of generic map)
	pay := data[0:K]*/
	pay := make([]messages.Foo, 1)
	end := s.W_selfsub1.Gather.Foo(pay)
	fmt.Println("Tail (" + strconv.Itoa(s.Ept.Self) + ") received:", pay)
	fmt.Println("Tail (" + strconv.Itoa(s.Ept.Self) + ") finished")
	return *end
}

func Server_middle(wg *sync.WaitGroup, K int, self int) *Middle.End {
	P1 := Proto1.New()
	M := P1.New_family_1_W_1toKsub1and2toK(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN(PORT+self); err != nil {
		panic(err)
	}
	defer ss.Close()
	if self == 2 {
		if err = M.W_1toKsub1_not_2toK_Accept(self-1, ss, FORMATTER()); err != nil {
			panic(err)
		}
	} else {
		if err = M.W_1toKsub1and2toK_Accept(self-1, ss, FORMATTER()); err != nil {
			panic(err)
		}
	}
	fmt.Println("Middle (" + strconv.Itoa(M.Self) + ") accepted", self-1, "on", PORT+self)
	if self == K - 1 {
		if err := M.W_2toK_not_1toKsub1_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER()); err != nil {
			panic(err)
		}
	} else {
		if err := M.W_1toKsub1and2toK_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER()); err != nil {
			panic(err)
		}
	}
	fmt.Println("Middle (" + strconv.Itoa(M.Self) + ") connected to", self+1, "on", PORT+self+1)
	end := M.Run(runMiddle)
	wg.Done()
	return &end
}

func runMiddle(s *Middle.Init) Middle.End {
	//pay := make([]int, 1)
	pay := make([]messages.Foo, 1)

	s2 := s. W_selfsub1. Gather. Foo(pay)

	fmt.Println("Middle (" + strconv.Itoa(s.Ept.Self) + ") received:", pay)
	//pay = []int{s.Ept.Self}
	pay = []messages.Foo{messages.Foo{s.Ept.Self}}

	end := s2. W_selfplus1. Scatter. Foo(pay)

	fmt.Println("Middle (" + strconv.Itoa(s.Ept.Self) + ") sent Foo:", pay)
	fmt.Println("Middle (" + strconv.Itoa(s.Ept.Self) + ") finished")
	return *end
}

func Client_head(wg *sync.WaitGroup, K int, self int) *Head.End {
	P1 := Proto1.New()
	L := P1.New_family_1_W_1toKsub1_not_2toK(K, self)  // Endpoint needs n to check self
	if err := L.
			W_1toKsub1and2toK_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER());
			//W_2toK_not_1toKsub1_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER());  // Target variant (M/R) not constrained, but safe to use either
			err != nil {
		panic(err)
	}
	fmt.Println("Head (" + strconv.Itoa(L.Self) + ") connected to", self+1, "on", PORT+self+1)
	end := L.Run(runHead)
	wg.Done()
	return &end
}

func runHead(s *Head.Init) Head.End {
	//pay := []int{s.Ept.Self}
	pay := []messages.Foo{messages.Foo{s.Ept.Self}}
	end := s.W_selfplus1.Scatter.Foo(pay)
	fmt.Println("Head (" + strconv.Itoa(s.Ept.Self) + ") sent Foo:", pay)
	fmt.Println("Head (" + strconv.Itoa(s.Ept.Self) + ") finished")
	return *end
}
















/*---

const PORT = 33333


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K + 1)

	go serverCode(wg, K)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for j := 1; j <= K; j++ {
		go clientCode(wg, K, j)
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

	//*
	end := S.Run(runS)
	/* /
	defer S.Close()
	end := runS(S.Init())
	//* /

	wg.Done()
	return &end
}

func runS(s *S_1.Init) S_1.End {
	end := s.Foreach(nestedS)
	return *end
}

func nestedS(s *S_1.Init_6) S_1.End {
	data := []int { 2, 3, 5, 7, 11, 13 }
	K := s.Ept.K  // Good API? -- generate param values as direct fields? (instead of generic map)
	pay := data[0:K]
	//end := s.W_I_Scatter_A(pay)
	end := s.W_I.Scatter.W_I_Scatter_A(pay)
	fmt.Println("S scattered A:", pay)
	return *end
}

func clientCode(wg *sync.WaitGroup, K int, self int) *W_1toK.End {
	P1 := Proto1.New()
	W := P1.New_W_1toK(K, self)  // Endpoint needs n to check self
	err := W.S_1to1_Dial(1, util.LOCALHOST, PORT+self, DIAL, FORMATTER())
	if err != nil {
		panic(err)
	}
	//fmt.Println("W(" + strconv.Itoa(W.Self) + ") ready to run")

	/*
	end := W.Run(runW)
	/* /
	defer W.Close()
	end := runW(W.Init())
	//* /

	wg.Done()
	return &end
}

func runW(w *W_1toK.Init) W_1toK.End {
	pay := make([]int, 1)
	//end := w.S_1_Gather_A(pay)
	end := w.S_1.Gather.S_1_Gather_A(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered:", pay)
	return *end
}
*/
