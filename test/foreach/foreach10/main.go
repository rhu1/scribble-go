//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foreach/foreach10
//$ bin/foreach10.exe

package main

import (
	"encoding/gob"
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

	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach10/messages"
	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach10/Foreach10/Proto1"
	W1 "github.com/rhu1/scribble-go-runtime/test/foreach/foreach10/Foreach10/Proto1/family_2/W_1to1_not_2to2and2toKsub1and3toK"
	//W2 "github.com/rhu1/scribble-go-runtime/test/foreach/foreach10/Foreach10/Proto1/family_1/W_2to2and2toKsub1_not_1to1and3toK"  
	M  "github.com/rhu1/scribble-go-runtime/test/foreach/foreach10/Foreach10/Proto1/family_2/W_2toKsub1and3toK_not_1to1and2to2"
	WK "github.com/rhu1/scribble-go-runtime/test/foreach/foreach10/Foreach10/Proto1/family_2/W_3toK_not_1to1and2to2and2toKsub1"

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

func init() {
	var foo messages.Foo
	gob.Register(&foo)
	var bar messages.Bar
	gob.Register(&bar)
}


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 4  // K >= 3

	wg := new(sync.WaitGroup)
	wg.Add(K)

	go server_WK(wg, K, K)

	for j := 3; j <= K-1; j++ {
		go server_M(wg, K, j)
	}

	//go server_W2(wg, K, 2)
	go server_M(wg, K, 2)

	time.Sleep(100 * time.Millisecond)

	go client_W1(wg, K, 1)

	wg.Wait()
}

// self == K
func server_WK(wg *sync.WaitGroup, K int, self int) *WK.End {
	P1 := Proto1.New()
	WK := P1.New_family_2_W_3toK_not_1to1and2to2and2toKsub1(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN(PORT+self); err != nil {
		panic(err)
	}
	defer ss.Close()
	// family 1: K > 3 -- so must accept from M -- but could also use "interoperably" between families
	if err = WK.W_2toKsub1and3toK_not_1to1and2to2_Accept(self-1, ss, FORMATTER());
			err != nil {
		panic(err)
	}
			// FIXME: W_3toK_not_1to1and2to2and2toKsub1_Accept ??
	fmt.Println("WK (" + strconv.Itoa(WK.Self) + ") accepted", self-1, "on", PORT+self)
	end := WK.Run(runWK)
	wg.Done()
	return &end
}

func runWK(s *WK.Init) WK.End {
	var end *WK.End
	switch c := s.W_selfsub1_Branch().(type) {
	case *WK.Foo: 
		var x messages.Foo
		end = c.Recv_Foo(&x)
		fmt.Println("WK (" + strconv.Itoa(s.Ept.Self) + ") received Foo:", x)
	case *WK.Bar: 
		var x messages.Bar
		end = c.Recv_Bar(&x)
		fmt.Println("WK (" + strconv.Itoa(s.Ept.Self) + ") received Bar:", x)
	}
	return *end
}

// K > 3
func server_M(wg *sync.WaitGroup, K int, self int) *M.End {
	P1 := Proto1.New()
	M := P1.New_family_2_W_2toKsub1and3toK_not_1to1and2to2(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN(PORT+self); err != nil {
		panic(err)
	}
	defer ss.Close()

	/*if self == 3 {
		if err = M.W_2to2and2toKsub1_not_1to1and3toK_Accept(self-1, ss, FORMATTER()); err != nil {  // FIXME: shouldn't have
			panic(err)
		}
	} else {*/
		if err = M.W_2toKsub1and3toK_not_1to1and2to2_Accept(self-1, ss, FORMATTER()); err != nil {
			panic(err)
		}
	//}
	fmt.Println("M (" + strconv.Itoa(M.Self) + ") accepted", self-1, "on", PORT+self)

	if self == K-1 {
		if err := M.W_3toK_not_1to1and2to2and2toKsub1_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER()); err != nil {
			panic(err)
		}
	} else {
		if err := M.W_2toKsub1and3toK_not_1to1and2to2_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER()); err != nil {
			panic(err)
		}
	}
	fmt.Println("M (" + strconv.Itoa(M.Self) + ") connected to", self+1, "on", PORT+self+1)

	end := M.Run(runM)
	wg.Done()
	return &end
}

