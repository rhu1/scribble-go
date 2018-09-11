//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/deleg/deleg01
//$ bin/deleg01.exe

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

	//"github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/chans"
	"github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/messages"
	"github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/Deleg1/Proto1"
	S "github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/Deleg1/Proto1/S_1to1"
	W "github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/Deleg1/Proto1/W_1to1"
	"github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/Deleg1/Proto2"
	A "github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/Deleg1/Proto2/A_1to1"
	B "github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/Deleg1/Proto2/B_1to1"
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


func testProto2() {
	port := 33333
	wgProto2 := new(sync.WaitGroup)
	wgProto2.Add(1+1)
	go serverB(wgProto2, port)
	time.Sleep(100 * time.Millisecond)
	go clientA(wgProto2, port)
	wgProto2.Wait()
}

func serverB(wgProto2 *sync.WaitGroup, port int) *B.End {
	var err error
	P2 := Proto2.New()
	B := P2.New_B_1to1(1)
	ss, err := LISTEN(port)
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	if err := B.A_1to1_Accept(1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	end := B.Run(runB)
	wgProto2.Done()
	return &end
}

func runB(b *B.Init) B.End {
	pay := make([]messages.Bar, 1)
	end := *b.A_1to1_Gather_Bar(pay)
	fmt.Println("B gathered Bar:", pay)
	return end
}

func clientA(wgProto2 *sync.WaitGroup, port int) *A.End {
	P2 := Proto2.New()
	A := P2.New_A_1to1(1)
	if err := A.B_1to1_Dial(1, util.LOCALHOST, port, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	end := A.Run(runA)
	wgProto2.Done()
	return &end
}

func runA(a *A.Init) A.End {
	data := []messages.Bar{messages.Bar{"a"}, messages.Bar{"b"}, messages.Bar{"c"}}
	pay := data[1:2]
	end := *a.B_1to1_Scatter_Bar(pay)
	fmt.Println("A scattered Bar:", pay)
	return end
}


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//testProto2()

	wgProto1 := new(sync.WaitGroup)
	wgProto1.Add(1+1)
	wgProto2 := new(sync.WaitGroup)
	wgProto2.Add(1+1)
	go serverB(wgProto2, 33333)
	go serverS(wgProto1, 8888)
	go clientW(wgProto1, wgProto2, 8888)
	wgProto1.Wait()
	wgProto2.Wait()
}

func serverS(wgProto1 *sync.WaitGroup, port int) *S.End {
	var err error
	P1 := Proto1.New()
	S := P1.New_S_1to1(1)
	ss, err := LISTEN(port)
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	if err := S.W_1to1_Accept(1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	end := S.Run(runS)
	wgProto1.Done()
	return &end
}

func runS(s *S.Init) S.End {
	P2 := Proto2.New()
	epA := P2.New_A_1to1(1)
	if err := epA.B_1to1_Dial(1, util.LOCALHOST, 33333, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	pay := []A.Init{*epA.Init}
	end := s.W_1to1_Scatter_Foo(pay)
	fmt.Println("S delegated Foo@A:")
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
	pay := make([]A.Init, 1)
	end := w.S_1to1_Gather_Foo(pay)
	runA(&pay[0])
	return *end
}
