//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/pair/pair06
//$ bin/pair06.exe

//go:generate scribblec-param.sh Pair6.scr -d . -param Proto1 github.com/rhu1/scribble-go-runtime/test/pair/pair06/Pair6 -param-api S -param-api W

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

	"github.com/rhu1/scribble-go-runtime/test/pair/pair06/Pair6/Proto1"
	T "github.com/rhu1/scribble-go-runtime/test/pair/pair06/Pair6/Proto1/family_1/W_l1r1plusl1r0toK_not_l1r1toKsubl1r0"
	M "github.com/rhu1/scribble-go-runtime/test/pair/pair06/Pair6/Proto1/family_1/W_l1r1plusl1r0toKandl1r1toKsubl1r0"
	B "github.com/rhu1/scribble-go-runtime/test/pair/pair06/Pair6/Proto1/family_1/W_l1r1toKsubl1r0_not_l1r1plusl1r0toK"
	"github.com/rhu1/scribble-go-runtime/test/util"
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


const PORT = 8888


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	x := 4;
	y := 2;
	K := session2.XY(x, y)

	wg := new(sync.WaitGroup)
	wg.Add(K.Flatten(K))

	for j := 1; j <= y; j = j+1 {
		go server_T(wg, K, session2.XY(x, j))
	}

	for i := 2; i < x; i = i+1 {
		for j := 1; j <= y; j = j+1 {
			go server_M(wg, K, session2.XY(i, j))
		}
	}

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for j := 1; j <= y; j = j+1 {
		go client_B(wg, K, session2.XY(1, j))
	}

	wg.Wait()
}


// self.X == K.X
func server_T(wg *sync.WaitGroup, K session2.Pair, self session2.Pair) *T.End {
	var err error
	var ss transport2.ScribListener
	P1 := Proto1.New()
	T := P1.New_family_1_W_l1r1plusl1r0toK_not_l1r1toKsubl1r0(K, self)
	if ss, err = LISTEN(PORT+self.Flatten(K)); err != nil {
		panic(err)
	}
	defer ss.Close()
	// Accept from below
	if err = T.W_l1r1plusl1r0toKandl1r1toKsubl1r0_Accept(self.Sub(session2.XY(1, 0)), ss, FORMATTER()); err != nil {
		panic(err)
	}
	//fmt.Println("T ready to run")
	end := T.Run(runT)
	wg.Done()
	return &end
}

func runT(s *T.Init) T.End {
	pay := make([]string, 1)
	end := s.W_selfpluslneg1r0_Gather_Foo(pay)
	fmt.Println("T("+s.Ept.Self.String()+") gathered Foo:", pay)
	return *end
}

/*
var seed = rand.NewSource(time.Now().UnixNano())
var rnd = rand.New(seed)
//var count = 1
*/

// self.X < K.X
func server_M(wg *sync.WaitGroup, K session2.Pair, self session2.Pair) *M.End {
	var err error
	var ss transport2.ScribListener
	P1 := Proto1.New()
	M := P1.New_family_1_W_l1r1plusl1r0toKandl1r1toKsubl1r0(K, self)
	if ss, err = LISTEN(PORT + self.Flatten(K)); err != nil {
		panic(err)
	}
	defer ss.Close()
	// Accept from below
	if self.X == 2 {
		if err = M.W_l1r1toKsubl1r0_not_l1r1plusl1r0toK_Accept(session2.XY(1, self.Y), ss, FORMATTER()); err != nil {
			panic(err)
		}
	} else {
		if err = M.W_l1r1plusl1r0toKandl1r1toKsubl1r0_Accept(self.Sub(session2.XY(1, 0)), ss, FORMATTER()); err != nil {
			panic(err)
		}
	}
	// Dial to above
	if (self.X == K.X-1) {
		peer := session2.XY(K.X, self.Y)
		err := M.W_l1r1plusl1r0toKandl1r1toKsubl1r0_Dial(peer, util.LOCALHOST, PORT+peer.Flatten(K), DIAL, FORMATTER())
		if err != nil {
			panic(err)
		}
	} else {
		peer := self.Plus(session2.XY(1, 0))
		err := M.W_l1r1plusl1r0toK_not_l1r1toKsubl1r0_Dial(peer, util.LOCALHOST, PORT+peer.Flatten(K), DIAL, FORMATTER())
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println("M ready to run")
	end := M.Run(runM)
	wg.Done()
	return &end
}

func runM(s *M.Init) M.End {
	pay := make([]string, 1)
	s2 := s.W_selfpluslneg1r0_Gather_Foo(pay)
	fmt.Println("M("+s.Ept.Self.String()+") gathered Foo:", pay)
	pay = []string{pay[0] + "thenM" + s.Ept.Self.String()}
	end := s2.W_selfplusl1r0_Scatter_Foo(pay)
	fmt.Println("M("+s.Ept.Self.String()+") scattered Foo:", pay)
	return *end
}

// self.X == 1
func client_B(wg *sync.WaitGroup, K session2.Pair, self session2.Pair) *B.End {
	P1 := Proto1.New()
	B := P1.New_family_1_W_l1r1toKsubl1r0_not_l1r1plusl1r0toK(K, self)
	peer := session2.XY(2, self.Y)
	// Dial to above
	if err := B.W_l1r1plusl1r0toKandl1r1toKsubl1r0_Dial(peer, util.LOCALHOST, PORT+peer.Flatten(K), DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	end := B.Run(runB)
	wg.Done()
	return &end
}

func runB(s *B.Init) B.End {
	pay := []string{"B" + s.Ept.Self.String()}
	end := s.W_selfplusl1r0_Scatter_Foo(pay)
	fmt.Println("B("+s.Ept.Self.String()+") scattered Foo:", pay)
	return *end
}
