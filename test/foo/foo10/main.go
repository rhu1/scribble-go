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
	S_1 "github.com/rhu1/scribble-go-runtime/test/foo/foo10/Foo10/Proto1/S_1to1"
	W_1 "github.com/rhu1/scribble-go-runtime/test/foo/foo10/Foo10/Proto1/W_1to1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

const PORT = 8888

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
	acc := tcp.NewAcceptor(strconv.Itoa(PORT))
	S.W_1to1_Accept(1, acc)
	end := S.Run(runS)
	wg.Done()
	return end
}

func runS(s *S_1.Init) S_1.End {
	//pay := []string{"abc"}
	pay := [][]string{[]string{"abc", "def"}}
	end := s.W_1to1_Scatter_A(pay)
	fmt.Println("S scattered:", pay)
	return *end
}

func client(wg *sync.WaitGroup) *W_1.End {
	P1 := Proto1.New()
	W := P1.New_W_1to1(1)
	req := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT))
	W.S_1to1_Dial(1, req)
	end := W.Run(runW)
	wg.Done()
	return end
}

func runW(w *W_1.Init) W_1.End {
	//pay := make([]string, 1)	
	pay := make([][]string, 1)	
	end := w.S_1to1_Gather_A(pay)
	fmt.Println("W gathered:", pay)
	return *end
}
