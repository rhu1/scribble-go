package util

import "fmt"
import "strconv"
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

type IntPairInterval struct {
	Start	session2.Pair
	End session2.Pair
}

/*func (ival IntPairInterval) String() string {

}*/

func (ival IntPairInterval) IsEmpty() bool {
	return ival.Start.Gt(ival.End);
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
