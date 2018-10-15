//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/hadamard
//$ bin/hadamard.exe

//go:generate scribblec-param.sh Hadamard.scr -d . -param Proto1 github.com/rhu1/scribble-go-runtime/test/hadamard/Hadamard -param-api A -param-api B -param-api C

package main

import (
	"encoding/gob"
	"fmt"
	"log"
	//"math/rand"
	//"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/twodim/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/hadamard/message"
	"github.com/rhu1/scribble-go-runtime/test/hadamard/Hadamard/Proto1"
	A "github.com/rhu1/scribble-go-runtime/test/hadamard/Hadamard/Proto1/A_l1r1toK"
	B "github.com/rhu1/scribble-go-runtime/test/hadamard/Hadamard/Proto1/B_l1r1toK"
	C "github.com/rhu1/scribble-go-runtime/test/hadamard/Hadamard/Proto1/C_l1r1toK"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = gob.Register
var _ = shm.Dial
var _ = tcp.Dial


/*
var LISTEN = tcp.Listen
var DIAL = tcp.Dial
var FORMATTER = func() *session2.GobFormatter { return new(session2.GobFormatter) } 

func init() {
	var tmp message.Val
	gob.Register(&tmp)
}
/*/
var LISTEN = shm.Listen
var DIAL = shm.Dial
var FORMATTER = func() *session2.PassByPointer { return new(session2.PassByPointer) } 
//*/


const PORT_CA = 7777
const PORT_CB = 8888
//const PORT_CC = 9999


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := session2.XY(2, 3)

	wg := new(sync.WaitGroup)
	wg.Add(K.Flatten(K) * 3)

	for j := session2.XY(1, 1); j.Lte(K); j = j.Inc(K) {
		go server_A(wg, K, j)
		go server_B(wg, K, j)
	}
	/*for j := session2.XY(1, 2); j.Lte(K); j = j.Inc(K) {
		go server_C(wg, K, j)
	}*/

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for j := session2.XY(1, 1); j.Lte(K); j = j.Inc(K) {
		go client_C(wg, K, j)
	}

	wg.Wait()
}


func server_A(wg *sync.WaitGroup, K session2.Pair, self session2.Pair) *A.End {
	var ss transport2.ScribListener
	var err error
	P1 := Proto1.New()
	A := P1.New_A_l1r1toK(K, self)
	ss, err = LISTEN(PORT_CA+self.Flatten(K))
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	err = A.C_l1r1toK_Accept(self, ss, FORMATTER())
	if err != nil {
		panic(err)
	}
	//fmt.Println("A ready to run")
	end := A.Run(runA)
	wg.Done()
	return &end
}

func runA(s *A.Init) A.End {
	data := []message.Val{message.Val{"A", s.Ept.Self.Flatten(s.Ept.K)}}
	return *s.C_selfplusl0r0_Scatter_Val(data)
}


func server_B(wg *sync.WaitGroup, K session2.Pair, self session2.Pair) *B.End {
	var ss transport2.ScribListener
	var err error
	P1 := Proto1.New()
	B := P1.New_B_l1r1toK(K, self)
	ss, err = LISTEN(PORT_CB+self.Flatten(K))
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	err = B.C_l1r1toK_Accept(self, ss, FORMATTER())
	if err != nil {
		panic(err)
	}
	//fmt.Println("B ready to run")
	end := B.Run(runB)
	wg.Done()
	return &end
}

func runB(s *B.Init) B.End {
	data := []message.Val{message.Val{"B", s.Ept.Self.Flatten(s.Ept.K)}}
	return *s.C_selfplusl0r0_Scatter_Val(data)
}


/*var seed = rand.NewSource(time.Now().UnixNano())
var rnd = rand.New(seed)
//var count = 1*/

func client_C(wg *sync.WaitGroup, K session2.Pair, self session2.Pair) *C.End {
	P1 := Proto1.New()
	C := P1.New_C_l1r1toK(K, self)
	/*ss, err = LISTEN(PORT_CC+self.Flatten(K))
	if err != nil {
		panic(err)
	}
	defer ss.Close()
	err = C.C_l1r1toK_Accept(self, ss, FORMATTER())  // No dial/accept generated between C's
	if err != nil {
		panic(err)
	}*/
	if err := C.A_l1r1toK_Dial(self, util.LOCALHOST, PORT_CA+self.Flatten(K), DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	if err := C.B_l1r1toK_Dial(self, util.LOCALHOST, PORT_CB+self.Flatten(K), DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	end := C.Run(runC)
	wg.Done()
	return &end
}

func runC(c *C.Init) C.End {
	pay := make([]message.Val, 1)
	c2 := c.A_selfplusl0r0_Gather_Val(pay)
	fmt.Println("C("+c.Ept.Self.String()+") received from A:", pay)
	end := c2.B_selfplusl0r0_Gather_Val(pay)
	fmt.Println("C("+c.Ept.Self.String()+") received from B:", pay)
	return *end
}
