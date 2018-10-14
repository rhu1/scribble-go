//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/pair/pair01
//$ bin/pair01.exe

//go:generate scribblec-param.sh Pair1.scr -d . -param Proto1 github.com/rhu1/scribble-go-runtime/test/pair/pair01/Pair1 -param-api S -param-api W

package main

import (
	"encoding/gob"
	"fmt"
	"log"
	//"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/twodim/session2"
	//"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"

	"github.com/rhu1/scribble-go-runtime/test/pair/pair01/Pair1/Proto1"
	S11 "github.com/rhu1/scribble-go-runtime/test/pair/pair01/Pair1/Proto1/S_l1r1tol1r1"
	W11 "github.com/rhu1/scribble-go-runtime/test/pair/pair01/Pair1/Proto1/W_l1r1tol1r1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = gob.Register
var _ = shm.Dial
var _ = tcp.Dial


//*
var LISTEN = tcp.Listen
var DIAL = tcp.Dial
var FORMATTER = func() *session2.GobFormatter { return new(session2.GobFormatter) } 
/*/
var LISTEN = shm.Listen
var DIAL = shm.Dial
var FORMATTER = func() *session2.PassByPointer { return new(session2.PassByPointer) } 
//*/


const PORT = 8888

/*
func init() {
	var tmp int
	gob.Register(&tmp)  // Problem is something to do with this? -- panic: gob: registering duplicate names for *int: "int" != "*int" 
}
//*/

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	p11 := session2.XY(1, 1)

	wg := new(sync.WaitGroup)
	wg.Add(1 + 1)

	go server_S11(wg, p11)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	go clientCode(wg, p11)

	wg.Wait()
}

//for i, j := (session2.Pair{1, 1}), 0; i.Lte(session2.Pair{1, 1}); i, j = i.Inc(session2.Pair{1,1}), j+1 {

func server_S11(wg *sync.WaitGroup, p11 session2.Pair) *S11.End {
	var err error
	P1 := Proto1.New()
	self := p11
	S := P1.New_S_l1r1tol1r1(self)
	ss, err := LISTEN(PORT+self.Flatten(p11))
	if err != nil {
		panic(err)
	}
		err = S.W_l1r1tol1r1_Accept(p11, ss, FORMATTER())
		if err != nil {
			panic(err)
		}
	//fmt.Println("S ready to run")
	end := S.Run(runS)
	wg.Done()
	return &end
}


func runS(s *S11.Init) S11.End {
	data := []int{2, 3, 5}
	end := s.W_l1r1_Scatter_Foo(data)
	fmt.Println("S scattered:", data)
	return *end
}

func clientCode(wg *sync.WaitGroup, p11 session2.Pair) *W11.End {
	P1 := Proto1.New()
	self := p11
	W := P1.New_W_l1r1tol1r1(self)
	err := W.S_l1r1tol1r1_Dial(p11, util.LOCALHOST, PORT+self.Flatten(p11), DIAL, FORMATTER())
	if err != nil {
		panic(err)
	}
	end := W.Run(runW)
	wg.Done()
	return &end
}

func runW(w *W11.Init) W11.End {
	data := make([]int, 1)
	end := w.S_l1r1_Gather_Foo(data)
	fmt.Println("W("+w.Ept.Self.String()+") gathered:", data)
	return *end
}
