//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo09
//$ bin/foo09.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo09/Foo9/Proto1"
	W_1 "github.com/rhu1/scribble-go-runtime/test/foo/foo09/Foo9/Proto1/W_1to1_not_2to2"
	W_2 "github.com/rhu1/scribble-go-runtime/test/foo/foo09/Foo9/Proto1/W_2to2_not_1to1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

// Bypass bloody annoying Go "unused import" errors
var _ = strconv.Itoa
var _ = tcp.NewAcceptor
var _ = shm.NewConnector
var _ = util.Copy

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	//*
	port := strconv.Itoa(PORT)
	acc := tcp.NewAcceptor(port)
	req := tcp.NewRequestor(util.LOCALHOST, port)
	/*/
	acc := shm.NewConnector()  // FIXME: shm acceptor/requestor API
	req := acc
	//*/

	go server(wg, acc)

	time.Sleep(100 * time.Millisecond)

	go client(wg, req)

	wg.Wait()
}

func server(wg *sync.WaitGroup, acc transport.Transport) *W_1.End {
	P1 := Proto1.New()
	W1 := P1.New_W_1to1_not_2to2(1)  // FIXME: internalise constants
	W1.W_2to2_not_1to1_Accept(2, acc)
	end := W1.Run(runW1)
	wg.Done()
	return end
}

func runW1(s *W_1.Init) W_1.End {
	pay := make([]int, 1)
	end := s.W_2to2_Scatter_A([]int{1111}).  // FIXME: unary Send special case
		     W_2to2_Gather_B(pay)
	fmt.Println("W(1) gathered:", pay)
	return *end
}

func client(wg *sync.WaitGroup, req transport.Transport) *W_2.End {
	P1 := Proto1.New()
	W2 := P1.New_W_2to2_not_1to1(2)
	W2.W_1to1_not_2to2_Dial(1, req)
	end := W2.Run(runW2)
	wg.Done()
	return end
}

func runW2(s *W_2.Init) W_2.End {
	pay := make([]int, 1)
	end := s.W_1to1_Gather_A(pay).
             W_1to1_Scatter_B([]int{2222})
	fmt.Println("W(2) gathered:", pay)
	return *end
}
