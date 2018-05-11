//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foreach/foreach03
//$ bin/foreach03.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach03/Foreach3/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach03/Foreach3/Proto1/S_1toK1"
	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach03/Foreach3/Proto1/W_1toK2"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K1 := 2
	K2 := 3

	wg := new(sync.WaitGroup)
	wg.Add(K1 + K2)

	for j := 1; j <= K1; j++ {
		go serverCode(wg, K1, K2, j)
	}

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for j := 1; j <= K2; j++ {
		go clientCode(wg, K1, K2, j)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, K1 int, K2 int, self int) *S_1toK1.End {
	var err error
	P1 := Proto1.New()
	S := P1.New_S_1toK1(K2, K1, self)  // FIXME: order
	as := make([]*tcp.TcpListener, K2)
	//as := make([]*shm.ShmListener, K)
	for j := 1; j <= K2; j++ {
		as[j-1], err = tcp.Listen(PORT + K2*(self-1) + j)
		//as[j-1], err = shm.Listen(PORT+j)
		if err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	for j := 1; j <= K2; j++ {
		err := S.W_1toK2_Accept(j, as[j-1], 
			new(session2.GobFormatter))
			//new(session2.PassByPointer))
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println("S ready to run")
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1toK1.Init) S_1toK1.End {
	return *s.Foreach(nested)
}

func nested(s *S_1toK1.Init_14) S_1toK1.End {
	data := []int { 2, 3, 5, 7, 11, 13 }
	Self := s.Ept.Self 
	j := s.Ept.K2*(Self-1) + s.Ept.Params["I2"]-1
	pay := data[j:j+1]
	end := s.W_I2toI2_Scatter_A(pay)
	fmt.Println("S(" + strconv.Itoa(Self) + ") scattered A:", pay)
	return *end
}

func clientCode(wg *sync.WaitGroup, K1 int, K2 int, self int) *W_1toK2.End {
	P1 := Proto1.New()
	W := P1.New_W_1toK2(K1, K2, self)  // Endpoint needs n to check self
	for j := 1; j <= K1; j++ {
		err := W.S_1toK1_Dial(j, util.LOCALHOST, PORT + (K2*(j-1) + self),
				tcp.Dial, new(session2.GobFormatter))
				//shm.Dial, new(session2.PassByPointer))
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println("W(" + strconv.Itoa(W.Self) + ") ready to run")
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1toK2.Init) W_1toK2.End {
	return *w.Foreach(gather)
}

func gather(w *W_1toK2.Init_6) W_1toK2.End {
	pay := make([]int, w.Ept.K1)
	end := w.S_I1toI1_Gather_A(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered:", pay)
	return *end
}
