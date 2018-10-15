//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/game/game01
//$ bin/game01.exe

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
	"github.com/rhu1/scribble-go-runtime/test/game/game01/Game1/Game"
	A "github.com/rhu1/scribble-go-runtime/test/game/game01/Game1/Game/A_1to1"
	B "github.com/rhu1/scribble-go-runtime/test/game/game01/Game1/Game/B_1to1"
	C "github.com/rhu1/scribble-go-runtime/test/game/game01/Game1/Game/C_1to1"

	"github.com/rhu1/scribble-go-runtime/test/game/game01/Game1/Proto1"
	P_1K "github.com/rhu1/scribble-go-runtime/test/game/game01/Game1/Proto1/P_1toK"
	Q    "github.com/rhu1/scribble-go-runtime/test/game/game01/Game1/Proto1/Q_1to1"

	//"github.com/rhu1/scribble-go-runtime/test/util"
)


var _ = gob.Register
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


const PORT_B = 8888
const PORT_C = 9999
const PORT_Q = 7777


/*func init() {
	var foo messages.Foo
	gob.Register(&foo)
}*/



const	K = 3


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//test()
	run()
}

func run() {
	wg := new(sync.WaitGroup)
	wg.Add(2*K + 1 + K)

	for id := 10; id <= 10*K; id = id+10 {
		go server_B(wg, id)
	}

	for id := 10; id <= 10*K; id = id+10 {
		go server_C(wg, id)
	}

	time.Sleep(100 * time.Millisecond)

	go server_Q(wg, K)

	time.Sleep(100 * time.Millisecond)

	for j := 1; j <= K; j = j+1 {
		go client_P1K(wg, K, j)
	}

	wg.Wait()
}

func server_Q(wg *sync.WaitGroup, K int) *Q.End {
	P1 := Proto1.New()
	Q := P1.New_Q_1to1(K, 1)
	var err error
	as := make([]transport2.ScribListener, K)
	for j := 1; j <= K; j++ {
		as[j-1], err = LISTEN(PORT_Q+j)
		if err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}
	for j := 1; j <= K; j++ {
		err := Q.P_1toK_Accept(j, as[j-1], FORMATTER())
		if err != nil {
			panic(err)
		}
		fmt.Println("Q (" + strconv.Itoa(Q.Self) + ") accepted P (" + strconv.Itoa(j) + ") on", PORT_Q+j)
	}
	end := Q.Run(runQ)
	wg.Done()
	return &end
}

func runQ(s *Q.Init) Q.End {
	As := make([]*A.Init, K)
	for j := 1; j <= K; j = j+1 {
		Game := Game.New()
		A := Game.New_A_1to1(1)
		if err := A.B_1to1_Dial(1, "localhost", PORT_B+(10*j), DIAL, FORMATTER()); err != nil {
			panic(err)
		}
		fmt.Println("Q/A (" + strconv.Itoa(10*j) + ") connected to B(" + strconv.Itoa(10*j) + ") on", PORT_B+(10*j))
		if err := A.C_1to1_Dial(1, "localhost", PORT_C+(10*j), DIAL, FORMATTER()); err != nil {
			panic(err)
		}
		fmt.Println("Q/A (" + strconv.Itoa(10*j) + ") connected to C(" + strconv.Itoa(10*j) + ") on", PORT_C+(10*j))
		As[j-1] = A.Init()
	}
	end := s.P_1toK_Scatter_Play(As)
	return *end;
}

func client_P1K(wg *sync.WaitGroup, K int, self int) *P_1K.End {
	P1 := Proto1.New()
	P := P1.New_P_1toK(K, self)
	if err := P.Q_1to1_Dial(1, "localhost", PORT_Q+self, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("P (" + strconv.Itoa(P.Self) + ") connected to Q(1) on", PORT_Q+self)
	end := P.Run(runP)
	wg.Done()
	return &end
}

func runP(s *P_1K.Init) P_1K.End {
	As := make([]*A.Init, 1)
	end := s.Q_1_Gather_Play(As);
	runA(As[0])	
	fmt.Println("P (" + strconv.Itoa(s.Ept.Self) + ") done")
	return *end
}

func server_B(wg *sync.WaitGroup, id int) *B.End {
	Game := Game.New()
	B := Game.New_B_1to1(1)
	var err error
	var ss transport2.ScribListener
	if ss, err = LISTEN(PORT_B+id); err != nil {
		panic(err)
	}
	defer ss.Close()
	if err = B.A_1to1_Accept(1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("B (" + strconv.Itoa(id) + ") accepted A (" + strconv.Itoa(id) + ") on", PORT_B+id)
	end := B.Run(runB)
	wg.Done()
	return &end
}

func runB(s *B.Init) B.End {
	var end *B.End
	switch c := s.A_1_Branch().(type) {
	case *B.Foo: 
		return runB(c.Recv_Foo())
		fmt.Println("B(" + strconv.Itoa(c.Ept.Self) + ") received Foo:")
	case *B.Bar: 
		end = c.Recv_Bar()
		fmt.Println("B(" + strconv.Itoa(c.Ept.Self) + ") received Bar:")
	}
	return *end
}

func server_C(wg *sync.WaitGroup, id int) *C.End {
	Game := Game.New()
	C := Game.New_C_1to1(1)
	var err error
	var ss transport2.ScribListener
	if ss, err = LISTEN(PORT_C+id); err != nil {
		panic(err)
	}
	defer ss.Close()
	if err = C.A_1to1_Accept(1, ss, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("C (" + strconv.Itoa(id) + ") accepted A (" + strconv.Itoa(id) + ") on", PORT_C+id)
	end := C.Run(runC)
	wg.Done()
	return &end
}

func runC(s *C.Init) C.End {
	var end *C.End
	switch c := s.A_1_Branch().(type) {
	case *C.Foo_C_Init: 
		return runC(c.Recv_Foo())
		fmt.Println("C(" + strconv.Itoa(s.Ept.Self) + ") received Foo:")
	case *C.Bar_C_Init: 
		end = c.Recv_Bar()
		fmt.Println("C(" + strconv.Itoa(s.Ept.Self) + ") received Bar:")
	}
	return *end
}

func client_A(wg *sync.WaitGroup, id int) *A.End {
	Game := Game.New()
	A := Game.New_A_1to1(1)
	if err := A.B_1to1_Dial(1, "localhost", PORT_B+id, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("A (" + strconv.Itoa(id) + ") connected to B(" + strconv.Itoa(id) + ") on", PORT_B+id)
	if err := A.C_1to1_Dial(1, "localhost", PORT_C+id, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("A (" + strconv.Itoa(id) + ") connected to B(" + strconv.Itoa(id) + ") on", PORT_C+id)
	end := A.Run(runA)
	wg.Done()
	return &end
}

func runA(s *A.Init) A.End {
	end := s.B_1_Scatter_Bar().C_1_Scatter_Bar()
	fmt.Println("A (" + strconv.Itoa(s.Ept.Self) + ") done")
	return *end
}

func test() {
	wg := new(sync.WaitGroup)
	wg.Add(3*K)

	for id := 10; id <= 10*K; id = id+10 {
		go server_B(wg, id)
	}

	for id := 10; id <= 10*K; id = id+10 {
		go server_C(wg, id)
	}

	time.Sleep(100 * time.Millisecond)

	for id := 10; id <= 10*K; id = id+10 {
		go client_A(wg, id)
	}

	wg.Wait()
}
