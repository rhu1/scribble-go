//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/fib/fib01
//$ bin/fib01.exe

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

	"github.com/rhu1/scribble-go-runtime/test/fib/Fib/Proto1"
	F1     "github.com/rhu1/scribble-go-runtime/test/fib/Fib/Proto1/family_1/Fib_1toKsub2_not_2toKsub1and3toK"
	F2     "github.com/rhu1/scribble-go-runtime/test/fib/Fib/Proto1/family_1/Fib_1toKsub2and2toKsub1_not_3toK"
	M      "github.com/rhu1/scribble-go-runtime/test/fib/Fib/Proto1/family_1/Fib_1toKsub2and2toKsub1and3toK"
	FKsub1 "github.com/rhu1/scribble-go-runtime/test/fib/Fib/Proto1/family_1/Fib_2toKsub1and3toK_not_1toKsub2"
	FK     "github.com/rhu1/scribble-go-runtime/test/fib/Fib/Proto1/family_1/Fib_3toK_not_1toKsub2and2toKsub1"

	//"github.com/rhu1/scribble-go-runtime/test/util"
)

var _ = fmt.Print
var _ = strconv.AppendBool
var _ = time.After

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


const PORTsub2 = 33333  // For accepting/dialling self-/+2
const PORTsub1 = 44444  // For accepting/dialling self-/+1


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	K := 5

	wg := new(sync.WaitGroup)
	wg.Add(K)

	go serverFK(wg, K)
	go serverclientFKsub1(wg, K)
	for j := 3; j <= K-2; j++ {
		go serverclientM(wg, K, j)
	}
	go clientF2(wg, K)

	time.Sleep(400 * time.Millisecond)

	go clientF1(wg, K)

	wg.Wait()
}