func runM(s *M.Init) M.End {
	var end *M.End
	switch c := s.W_selfsub1_Branch().(type) {
	case *M.Foo_W_Init:  // CHECKME: case type name vs. serverWK -- "repeat" message labels already supported?
		var x messages.Foo
		s2 := c.Recv_Foo(&x)
		fmt.Println("M (" + strconv.Itoa(s.Ept.Self) + ") received Foo:", x)
		pay := []messages.Foo{messages.Foo{s.Ept.Self}}
		end = s2.W_selfplus1_Scatter_Foo(pay)
		fmt.Println("M (" + strconv.Itoa(s.Ept.Self) + ") scattered Foo:", pay)
	case *M.Bar_W_Init:
		var x messages.Bar
		s3 := c.Recv_Bar(&x)
		fmt.Println("M (" + strconv.Itoa(s.Ept.Self) + ") received Bar:", x)
		pay := []messages.Bar{messages.Bar{strconv.Itoa(s.Ept.Self)}}
		end = s3.W_selfplus1_Scatter_Bar(pay)
		fmt.Println("M (" + strconv.Itoa(s.Ept.Self) + ") scattered Foo:", pay)
	}
	return *end
}

/*
// self == 2
func server_W2(wg *sync.WaitGroup, K int, self int) *W2.End {
	P1 := Proto1.New()
	M := P1.New_family_1_W_2to2and2toKsub1_not_1to1and3toK(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN(PORT+self); err != nil {
		panic(err)
	}
	defer ss.Close()

	if err = M.W_1to1_not_2to2and2toKsub1and3toK_Accept(self-1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("W2 (" + strconv.Itoa(M.Self) + ") accepted", self-1, "on", PORT+self)

	if K == 3 {  // Doesn't really matter which, both OK?
		if err := M.W_3toK_not_1to1and2to2and2toKsub1_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER()); err != nil {
			panic(err)
		}
	} else {
		if err := M.W_2toKsub1and3toK_not_1to1and2to2_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER()); err != nil {
			panic(err)
		}
	}
	fmt.Println("W2 (" + strconv.Itoa(M.Self) + ") connected to", self+1, "on", PORT+self+1)

	end := M.Run(runW2)
	wg.Done()
	return &end
}

func runW2(s *W2.Init) W2.End {
	var end *W2.End
	switch c := s.W_1_Branch().(type) {
	case *W2.Foo_W_Init:
		var x messages.Foo
		s2 := c.Recv_Foo(&x)
		fmt.Println("W2 (" + strconv.Itoa(s.Ept.Self) + ") received Foo:", x)
		pay := []messages.Foo{messages.Foo{s.Ept.Self}}
		end = s2.W_3_Scatter_Foo(pay)
		fmt.Println("W2 (" + strconv.Itoa(s.Ept.Self) + ") scattered Foo:", pay)
	case *W2.Bar_W_Init:
		var x messages.Bar
		s3 := c.Recv_Bar(&x)
		fmt.Println("W2 (" + strconv.Itoa(s.Ept.Self) + ") received Bar:", x)
		pay := []messages.Bar{messages.Bar{strconv.Itoa(s.Ept.Self)}}
		end = s3.W_3_Scatter_Bar(pay)
		fmt.Println("W2 (" + strconv.Itoa(s.Ept.Self) + ") scattered Foo:", pay)
	}
	return *end
}
//*/

// self == 1
func client_W1(wg *sync.WaitGroup, K int, self int) *W1.End {
	P1 := Proto1.New()
	W1 := P1.New_family_2_W_1to1_not_2to2and2toKsub1and3toK(K, self)
	if err := W1.W_2toKsub1and3toK_not_1to1and2to2_Dial(self+1, util.LOCALHOST, PORT+self+1, DIAL, FORMATTER());
			err != nil {
		panic(err)
	}
	fmt.Println("W1 (" + strconv.Itoa(W1.Self) + ") connected to", self+1, "on", PORT+self+1)
	end := W1.Run(runW1)
	wg.Done()
	return &end
}

func runW1(s *W1.Init) W1.End {
	seed := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(seed)
	var end *W1.End
	if rnd.Intn(2) < 1 {
		pay := []messages.Foo{messages.Foo{s.Ept.Self}}
		end = s.W_2_Scatter_Foo(pay)
		fmt.Println("W1 (" + strconv.Itoa(s.Ept.Self) + ") scattered Foo:", pay)
	} else {
		pay := []messages.Bar{messages.Bar{strconv.Itoa(s.Ept.Self)}}
		end = s.W_2_Scatter_Bar(pay)
		fmt.Println("W1 (" + strconv.Itoa(s.Ept.Self) + ") scattered Bar:", pay)
	}
	return *end
}
