//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo/foo01
//$ bin/foo01.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	//"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo01/Foo1/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/foo/foo01/Foo1/Proto1/S_1To1"
	"github.com/rhu1/scribble-go-runtime/test/foo/foo01/Foo1/Proto1/W_1Ton"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	n := 3

	wg := new(sync.WaitGroup)
	wg.Add(n + 1)

	go serverCode(wg, n)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for i := 1; i <= n; i++ {
		go clientCode(wg, n, i)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, n int) *S_1To1.End {
	/*conns :=  make([]tcp.ConnCfg, n)
	for i := 0; i < n; i++ {
		conns[i] = tcp.NewConnection("...", strconv.Itoa(PORT+i))
	}*/

	P1 := Proto1.New()
	S := P1.New_S_1To1(n, 1)
	as := make([]tcp.ConnCfg, n)
	for i := 1; i <= n; i++ {
		as[i-1] = tcp.NewAcceptor(strconv.Itoa(PORT+i))
	}
	for i := 1; i <= n; i++ {
		/*err := session.Accept(S, P1.W.Name(), i, conn)
		if err != nil {
			log.Fatalf("failed to create connection to W %d: %v", i, err)
		}*/
		S.Accept("W", i, as[i-1])
	}
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1To1.Init) S_1To1.End {
	data := []int { 2, 3, 5, 7, 11, 13 }
	n := s.Ept.Params["n"]  // Good API?
	pay := data[0:n]
	end := s.Scatter_W_1Ton_A(pay)
	fmt.Println("S scattered A:", pay)
	return *end
}

func clientCode(wg *sync.WaitGroup, n int, self int) *W_1Ton.End {
	P1 := Proto1.New()

	W := P1.New_W_1Ton(n, self)
	conn := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+self))
	W.Dial("S", 1, conn)
	/*err := session.Connect(W, P1.S.Name(), 1, conn)
	if err != nil {
		log.Fatalf("failed to create connection to Auctioneer: %v", err)
	}*/
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1Ton.Init) W_1Ton.End {
	data := make([]int, 1)
	end := w.Gather_S_1To1_A(data)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered:", data)
	return *end
}
