//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/sigs/sigs01
//$ bin/sigs01.exe

package main

import (
	"encoding/gob"
	"fmt"
	"log"
	//"strconv"
	"math/rand"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"
	"github.com/rhu1/scribble-go-runtime/test/util"

	"github.com/rhu1/scribble-go-runtime/test/sigs/sigs01/messages"
	"github.com/rhu1/scribble-go-runtime/test/sigs/sigs01/Sigs1/Proto1"
	S_1 "github.com/rhu1/scribble-go-runtime/test/sigs/sigs01/Sigs1/Proto1/S_1to1"
	W_1 "github.com/rhu1/scribble-go-runtime/test/sigs/sigs01/Sigs1/Proto1/W_1to1"
)

const PORT = 8888

func init() {
	gob.Register(&messages.Foo{})
	gob.Register(&messages.Bar{})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go server(wg)

	time.Sleep(100 * time.Millisecond)

	go client(wg)

	wg.Wait()
}

func server(wg *sync.WaitGroup) *S_1.End {
	P1 := Proto1.New()
	S := P1.New_S_1to1(1)
	ss, err := tcp.Listen(PORT)
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	S.W_1to1_Accept(1, ss, new(session2.GobFormatter))
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1.Init) S_1.End {
	seed := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(seed)

	var end *S_1.End
	if rnd.Intn(2) < 1 {
		pay := []messages.Foo { messages.Foo{123} }

		//...FIXME: send should take pointers for interface compatibility with pointer-passing -- no? slice is already a pointer indirection?

		end = s.W_1to1_Scatter_Foo(pay)
		fmt.Println("S scattered:", pay)
	} else {
		pay := []messages.Bar { messages.Bar{"abc"} }
		end = s.W_1to1_Scatter_Bar(pay)
		fmt.Println("S scattered:", pay)
	}
	return *end
}

func client(wg *sync.WaitGroup) *W_1.End {
	P1 := Proto1.New()
	W := P1.New_W_1to1(1)
	err := W.S_1to1_Dial(1, util.LOCALHOST, PORT, tcp.Dial, new(session2.GobFormatter))
	if err != nil {
		panic(err)
	}
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1.Init) W_1.End {
	var end *W_1.End
	switch c := w.S_1to1_Branch().(type) {
	case *W_1.Foo:
		var pay messages.Foo
		end = c.Recv_Foo(&pay)
		fmt.Println("W gathered Foo:", pay)
	case *W_1.Bar:
		var pay messages.Bar
		end = c.Recv_Bar(&pay)
		fmt.Println("W gathered Bar:", pay)
	default:
		panic("Won't get here: ")
	}
	return *end
}
