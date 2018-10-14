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
	if 	p.X < max.X {
		return XY(p.X+1, p.Y)	
	} else {//if p.Y < max.Y {
		return XY(1, p.Y+1)	
	} /*else {
		panic("Bad inc:")  // Inconvenient in for-loops
	}*/
}

func (p Pair) Flatten(max Pair) int {
	return p.X * max.Y + max.Y;
}

func (p Pair) Tostring() string {
	return "(" + strconv.Itoa(p.X) + ", " + strconv.Itoa(p.Y) + ")";
}