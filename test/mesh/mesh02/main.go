//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/mesh/mesh02
//$ bin/mesh02.exe

package main

import (
	"encoding/gob"
	"fmt"
	"log"

	//"math/rand"
	//"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"
	"github.com/rhu1/scribble-go-runtime/runtime/twodim/session2"

	"github.com/rhu1/scribble-go-runtime/test/mesh/mesh02/Mesh2/Proto3"
	L "github.com/rhu1/scribble-go-runtime/test/mesh/mesh02/Mesh2/Proto3/family_1/W_l1r1toK1wsubl0r1_not_l1r2toK1w"
	M "github.com/rhu1/scribble-go-runtime/test/mesh/mesh02/Mesh2/Proto3/family_1/W_l1r1toK1wsubl0r1andl1r2toK1w"
	R "github.com/rhu1/scribble-go-runtime/test/mesh/mesh02/Mesh2/Proto3/family_1/W_l1r2toK1w_not_l1r1toK1wsubl0r1"
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

	h := 4;
	w := 3;
	K1w := session2.XY(h, w)

	wg := new(sync.WaitGroup)
	wg.Add(K1w.Flatten(K1w))

	for i := 1; i <= h; i = i+1 {
		go server_R(wg, K1w, session2.XY(i, w))
	}

	for i := 1; i <= h; i = i+1 {
		for j := 2; j < w; j = j+1 {
			go server_M(wg, K1w, session2.XY(i, j))
		}
	}

	time.Sleep(100 * time.Millisecond) //2017/12/11 11:21:40 cannot connect to 127.0.0.1:8891: dial tcp 127.0.0.1:8891: connectex: No connection could be made because the target machine actively refused it.

	for i := 1; i <= h; i = i+1 {
		go client_L(wg, K1w, session2.XY(i, 1))
	}

	wg.Wait()
}


// self.Y == K1w.X
func server_R(wg *sync.WaitGroup, K1w session2.Pair, self session2.Pair) *R.End {
	var err error
	var ss transport2.ScribListener
	P3 := Proto3.New()
	R := P3.New_family_1_W_l1r2toK1w_not_l1r1toK1wsubl0r1(K1w, self)
	if ss, err = LISTEN(PORT+self.Flatten(K1w)); err != nil {
		panic(err)
	}
	defer ss.Close()
	// Accept from below
	if err = R.W_l1r1toK1wsubl0r1andl1r2toK1w_Accept(self.Sub(session2.XY(0, 1)), ss, FORMATTER()); err != nil {
		panic(err)
	}
	//fmt.Println("R ready to run")
	end := R.Run(runR)
	wg.Done()
	return &end
}

func runR(s *R.Init) R.End {
	pay := make([]string, 1)
	end := s.W_selfplusl0rneg1_Gather_Foo(pay)
	fmt.Println("R("+s.Ept.Self.String()+") gathered Foo:", pay)
	return *end
}


/*
var seed = rand.NewSource(time.Now().UnixNano())
var rnd = rand.New(seed)
//var count = 1
*/


// self.X < Kwh.X
func server_M(wg *sync.WaitGroup, K1w session2.Pair, self session2.Pair) *M.End {
	var err error
	var ss transport2.ScribListener
	P3 := Proto3.New()
	M := P3.New_family_1_W_l1r1toK1wsubl0r1andl1r2toK1w(K1w, self)
	if ss, err = LISTEN(PORT+self.Flatten(K1w)); err != nil {
		panic(err)
	}
	defer ss.Close()
	// Accept from below
	if (self.Y == 2) {
		if err = M.W_l1r1toK1wsubl0r1_not_l1r2toK1w_Accept(session2.XY(self.X, 1), ss, FORMATTER()); err != nil {
			panic(err)
		}
	} else {
		if err = M.W_l1r1toK1wsubl0r1andl1r2toK1w_Accept(self.Sub(session2.XY(0, 1)), ss, FORMATTER()); err != nil {
			panic(err)
		}
	}
	// Dial to above
	if (self.Y == K1w.Y-1) {
		peer := session2.XY(self.X, K1w.Y)
		err := M.W_l1r1toK1wsubl0r1andl1r2toK1w_Dial(peer, util.LOCALHOST, PORT+peer.Flatten(K1w), DIAL, FORMATTER())
		if err != nil {
			panic(err)
		}
	} else {
		peer := self.Plus(session2.XY(0, 1))
		err := M.W_l1r2toK1w_not_l1r1toK1wsubl0r1_Dial(peer, util.LOCALHOST, PORT+peer.Flatten(K1w), DIAL, FORMATTER())
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
	s2 := s.W_selfplusl0rneg1_Gather_Foo(pay)
	fmt.Println("M("+s.Ept.Self.String()+") gathered Foo:", pay)
	pay = []string{pay[0] + "thenM" + s.Ept.Self.String()}
	end := s2.W_selfplusl0r1_Scatter_Foo(pay)
	fmt.Println("M("+s.Ept.Self.String()+") scattered Foo:", pay)
	return *end
}

// self.Y == 1
func client_L(wg *sync.WaitGroup, K1w session2.Pair, self session2.Pair) *L.End {
	P3 := Proto3.New()
	L := P3.New_family_1_W_l1r1toK1wsubl0r1_not_l1r2toK1w(K1w, self)
	peer := session2.XY(self.X, 2)
	// Dial to above
	if err := L.W_l1r1toK1wsubl0r1andl1r2toK1w_Dial(peer, util.LOCALHOST, PORT+peer.Flatten(K1w), DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	end := L.Run(runL)
	wg.Done()
	return &end
}

func runL(s *L.Init) L.End {
	pay := []string{"L" + s.Ept.Self.String()}
	end := s.W_selfplusl0r1_Scatter_Foo(pay)
	fmt.Println("L("+s.Ept.Self.String()+") scattered Foo:", pay)
	return *end
}
