package util

import "testing"


/*func TestIntIntervalSub1(t *testing.T) {
	x := util.IntInterval{1,4}
	y := util.IntInterval{2,3} 
	z := x.IntIntervalSub(y)
	if z != util.IntInterval{...}
	t.Error("... ", v)
}*/

func TestIntIntervalSub2(t *testing.T) {
	x := IntInterval{1,3}
	y := IntInterval{2,4}
	z := x.subIntInterval(y)
	if z != (IntInterval{1,1}) {
		t.Error("Expecting [1,1], got: ", z)
	}
}

func TestIntIntervalSub2a(t *testing.T) {
	x := IntInterval{1,4}
	y := IntInterval{3,4}
	z := x.subIntInterval(y)
	if z != (IntInterval{1,2}) {
		t.Error("Expecting [1,2], got: ", z)
	}
}

func TestIntIntervalSub3(t *testing.T) {
	x := IntInterval{1,2}
	y := IntInterval{3,4}
	z := x.subIntInterval(y)
	if z != (IntInterval{1,2}) {
		t.Error("Expecting [1,2], got: ", z)
	}
}

func TestIntIntervalSub3a(t *testing.T) {
	x := IntInterval{1,2}
	y := IntInterval{2,4}
	z := x.subIntInterval(y)
	if z != (IntInterval{1,1}) {
		t.Error("Expecting [1,1], got: ", z)
	}
}

func TestIntIntervalSub4(t *testing.T) {
	x := IntInterval{2,3}
	y := IntInterval{1,4}
	z := x.subIntInterval(y)
	if !z.IsEmpty() {
		t.Error("Expecting empty, got: ", z)
	}
}

func TestIntIntervalSub4a(t *testing.T) {
	x := IntInterval{1,3}
	y := IntInterval{1,4}
	z := x.subIntInterval(y)
	if !z.IsEmpty() {
		t.Error("Expecting empty, got: ", z)
	}
}

func TestIntIntervalSub5(t *testing.T) {
	x := IntInterval{2,4}
	y := IntInterval{1,3}
	z := x.subIntInterval(y)
	if z != (IntInterval{4,4}) {
		t.Error("Expecting [4,4], got: ", z)
	}
}

func TestIntIntervalSub5a(t *testing.T) {
	x := IntInterval{3,4}
	y := IntInterval{1,3}
	z := x.subIntInterval(y)
	if z != (IntInterval{4,4}) {
		t.Error("Expecting [4,4], got: ", z)
	}
}

func TestIntIntervalSub6(t *testing.T) {
	x := IntInterval{3,4}
	y := IntInterval{1,2}
	z := x.subIntInterval(y)
	if z != (IntInterval{3, 4}) {
		t.Error("Expecting [3,4], got: ", z)
	}
}
