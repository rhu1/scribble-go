//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo11
//$ bin/foo11.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo11/Foo11/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	n := 3

	wg := new(sync.WaitGroup)
	wg.Add(n+1)

	go serverCode(wg, n)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for i := 1; i <= n; i++ {
		go clientCode(wg, n, i)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, n int) *Proto1.Proto1_S_1To1_End {
	P1 := Proto1.NewProto1()

	S := P1.NewProto1_S_1To1(n, 1)
	for i := 1; i <= n; i++ {
		conn := tcp.NewAcceptor(strconv.Itoa(PORT+i))
		/*if err != nil {
			log.Fatalf("failed to create connection to W %d: %v", i, err)
		}*/
		S.Accept(P1.W, i, conn)
	}
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	var x int
	end = s1.Split_W_1Tok_(1, util.Copy).Reduce_W_1Tok_(&x, util.Sum)
	fmt.Println("S:", x)

	wg.Done()
	return end
}

func clientCode(wg *sync.WaitGroup, n int, self int) *Proto1.Proto1_W_1Tok_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1Tok(n, self)
	conn := tcp.NewConnection(util.LOCALHOST, strconv.Itoa(PORT+self))
	W.Request(P1.S, 1, conn)
	/*if err != nil {
		log.Fatalf("failed to create connection to Auctioneer: %v", err)
	}*/
	var w1 *Proto1.Proto1_W_1Tok_1 = W.Init()

	var x int
	end := w1.Reduce_S_1To1_(&x, util.Sum).Split_S_1To1_(x+1, util.Copy)
	fmt.Println("W:", x)

	wg.Done()
	return end
}
