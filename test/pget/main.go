//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foreach/foreach13
//$ bin/foreach13.exe

package main

import (
	"encoding/gob"
	"fmt"
	"log"
	//"math/rand"
	//"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	//"github.com/rhu1/scribble-go-runtime/test/pget/messages"
	"github.com/rhu1/scribble-go-runtime/test/pget/PGet/Foreach"
	F1 "github.com/rhu1/scribble-go-runtime/test/pget/PGet/Foreach/F_1to1and1toK"
	F2K "github.com/rhu1/scribble-go-runtime/test/pget/PGet/Foreach/F_1toK_not_1to1"
	M "github.com/rhu1/scribble-go-runtime/test/pget/PGet/Foreach/family_1/M_1to1"
	S "github.com/rhu1/scribble-go-runtime/test/pget/PGet/Foreach/family_1/S_1to1"

	//"github.com/rhu1/scribble-go-runtime/test/util"
)


var _ = shm.Dial
var _ = tcp.Dial


//*
var LISTEN_FS = tcp.Listen
var DIAL_FS = tcp.Dial
var FORMATTER_FS = func() *session2.GobFormatter { return new(session2.GobFormatter) } 
/*/
var LISTEN_FS = shm.Listen
var DIAL_FS = shm.Dial
var FORMATTER_FS = func() *session2.PassByPointer { return new(session2.PassByPointer) } 
//*/

//*
var LISTEN_MF = tcp.Listen
var DIAL_MF = tcp.Dial
var FORMATTER_MF = func() *session2.GobFormatter { return new(session2.GobFormatter) } 
/*/
var LISTEN_MF = shm.Listen
var DIAL_MF = shm.Dial
var FORMATTER_MF = func() *session2.PassByPointer { return new(session2.PassByPointer) } 
//*/


const PORT_F = 8888
const PORT_S = 9999



func init() {
	var foo messages.Foo
	gob.Register(&foo)
}


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 4

	wg := new(sync.WaitGroup)
	wg.Add(KA+KB)

	go server_S(wg, K, 1)

	go server_F1(wg, K, 1)
	for j := 2; j <= K; j++ {
		go server_F2k(wg, K, j)
	}

	time.Sleep(100 * time.Millisecond)

	go client_M(wg, K, 1)

	wg.Wait()
}

func server_S(wg *sync.WaitGroup, K int, self int) *S.End {
	P1 := Foreach.New()
	S := P1.New_family_1_S_1to1(K, self)
	var err error
	as := make([]transport2.ScribListener, K)
	for j := 1; j <= K; j++ {
		if as[j-1], err = LISTEN_FS(PORT_S+j); err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	if err = S.F_1to1and1toK_Accept(1, as[1], FORMATTER_FS()); err != nil {
		panic(err)
	}
	fmt.Println("S (" + strconv.Itoa(S.Self) + ") accepted F (" + strconv.Itoa(1) + ") on", PORT_S+1)
	for j := 2; j <= K; j++ {
		if err = S.F_1toK_not_1to1_Accept(j, as[j-1], FORMATTER_FS()); err != nil {
			panic(err)
		}
		fmt.Println("S (" + strconv.Itoa(S.Self) + ") accepted F (" + strconv.Itoa(j) + ") on", PORT_S+j)
	}
	end := S.Run(runS)
	wg.Done()
	return &end
}

func runS(s *S.Init) S.End {
	return *s.F_1_Gather_Head().F_1_Scatter_Res()
}

func nestedS(s *S.Init_22) S.End {
	return *s.F_I_Gather_Get().F_I_Scatter_Res()
}

// K > 1
func server_F2K(wg *sync.WaitGroup, K int, self int) *F2K.End {
	P1 := Foreach.New()
	F := P1.New_F_1toK_not_1to1(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN_MF(PORT_F+self); err != nil {
			panic(err)
		}
	defer ss.Close()
	if err = F.M_1to1_Accept(1, ss, FORMATTER_MF()); err != nil {
		panic(err)
	}
	fmt.Println("F (" + strconv.Itoa(F.Self) + ") accepted M (1) on", PORT_F+self)
	if err = F.S_1to1_Dial(1, "localhost", PORT_S+self, DIAL_FS, FORMATTER_FS()); err != nil {
		panic(err)
	}
	fmt.Println("F (" + strconv.Itoa(F.Self) + ") connmected S (1) on", PORT_S+self)
	end := F.Run(runF2K)
	wg.Done()
	return &end
}

func runF2k(s *F2K.Init) F2K.End {
	return *s.M_1_Gather_Job().S_1_Scatter_Get().S_1_Gather_Res().M_1_Scatter_Data()
}

// self = 1
func server_F1(wg *sync.WaitGroup, K int, self int) *F1.End {
	P1 := Foreach.New()
	F := P1.New_F_1to1and1toK(K, self)
	var ss transport2.ScribListener
	var err error
	if ss, err = LISTEN_MF(PORT_F+self); err != nil {
			panic(err)
		}
	defer ss.Close()
	if err = F.M_1to1_Accept(1, ss, FORMATTER_MF()); err != nil {
		panic(err)
	}
	fmt.Println("F (" + strconv.Itoa(F.Self) + ") accepted M (1) on", PORT_F+self)
	if err = F.S_1to1_Dial(1, "localhost", PORT_S+self, DIAL_FS, FORMATTER_FS()); err != nil {
		panic(err)
	}
	fmt.Println("F (" + strconv.Itoa(F.Self) + ") connmected S (1) on", PORT_S+self)
	end := F.Run(runF2K)
	wg.Done()
	return &end
}

func runF1(s *F1.Init) F1.End {
	return *s.S_1_Scatter_Head().S_1_Gather_Res().M_1_Gather_Job().S_1_Scatter_Get().S_1_Gather_Res().M_1_Scatter_Data()
}

func client_M(wg *sync.WaitGroup, K, self int) *M.End {
	P1 := Foreach.New()
	M := P1.New_family_1_M_1to1(K, self)
	var ss transport2.ScribListener
	var err error
	if err := M.F_1to1and1toK_Dial(1, "localhost", PORT_F+1, DIAL_MF, FORMATTER_MF()); err != nil {
		panic(err)
	}
	fmt.Println("M (" + strconv.Itoa(M.Self) + ") connected to F(1) on", PORT_F+1)
	for j := 2; j <= K; j++ {
		if err := M.F_1toK_not_1to1(j, "localhost", PORT+j, DIAL_MF, FORMATTER_MF()); err != nil {
			panic(err)
		}
		fmt.Println("M (" + strconv.Itoa(A.Self) + ") connected to F(" + strconv.Itoa(j) + ") on", PORT+j)
	}
	end := M.Run(runM)
	wg.Done()
	return &end
}

func runM(s *M.Init) M.End {
	s.F_1_Gather_Meta().F_1toK_Scatter_Job()
}

/*func nestedM(s *M.Init_8) M.End {
	pay := []messages.Foo{messages.Foo{s.Ept.Self}}
	end := s.B_J_Scatter_Foo(pay)
	fmt.Println("A (" + strconv.Itoa(s.Ept.Self) + ") scattered to B (" + strconv.Itoa(s.Ept.Params["J"]) + ") Foo:", pay)
	return *end
}*/
