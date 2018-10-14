package session2

import "strconv"

// A Pair is a X-Y integer coordinate.
type Pair struct {
	X, Y int
}

// XY creates a pair from an (x,y) coordinate.
func XY(x, y int) Pair {
	return Pair{X: x, Y: y}
}

func (p1 Pair) Lte(p2 Pair) bool {
	return p1.X <= p2.X && p1.Y <= p2.Y;	
}

func (p Pair) Inc(max Pair) Pair {
	if 	p.Y < max.Y {
		return XY(p.X, p.Y+1)	
	} else {//if p.Y < max.Y {
		return XY(p.X+1, 1)	
	} /*else {
		panic("Bad inc:")  // Inconvenient in for-loops
	}*/
}

// Assumes base default is (1, 1), and count (1,1), (1,2), (1,3), ..., (2,1), ...
func (p Pair) Flatten(max Pair) int {
	return ((p.X-1) * max.Y) + p.Y;
}

func (p Pair) Tostring() string {
	return "(" + strconv.Itoa(p.X) + ", " + strconv.Itoa(p.Y) + ")";
}