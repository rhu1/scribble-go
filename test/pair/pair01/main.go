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


/*
var LISTEN = tcp.Listen
var DIAL = tcp.Dial
var FORMATTER = func() *session2.GobFormatter { return new(session2.GobFormatter) } 
//*/
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

	//K := 3

	wg := new(sync.WaitGroup)
	wg.Add(1 + 1)

	go server_S11(wg)

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	//for i := 1; i <= K; i++ {
		go clientCode(wg)
	//}

	wg.Wait()
}

//for i, j := (session2.Pair{1, 1}), 0; i.Lte(session2.Pair{1, 1}); i, j = i.Inc(session2.Pair{1,1}), j+1 {

func server_S11(wg *sync.WaitGroup) *S11.End {
	var err error
	P1 := Proto1.New()
	self := session2.XY(1, 1)
	S := P1.New_S_l1r1tol1r1(self)
	/*as := make([]transport2.ScribListener, K)
	for j := 1; j <= K; j++ {
		as[j-1], err = LISTEN(PORT+j)
		if err != nil {
			panic(err)
		}
		defer as[j-1].Close()
	}*/
	ss, err := LISTEN(PORT+self.Flatten(session2.XY(1,1)))
	if err != nil {
		panic(err)
	}
	//for j := 1; j <= K; j++ {
		err = S.W_l1r1tol1r1_Accept(session2.XY(1,1), ss, FORMATTER())
		if err != nil {
			panic(err)
		}
	//}
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

func clientCode(wg *sync.WaitGroup) *W11.End {
	P1 := Proto1.New()
	self := session2.XY(1, 1)
	W := P1.New_W_l1r1tol1r1(self)
	err := W.S_l1r1tol1r1_Dial(session2.XY(1,1), util.LOCALHOST, PORT+self.Flatten(session2.XY(1,1)), DIAL, FORMATTER())
	if err != nil {
		panic(err)
	}
	end := W.Run(runW)
	wg.Done()
	return &end
}

func runW(w *W11.Init) W11.End {
	data := make([]int, 1)
	end := w.S_l1r1_Gather_Foo(data)  // FIXME: panic: interface conversion: interface {} is int, not *int -- cf. gob.Register in commented init() ?
	fmt.Println("W(" + w.Ept.Self.Tostring() + ") gathered:", data)
	return *end
}
