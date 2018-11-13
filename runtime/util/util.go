package util

import "bufio"
import "fmt"
import "io/ioutil"
import "os"
import "os/exec"
import "strconv"
import "strings"
import "github.com/rhu1/scribble-go-runtime/runtime/twodim/session2"

var _ = fmt.Errorf

type IntInterval struct {
	Start	int
	End int
}

func (ival IntInterval) String() string {
	return "[" + strconv.Itoa(ival.Start) + ", " + strconv.Itoa(ival.End) + "]";
}

func (ival IntInterval) IsEmpty() bool {
	return ival.Start > ival.End;
}

// Pre len(ivals) > 0
func IsectIntIntervals(ivals []IntInterval) IntInterval {
	if len(ivals) == 1 {
		return ivals[0]
	} else {
		ivals = isectIntInterval(ivals)
		if ivals[0].IsEmpty() {
			return ivals[0]
		}
		return IsectIntIntervals(ivals)
	}
}

// Pre: len(ivals) > 1
// Intersct ivals[0] and ivals[1] into ivals[1] and trim
func isectIntInterval(ivals []IntInterval) []IntInterval {
	//fmt.Print("intersecting: " + ivals[0].String() + " ,, " + ivals[1].String() + " = ")
	if ivals[0].Start > ivals[1].End {
		ivals[1] = IntInterval{ivals[0].Start, ivals[1].End}
	} else if ivals[1].Start > ivals[0].End {
		ivals[1] = IntInterval{ivals[1].Start, ivals[0].End}
	} else {
		start := max(ivals[0].Start, ivals[1].Start)
		end := min(ivals[1].End, ivals[0].End)
		ivals[1] = IntInterval{start, end}
	}
	//fmt.Println(ivals)
	return ivals[1:]
}

func (i IntInterval) SubIntIntervals(ivals []IntInterval) IntInterval {
	if len(ivals) == 0 {
		return i
	}
	return i.subIntInterval(ivals[0]).SubIntIntervals(ivals[1:])
}

func (i1 IntInterval) subIntInterval(i2 IntInterval) IntInterval {
	var s int
	var e int
	if i1.Start < i2.Start {
		s = i1.Start
		if i2.Start <= i1.End {
			if i2.End < i1.End {
				panic("TODO: " + i1.String() + " - " + i2.String())
			} else {  // i1.End <= i2.End
				e = i2.Start - 1
			}
		}	else {  // i1.End < i2.Start
			e = i1.End
		}
	}	else {  // i1.Start >= i2.Start
		if i1.Start <= i2.End {
			if i1.End <= i2.End {
				s = 0	
				e = -1
			} else {  // i2.End < i1.End
				s = i2.End + 1	
				e = i1.End
			}
		} else {  // i2.End < i1.Start
			s = i2.End + 1	
			e = i1.End
		}
	}
	return IntInterval{s, e}
}

func min(x int, y int) int {
	if x < y {
		return	x
	}
	return y
}

func max(x int, y int) int {
	if x > y {
		return	x
	}
	return y
}



// All duplicated for another type, because Go
type IntPairInterval struct {
	Start	session2.Pair
	End session2.Pair
}

func (ival IntPairInterval) String() string {
	return "[" + ival.Start.String() + ", " + ival.End.String() + "]"
}

func (ival IntPairInterval) IsEmpty() bool {
	return ival.Start.Gt(ival.End);
}

func IsectIntPairIntervals(ivals []IntPairInterval) IntPairInterval {
	if len(ivals) == 1 {
		return ivals[0]
	} else {
		ivals = isectIntPairInterval(ivals)
		if ivals[0].IsEmpty() {
			return ivals[0]
		}
		return IsectIntPairIntervals(ivals)
	}
}

