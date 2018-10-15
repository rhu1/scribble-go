package util

import "github.com/rhu1/scribble-go-runtime/runtime/twodim/session2"

type IntInterval struct {
	Start	int
	End int
}

func (ival IntInterval) IsEmpty() bool {
	return ival.Start > ival.End;
}

type IntPairInterval struct {
	Start	session2.Pair
	End session2.Pair
}

func (ival IntPairInterval) IsEmpty() bool {
	return ival.Start.Gt(ival.End);
}

func IsectIntIntervals(ivals []IntInterval) IntInterval {
	return IntInterval{1,1}	
}

func checkFamily() {
	
}
