//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo07
//$ bin/foo07.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo07/Foo7/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/foo/foo07/Foo7/Proto1/S_1To1"
	"github.com/rhu1/scribble-go-runtime/test/foo/foo07/Foo7/Proto1/W_1ToK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

// Bypass bloody annoying Go "unused import" errors
var _ = strconv.Itoa
var _ = tcp.NewAcceptor
var _ = shm.NewConnector

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K + 1)

	go server(wg, K)

	time.Sleep(100 * time.Millisecond)

	for j := 1; j <= K; j++ {
		go client(wg, K, j)
	}

	wg.Wait()
}

func server(wg *sync.WaitGroup, K int) *S_1To1.End {
	P1 := Proto1.New()
	S := P1.New_S_1To1(K, 1)
	as := make([]tcp.ConnCfg, K)
	for j := 1; j <= K; j++ {
		as[j-1] = tcp.NewAcceptor(strconv.Itoa(PORT+j))
	}
	for j := 1; j <= K; j++ {
		S.W_1ToK_Accept(j, as[j-1])
	}
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1To1.Init) S_1To1.End {
	data := [][]byte{ []byte{2, 3}, []byte{5, 7, 11, 13}, []byte{}, []byte{17, 19, 23}, []byte{29} }
	pay := data[0:s.Ept.K]
	end := s.W_1ToK_Scatter_Norman(pay).
	         W_1ToK_Gather_Mother(pay)
	fmt.Println("S gathered:", pay)
	return *end
}

/*func foo(bss [][]byte) []byte {
	bs := make([]byte, len(bss[0]))
	copy(bs, bss[0])
	for i := 1; i < len(bss); i++ {
		tmp := bss[i]
		for	j := 0; j < len(tmp); j++ {
			bs[j] = bs[j] + tmp[j]	
		}
	}
	return bs
}*/

func client(wg *sync.WaitGroup, K int, self int) *W_1ToK.End {
	P1 := Proto1.New()
	W := P1.New_W_1ToK(K, self)
	req := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+self))
	W.S_1To1_Dial(1, req)
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1ToK.Init) W_1ToK.End {
	pay := make([][]byte, 1)
	w2 := w.S_1To1_Gather_Norman(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered:", pay)
	return *w2.S_1To1_Scatter_Mother(pay)	
}