// Pre: len(ivals) > 1
// Intersct ivals[0] and ivals[1] into ivals[1] and trim
func isectIntPairInterval(ivals []IntPairInterval) []IntPairInterval {
	if ivals[0].Start.Gt(ivals[1].End) {
		ivals[1] = IntPairInterval{ivals[0].Start, ivals[1].End}
	} else if ivals[1].Start.Gt(ivals[0].End) {
		ivals[1] = IntPairInterval{ivals[1].Start, ivals[0].End}
	} else {
		start := maxIntPair(ivals[0].Start, ivals[1].Start)
		end := minIntPair(ivals[1].End, ivals[0].End)
		ivals[1] = IntPairInterval{start, end}
	}
	return ivals[1:]
}

func minIntPair(x session2.Pair, y session2.Pair) session2.Pair {
	if x.Lt(y) {
		return	x
	}
	return y
}

func maxIntPair(x session2.Pair, y session2.Pair) session2.Pair {
	if x.Gt(y) {
		return	x
	}
	return y
}

// Adapted from Z3Wrapper.checkSat
func CheckSat(//GoJob job, ProtocolDecl<Global> gpd, 
		smt2 string) bool {

	smt2 = "(declare-datatypes (T1 T2) ((Pair (mk-pair (fst T1) (snd T2)))))\n" +
			"(define-fun pair_max ((p!1 (Pair Int Int))) Int (ite (< (fst p!1) (snd p!1)) (snd p!1) (fst p!1) ) )\n" +
			"(define-fun pair_min ((p!1 (Pair Int Int))) Int (ite (< (fst p!1) (snd p!1)) (fst p!1) (snd p!1)))\n" +
			"(define-fun twopair_max ((p!1 (Pair Int Int)) (p!2 (Pair Int Int))) Int (ite (< (pair_max p!1) (pair_max p!2)) (pair_max p!2) (pair_max p!1)))\n" +
			"(define-fun twopair_min ((p!1 (Pair Int Int)) (p!2 (Pair Int Int))) Int (ite (< (pair_min p!1) (pair_min p!2)) (pair_min p!1) (pair_min p!2)))\n" +
			"(define-fun pair_lte ((p!1 (Pair Int Int)) (p!2 (Pair Int Int))) Bool (and (<= (fst p!1) (fst p!2)) (<= (snd p!1) (snd p!2)) ))\n" +
			"(define-fun pair_lt ((p!1 (Pair Int Int)) (p!2 (Pair Int Int))) Bool (and (pair_lte p!1 p!2) (not (= p!1 p!2))))\n" +
			"(define-fun pair_gte ((p!1 (Pair Int Int)) (p!2 (Pair Int Int))) Bool (pair_lte p!2 p!1))\n" +
			"(define-fun pair_gt ((p!1 (Pair Int Int)) (p!2 (Pair Int Int))) Bool (pair_lt p!2 p!1))\n" +
			"(define-fun pair_plus ((p!1 (Pair Int Int)) (p!2 (Pair Int Int))) (Pair Int Int) (mk-pair (+ (fst p!1) (fst p!2)) (+ (snd p!1) (snd p!2))))\n" +
			"(define-fun pair_sub ((p!1 (Pair Int Int)) (p!2 (Pair Int Int))) (Pair Int Int) (mk-pair (- (fst p!1) (fst p!2)) (- (snd p!1) (snd p!2))))\n" +
			smt2;
	smt2 = smt2 + "\n(check-sat)\n(exit)";

	file, err := ioutil.TempFile("", "scribble-go.*.smt2")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())

	w := bufio.NewWriter(file)
	if _, err = w.WriteString(smt2); err != nil {
		panic(err)
	}
	if err = w.Flush(); err != nil {
		panic(err)
	}

	var cmdOut []byte
	cmdName := "z3"
	cmdArgs := []string{file.Name()}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		//fmt.Fprintln(os.Stderr, "Error: ", err)
		panic(err)
	}
	res := strings.Trim(string(cmdOut), "\r\n")
	if res == "sat" {
		return true	
	} else if res == "unsat" {
		return false
	} else {
		panic(res)
	}
}
