//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo1
//$ bin/foo1.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/test/foo1/Foo1/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)


const PORT = 8888


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	n := 3

	wg := new(sync.WaitGroup)
	wg.Add(n+1)

	go serverCode(wg, n)

	time.Sleep(100 * time.Millisecond)  //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for i := 1; i <= n; i++ {
		go clientCode(wg, n, i)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, n int) *Proto1.Proto1_S_1To1_End {
	P1 := Proto1.NewProto1()

	S := P1.NewProto1_S_1To1(n, 1)
	for i := 1; i <= n; i++ {
		S.Accept(P1.W, i, util.LOCALHOST, strconv.Itoa(PORT+i))
	}
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	end = s1.Send_W_1Ton_a(1234, util.Copy)
	fmt.Println("S sent:", 1234)

	wg.Done()
	return end
}

func clientCode(wg *sync.WaitGroup, n int, self int) *Proto1.Proto1_W_1Ton_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1Ton(n, self)
	W.Connect(P1.S, 1, util.LOCALHOST, strconv.Itoa(PORT+self))
	var w1 *Proto1.Proto1_W_1Ton_1 = W.Init()

	var x int
	end := w1.Reduce_S_1To1_a(&x, util.Sum)
	fmt.Println("W received:", self, x)

	wg.Done()
	return end
}
