//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/deleg/deleg04
//$ bin/deleg04.exe

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

	"github.com/rhu1/scribble-go-runtime/test/deleg/deleg04/messages"
/*	"github.com/rhu1/scribble-go-runtime/test/deleg/deleg04/Deleg4/Proto1"
	S "github.com/rhu1/scribble-go-runtime/test/deleg/deleg04/Deleg4/Proto1/S_1to1"
	W "github.com/rhu1/scribble-go-runtime/test/deleg/deleg04/Deleg4/Proto1/W_1to1"*/
	"github.com/rhu1/scribble-go-runtime/test/deleg/deleg04/Deleg4/Proto2"
	A "github.com/rhu1/scribble-go-runtime/test/deleg/deleg04/Deleg4/Proto2/A_1to1"
	B "github.com/rhu1/scribble-go-runtime/test/deleg/deleg04/Deleg4/Proto2/B_1to1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

// Bypass bloody annoying Go "unused import" errors
var _ = strconv.Itoa
var _ = tcp.Dial
var _ = shm.Dial
var _ = transport2.ScribListener.Accept


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

	//*
	testProto2()
	/*/
	wgProto1 := new(sync.WaitGroup)
	wgProto1.Add(1+1)
	wgProto2 := new(sync.WaitGroup)
	wgProto2.Add(K+1)
	for j := 1; j <= K-1; j++ {  // N.B. up to K-1 -- K'th will be made by S and delegated
		go serverB(wgProto2, K, j)
	}
	time.Sleep(100 * time.Millisecond)
	go serverS(wgProto1, 8888, K)
	time.Sleep(100 * time.Millisecond)
	go clientW(wgProto1, wgProto2, 8888)
	time.Sleep(100 * time.Millisecond)
	go clientA(wgProto2, K)
	wgProto1.Wait()
	wgProto2.Wait()
	//*/
}

func serverB(wgProto2 *sync.WaitGroup) *B.End {
	//var err error
	P2 := Proto2.New()
	epB := P2.New_B_1to1(1)
	ss, err := LISTEN(PORT)
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	if err := epB.A_1to1_Accept(1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	/*
	end := epB.Run(runB)
	/*/
	defer epB.Close()
	end := runB(epB.Init())
	//*/
	wgProto2.Done()
	return &end
}

func runB(b *B.Init) B.End {
	pay := make([]messages.Bar, 1)
	b2 := *b.A_1to1_Gather_Bar(pay)
	fmt.Println("B gathered Bar:", pay)
	end := *b2.A_1to1_Gather_Bar(pay)
	fmt.Println("B gathered Bar:", pay)
	return end
}

func clientA(wgProto2 *sync.WaitGroup) *A.End {
	P2 := Proto2.New()
	A := P2.New_A_1to1(1)
	fmt.Println("A requesting connection to B")
	if err := A.B_1to1_Dial(1, util.LOCALHOST,  PORT, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("A connected to B")
	end := A.Run(runA)
	wgProto2.Done()
	return &end
}

func runA(a *A.Init) A.End {
	pay1 := []messages.Bar{messages.Bar{"1"}}
	a2 := *a.B_1to1_Scatter_Bar(pay1)
	fmt.Println("A scattered Bar:", pay1)
	pay2 := []messages.Bar{messages.Bar{"2"}}
	end := *a2.B_1to1_Scatter_Bar(pay2)
	fmt.Println("A scattered Bar:", pay2)
	return end
}

/*func serverS(wgProto1 *sync.WaitGroup, port int, K int) *S.End {
	var err error
	P1 := Proto1.New()
	epS := P1.New_S_1to1(1)
	ss, err := LISTEN(port)
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	if err := epS.W_1to1_Accept(1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	defer epS.Close()
	end := runS(epS.Init(), K)
	wgProto1.Done()
	return &end
}

func runS(s *S.Init, K int) S.End {
	P2 := Proto2.New()
	epB := P2.New_B_1toK(K, K)  // Delegated id hardcoded to K'th
	ss, err := LISTEN(PORT+K)
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	if err := epB.A_1to1_Accept(1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("S/B accepted connection from A")
	//defer epB.Close()  // FIXME
	pay := []*B.Init{epB.Init()}
	end := s.W_1to1_Scatter_Foo(pay)
	fmt.Println("S delegated Foo(Proto1@B[K]):")
	return *end
}

func clientW(wgProto1 *sync.WaitGroup, wgProto2 *sync.WaitGroup, port int) *W.End {
	P1 := Proto1.New()
	W := P1.New_W_1to1(1)
	if err := W.S_1to1_Dial(1, util.LOCALHOST, port, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	end := W.Run(runW)
	wgProto1.Done()
	wgProto2.Done()
	return &end
}

func runW(w *W.Init) W.End {
	pay := make([]*B.Init, 1)
	end := w.S_1to1_Gather_Foo(pay)
	fmt.Println("W received Foo(Proto1@B[K]):")
	runB(pay[0])  // FIXME: Close?
	return *end
}*/

func testProto2() {
	wgProto2 := new(sync.WaitGroup)
	wgProto2.Add(1+1)
	go serverB(wgProto2)
	time.Sleep(100 * time.Millisecond)
	go clientA(wgProto2)
	wgProto2.Wait()
}
