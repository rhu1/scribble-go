//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foreach/foreach13
//$ bin/foreach13.exe

package main

import (
	"encoding/gob"
	"fmt"
	"log"
	//"math/rand"
	//"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach13/messages"
	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach13/Foreach13/Proto1"
	A "github.com/rhu1/scribble-go-runtime/test/foreach/foreach13/Foreach13/Proto1/A_1toKA"
	B "github.com/rhu1/scribble-go-runtime/test/foreach/foreach13/Foreach13/Proto1/B_1toKB"

	//"github.com/rhu1/scribble-go-runtime/test/util"
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

func init() {
	var foo messages.Foo
	gob.Register(&foo)
}


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	KA := 2
	KB := 3

	wg := new(sync.WaitGroup)
	wg.Add(KA+KB)

	for i := 1; i <= KB; i++ {
		go server_B(wg, KA, KB, i)
	}

	time.Sleep(100 * time.Millisecond)

	for i := 1; i <= KA; i++ {
		go client_A(wg, KA, KB, i)
	}

	wg.Wait()
}

// self = K
func server_B(wg *sync.WaitGroup, KA int, KB int, self int) *B.End {
	P1 := Proto1.New()
	B := P1.New_B_1toKB(KA, KB, self)
	var err error
	as := make([]transport2.ScribListener, KA)
	for j := 1; j <= KA; j++ {
		if as[j-1], err = LISTEN(PORT+(self*KA)+j); err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	for j := 1; j <= KA; j++ {
		if err = B.A_1toKA_Accept(j, as[j-1], FORMATTER()); err != nil {
			panic(err)
		}
		fmt.Println("B (" + strconv.Itoa(B.Self) + ") accepted A (" + strconv.Itoa(j) + ") on", PORT+(self*KA)+j)
	}
	end := B.Run(runB)
	wg.Done()
	return &end
}

func runB(s *B.Init) B.End {
	return *s.Foreach(nestedB)
}

func nestedB(s *B.Init_15) B.End {
	pay := make([]messages.Foo, 1)
	end := s.A_I_Gather_Foo(pay)
	fmt.Println("B (" + strconv.Itoa(s.Ept.Self) + ") gathered:", pay)
	return *end;
}

func client_A(wg *sync.WaitGroup, KA int, KB int, self int) *A.End {
	P1 := Proto1.New()
	A := P1.New_A_1toKA(KA, KB, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN(PORT+self); err != nil {
		panic(err)
	}
	defer ss.Close();
	for j := 1; j <= KB; j++ {
		if err := A.B_1toKB_Dial(j, "localhost", PORT+(KA*j)+self, DIAL, FORMATTER()); err != nil {
			panic(err)
		}
		fmt.Println("A (" + strconv.Itoa(A.Self) + ") connected to B(" + strconv.Itoa(j) + ") on", PORT+(KA*j)+self)
	}
	end := A.Run(runA)
	wg.Done()
	return &end
}

func runA(s *A.Init) A.End {
	return *s.Foreach(nestedA)
}

func nestedA(s *A.Init_6) A.End {
	pay := []messages.Foo{messages.Foo{s.Ept.Self}}
	end := s.B_J_Scatter_Foo(pay)
	fmt.Println("A (" + strconv.Itoa(s.Ept.Self) + ") scattered to B (" + strconv.Itoa(s.Ept.Params["J"]) + ") Foo:", pay)
	return *end
}