// self == K
func serverFK(wg *sync.WaitGroup, K int) *FK.End {
	P1 := Proto1.New()
	FK := P1.New_family_1_Fib_3toK_not_1toKsub2and2toKsub1(K, K)
	portsub2 := PORTsub2+K
	portsub1 := PORTsub1+K
	var sssub2 transport2.ScribListener
	var sssub1 transport2.ScribListener
	var err error
	if sssub2, err = LISTEN(portsub2); err != nil {
		panic(err)
	}
	defer sssub2.Close()
	if sssub1, err = LISTEN(portsub1); err != nil {
		panic(err)
	}
	defer sssub1.Close()

	if err = FK.Fib_1toKsub2and2toKsub1and3toK_Accept(K-2, sssub2, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("FK (" + strconv.Itoa(FK.Self) + ") accepted", K-2, "on", portsub2)

	if err = FK.Fib_2toKsub1and3toK_not_1toKsub2_Accept(K-1, sssub1, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("FK (" + strconv.Itoa(FK.Self) + ") accepted", K-1, "on", portsub1)

	end := FK.Run(runFK)
	wg.Done()
	return &end
}

func runFK(s *FK.Init) FK.End {
	x := make([]int, 1)
	y := make([]int, 1)
	s2 := s.Fib_selfsub2_Gather_T(x)
	fmt.Println("FK (" + strconv.Itoa(s.Ept.Self) + ") received self-2:", x)
	end := s2.Fib_selfsub1_Gather_T(y)
	fmt.Println("FK (" + strconv.Itoa(s.Ept.Self) + ") received self-1:", y)

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nResult = ", x[0]+y[0])
	return *end
}


// self == K-1
func serverclientFKsub1(wg *sync.WaitGroup, K int) *FKsub1.End {
	self := K-1
	P1 := Proto1.New()
	FKsub1 := P1.New_family_1_Fib_2toKsub1and3toK_not_1toKsub2(K, self)
	aportsub2 := PORTsub2+self
	aportsub1 := PORTsub1+self
	dportplus1 := PORTsub1+K
	var sssub2 transport2.ScribListener
	var sssub1 transport2.ScribListener
	var err error
	if sssub2, err = LISTEN(aportsub2); err != nil {
		panic(err)
	}
	defer sssub2.Close()
	if sssub1, err = LISTEN(aportsub1); err != nil {
		panic(err)
	}
	defer sssub1.Close()

	var acceptsub2 func(int, transport2.ScribListener, session2.ScribMessageFormatter) error 
	if K == 5 {
		acceptsub2 = FKsub1.Fib_1toKsub2and2toKsub1_not_3toK_Accept
	} else {  // K > 5
		acceptsub2 = FKsub1.Fib_1toKsub2and2toKsub1and3toK_Accept	
	}
	if err = acceptsub2(self-2, sssub2, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("FKsub1 (" + strconv.Itoa(FKsub1.Self) + ") accepted", self-2, "on", aportsub2)

	if err = FKsub1.Fib_1toKsub2and2toKsub1and3toK_Accept(self-1, sssub1, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("FKsub1 (" + strconv.Itoa(FKsub1.Self) + ") accepted", self-1, "on", aportsub1)

	if err = FKsub1.Fib_3toK_not_1toKsub2and2toKsub1_Dial(K, "localhost", dportplus1, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("FKsub1 (" + strconv.Itoa(FKsub1.Self) + ") connected", K, "on", dportplus1)

	end := FKsub1.Run(runFKsub1)
	wg.Done()
	return &end
}

func runFKsub1(s *FKsub1.Init) FKsub1.End {
	x := make([]int, 1)
	y := make([]int, 1)
	s2 := s.Fib_selfsub2_Gather_T(x)
	fmt.Println("FKsub1 (" + strconv.Itoa(s.Ept.Self) + ") received self-2:", x)
	s3 := s2.Fib_selfsub1_Gather_T(y)
	fmt.Println("FKsub1 (" + strconv.Itoa(s.Ept.Self) + ") received self-1:", y)
	pay := []int{x[0]+y[0]}
	end := s3.Fib_selfplus1_Scatter_T(pay)
	fmt.Println("FKsub1 (" + strconv.Itoa(s.Ept.Self) + ") sent self+1:", pay)
	return *end
}


// self > 2 && self < K-1
func serverclientM(wg *sync.WaitGroup, K int, self int) *M.End {
	P1 := Proto1.New()
	M := P1.New_family_1_Fib_1toKsub2and2toKsub1and3toK(K, self)
	aportsub2 := PORTsub2+self
	aportsub1 := PORTsub1+self
	dportplus2 := PORTsub2+self+2
	dportplus1 := PORTsub1+self+1
	var sssub2 transport2.ScribListener
	var sssub1 transport2.ScribListener
	var err error
	if sssub2, err = LISTEN(aportsub2); err != nil {
		panic(err)
	}
	defer sssub2.Close()
	if sssub1, err = LISTEN(aportsub1); err != nil {
		panic(err)
	}
	defer sssub1.Close()

	var acceptsub2 func(int, transport2.ScribListener, session2.ScribMessageFormatter) error 
	if self == 3 {
		acceptsub2 = M.Fib_1toKsub2_not_2toKsub1and3toK_Accept	
	} else if self == 4 {
		acceptsub2 = M.Fib_1toKsub2and2toKsub1_not_3toK_Accept	
	} else {  // self >= 5
		acceptsub2 = M.Fib_1toKsub2and2toKsub1and3toK_Accept	
	}
	if err = acceptsub2(self-2, sssub2, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("M (" + strconv.Itoa(M.Self) + ") accepted", self-2, "on", aportsub2)

	var acceptsub1 func(int, transport2.ScribListener, session2.ScribMessageFormatter) error 
	if self == 3 {
		acceptsub1 = M.Fib_1toKsub2and2toKsub1_not_3toK_Accept	
	} else {  // self >= 4
		acceptsub1 = M.Fib_1toKsub2and2toKsub1and3toK_Accept	
	}	
	if err = acceptsub1(self-1, sssub1, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("M (" + strconv.Itoa(M.Self) + ") accepted", self-1, "on", aportsub1)

	var dialplus2 func(int, string, int, func(string, int) (transport2.BinChannel, error), session2.ScribMessageFormatter) error
	if self == K-2 {
		dialplus2 = M.Fib_3toK_not_1toKsub2and2toKsub1_Dial	
	} else if self == K-3 {
		dialplus2 = M.Fib_2toKsub1and3toK_not_1toKsub2_Dial
	} else {
		dialplus2 = M.Fib_1toKsub2and2toKsub1and3toK_Dial	
	}
	if err = dialplus2(self+2, "localhost", dportplus2, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("M (" + strconv.Itoa(M.Self) + ") connected", self+2, "on", dportplus2)

	var dialplus1 func(int, string, int, func(string, int) (transport2.BinChannel, error), session2.ScribMessageFormatter) error
	if self == K-2 {
		dialplus1 = M.Fib_2toKsub1and3toK_not_1toKsub2_Dial
	} else {  // self <= K-2
		dialplus1 = M.Fib_1toKsub2and2toKsub1and3toK_Dial
	}
	if err = dialplus1(self+1, "localhost", dportplus1, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("M (" + strconv.Itoa(M.Self) + ") connected", self+1, "on", dportplus1)

	end := M.Run(runM)
	wg.Done()
	return &end
}

func runM(s *M.Init) M.End {
	x := make([]int, 1)
	y := make([]int, 1)
	s2 := s.Fib_selfsub2_Gather_T(x)
	fmt.Println("M (" + strconv.Itoa(s.Ept.Self) + ") received self-2:", x)
	s3 := s2.Fib_selfsub1_Gather_T(y)
	fmt.Println("M (" + strconv.Itoa(s.Ept.Self) + ") received self-1:", y)
	pay := []int{x[0]+y[0]}
	s4 := s3.Fib_selfplus1_Scatter_T(pay)
	fmt.Println("M (" + strconv.Itoa(s.Ept.Self) + ") sent self+1:", pay)
	end := s4.Fib_selfplus2_Scatter_T(pay)
	fmt.Println("M (" + strconv.Itoa(s.Ept.Self) + ") sent self+2:", pay)
	return *end
}


// self == 2
func clientF2(wg *sync.WaitGroup, K int) *F2.End {
	self := 2
	P1 := Proto1.New()
	F2 := P1.New_family_1_Fib_1toKsub2and2toKsub1_not_3toK(K, self)
	dportplus2 := PORTsub2+self+2
	dportplus1 := PORTsub1+self+1
	var err error

	var dialplus2 func(int, string, int, func(string, int) (transport2.BinChannel, error), session2.ScribMessageFormatter) error
	if K == 5 {
		dialplus2 = F2.Fib_2toKsub1and3toK_not_1toKsub2_Dial
	} else {
		dialplus2 = F2.Fib_1toKsub2and2toKsub1and3toK_Dial
	}
	if err = dialplus2(self+2, "localhost", dportplus2, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("F2 (" + strconv.Itoa(F2.Self) + ") connected", self+2, "on", dportplus2)

	if err = F2.Fib_1toKsub2and2toKsub1and3toK_Dial(self+1, "localhost", dportplus1, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("F2 (" + strconv.Itoa(F2.Self) + ") connected", self+1, "on", dportplus1)

	end := F2.Run(runF2)
	wg.Done()
	return &end
}

func runF2(s *F2.Init) F2.End {
	pay := []int{1}
	s1 := s.Fib_selfplus1_Scatter_T(pay)
	fmt.Println("F2 (" + strconv.Itoa(s.Ept.Self) + ") sent self+1:", pay)
	end := s1.Fib_selfplus2_Scatter_T(pay)
	fmt.Println("F2 (" + strconv.Itoa(s.Ept.Self) + ") sent self+2:", pay)
	return *end
}


// self == 1
func clientF1(wg *sync.WaitGroup, K int) *F1.End {
	self := 1
	P1 := Proto1.New()
	F1 := P1.New_family_1_Fib_1toKsub2_not_2toKsub1and3toK(K, self)
	dportplus2 := PORTsub2+self+2
	var err error

	if err = F1.Fib_1toKsub2and2toKsub1and3toK_Dial(self+2, "localhost", dportplus2, DIAL, FORMATTER()); err != nil {
		panic(err)
	}
	fmt.Println("F1 (" + strconv.Itoa(F1.Self) + ") connected", self+2, "on", dportplus2)

	end := F1.Run(runF1)
	wg.Done()
	return &end
}

func runF1(s *F1.Init) F1.End {
	pay := []int{1}
	end := s.Fib_selfplus2_Scatter_T(pay)
	fmt.Println("F1 (" + strconv.Itoa(s.Ept.Self) + ") sent self+2:", pay)
	return *end
}
