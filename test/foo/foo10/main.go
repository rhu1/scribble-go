//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo10
//$ bin/foo10.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo10/Foo10/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go serverCode(wg)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	go clientCode(wg)

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup) *Proto1.Proto1_S_1To1_End {
	P1 := Proto1.NewProto1()

	S := P1.NewProto1_S_1To1(1)
	conn := tcp.NewAcceptor(strconv.Itoa(PORT))
	/*if err != nil {
		log.Fatalf("failed to create connection to W %d: %v", i, err)
	}*/
	S.Accept(P1.W, 1, conn)
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	end = s1.Split_W_1To1_a(1234, util.Copy)
	fmt.Println("S sent:", 1234)

	wg.Done()
	return end
}

func clientCode(wg *sync.WaitGroup) *Proto1.Proto1_W_1To1_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1To1(1)
	conn := tcp.NewConnection(util.LOCALHOST, strconv.Itoa(PORT))
	W.Request(P1.S, 1, conn)
	/*if err != nil {
		log.Fatalf("failed to create connection to Auctioneer: %v", err)
	}*/
	var w1 *Proto1.Proto1_W_1To1_1 = W.Init()

	var x int
	end := w1.Reduce_S_1To1_a(&x, util.Sum)
	fmt.Println("W received:", x)

	wg.Done()
	return end
}
