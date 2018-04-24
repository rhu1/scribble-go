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
	"github.com/rhu1/scribble-go-runtime/test/foo/foo01/Foo1/Proto1/W_1ToK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

const PORT = 8888

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

func serverCode(wg *sync.WaitGroup, K int) *S_1To1.End {
	/*conns :=  make([]tcp.ConnCfg, n)
	for i := 0; i < n; i++ {
		conns[i] = tcp.NewConnection("...", strconv.Itoa(PORT+i))
	}*/
	P1 := Proto1.New()
	S := P1.New_S_1To1(K, 1)
	as := make([]tcp.ConnCfg, K)
	for j := 1; j <= K; j++ {
		as[j-1] = tcp.NewAcceptor(strconv.Itoa(PORT+j))
	}
	for j := 1; j <= K; j++ {
		/*err := session.Accept(S, P1.W.Name(), i, conn)
		if err != nil {
			log.Fatalf("failed to create connection to W %d: %v", i, err)
		}*/
		S.W_1ToK_Accept(j, as[j-1])
	}
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1To1.Init) S_1To1.End {
	data := []int { 2, 3, 5, 7, 11, 13 }
	K := s.Ept.K  // Good API? -- generate param values as direct fields? (instead of generic map)
	pay := data[0:K]
	end := s.W_1ToK_Scatter_A(pay)
	fmt.Println("S scattered A:", pay)
	return *end
}

func clientCode(wg *sync.WaitGroup, K int, self int) *W_1ToK.End {
	P1 := Proto1.New()
	W := P1.New_W_1ToK(K, self)  // Endpoint needs n to check self
	req := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+self))
	W.S_1To1_Dial(1, req)
	/*err := session.Connect(W, P1.S.Name(), 1, conn)
	if err != nil {
		log.Fatalf("failed to create connection to Auctioneer: %v", err)
	}*/
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1ToK.Init) W_1ToK.End {
	pay := make([]int, 1)
	end := w.S_1To1_Gather_A(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered:", pay)
	return *end
}
