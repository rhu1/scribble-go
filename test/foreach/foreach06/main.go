//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foreach/foreach06
//$ bin/foreach06.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/foreach/foreach06/Foreach6/Proto1"
	S_1  "github.com/rhu1/scribble-go-runtime/test/foreach/foreach06/Foreach6/Proto1/family_1/S_1to1"
	W_1  "github.com/rhu1/scribble-go-runtime/test/foreach/foreach06/Foreach6/Proto1/family_1/W_1to1and1toK"
	W_2K "github.com/rhu1/scribble-go-runtime/test/foreach/foreach06/Foreach6/Proto1/family_1/W_1toK_not_1to1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = shm.Dial
var _ = tcp.Dial


/*
var LISTEN = tcp.Listen
var DIAL = tcp.Dial
var FORMATTER = func() *session2.GobFormatter { return new(session2.GobFormatter) } 
/*/
var LISTEN = shm.Listen
var DIAL = shm.Dial
var FORMATTER = func() *session2.PassByPointer { return new(session2.PassByPointer) } 
//*/


const PORT = 8888


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 3

	wg := new(sync.WaitGroup)
	wg.Add(K + 1)

	go serverCode(wg, K)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	go client1Code(wg, K)
	for j := 2; j <= K; j++ {
		go client2KCode(wg, K, j)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, K int) *S_1.End {
	var err error
	P1 := Proto1.New()
	S := P1.New_family_1_S_1to1(K, 1)
	as := make([]transport2.ScribListener, K)
	for j := 1; j <= K; j++ {
		as[j-1], err = LISTEN(PORT+j)
		if err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	if err := S.W_1to1and1toK_Accept(1, as[0], FORMATTER()); err != nil {
		panic(err)
	}
	for j := 2; j <= K; j++ {
		err := S.W_1toK_not_1to1_Accept(j, as[j-1], FORMATTER())
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println("S ready to run")
	end := S.Run(runS)
	wg.Done()
	return &end
}

func runS(s *S_1.Init) S_1.End {
	pay := []int{123}
	s1 := s.W_1_Scatter_A(pay)
	fmt.Println("S scattered A:", pay)
	end := s1.
					//Foreach(nested).
					Parallel(nested).
	         W_1_Gather_C(pay)
	fmt.Println("S gathered C:", pay)
	return *end
}

func nested(s *S_1.Init_7) S_1.End {
	pay := make([]int, 1)
	end := s.W_I_Gather_B(pay)
	fmt.Println("S gathered B:", pay)
	return *end
}

func client1Code(wg *sync.WaitGroup, K int) *W_1.End {
	self := 1
	P1 := Proto1.New()
	W := P1.New_family_1_W_1to1and1toK(K, self)  // Endpoint needs n to check self
	if err := W.S_1to1_Dial(1, util.LOCALHOST, PORT+self, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	//fmt.Println("W(" + strconv.Itoa(W.Self) + ") ready to run")
	end := W.Run(runW1)
	wg.Done()
	return &end
}

func runW1(w *W_1.Init) W_1.End {
	pay := make([]int, 1)
	w2 := w.S_1_Gather_A(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") gathered A:", pay)
	rep := []int{w.Ept.Self}
	w3 := w2.S_1_Scatter_B(rep)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") scattered B:", rep)
	rep = []int{pay[0]*pay[0]}
	end := w3.S_1_Scatter_C(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") scattered C:", rep)
	return *end
}

func client2KCode(wg *sync.WaitGroup, K int, self int) *W_2K.End {
	P1 := Proto1.New()
	W := P1.New_family_1_W_1toK_not_1to1(K, self)  // Endpoint needs n to check self
	if err := W.S_1to1_Dial(1, util.LOCALHOST, PORT+self, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	//fmt.Println("W(" + strconv.Itoa(W.Self) + ") ready to run")
	end := W.Run(runW)
	wg.Done()
	return &end
}

func runW(w *W_2K.Init) W_2K.End {
	pay := []int{w.Ept.Self}
	end := w.S_1_Scatter_B(pay)
	fmt.Println("W(" + strconv.Itoa(w.Ept.Self) + ") scattered B:", pay)
	return *end
}
//*/
