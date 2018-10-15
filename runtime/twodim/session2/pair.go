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
	//return p1.X <= p2.X && p1.Y <= p2.Y
	return p1.X < p2.X || (p1.X == p2.X && p1.Y <= p2.Y)
}

func (p1 Pair) Eq(p2 Pair) bool {
	return p1.X == p2.X && p1.Y == p2.Y;	
}

func (p1 Pair) Lt(p2 Pair) bool {
	return p1.Lte(p2) && !p1.Eq(p2);	
}

func (p1 Pair) Gt(p2 Pair) bool {
	return !p1.Lte(p2);	
}

func (p1 Pair) Gte(p2 Pair) bool {
	return !p1.Lt(p2);	
}

func (p Pair) Inc(max Pair) Pair {
	if p.Y < max.Y {
		return XY(p.X, p.Y+1)
	} else {
		return XY(p.X+1, 1)
	}
}

// Flatten converts p into an ordinal number.
//
// Couting starts from (1,1) and stays within
// the bounds of (1,1) and max. An example of
// counting order is:
//
// (1,1), (1,2), (1,3), ..., (2,1), ...
//
func (p Pair) Flatten(max Pair) int {
	return ((p.X - 1) * max.Y) + p.Y
}

// String returns the string representation.
func (p Pair) String() string {
	return "(" + strconv.Itoa(p.X) + ", " + strconv.Itoa(p.Y) + ")"
}

func (p1 Pair) Plus(p2 Pair) Pair {
	return XY(p1.X+p2.X, p1.Y+p2.Y)
}

func (p1 Pair) Sub(p2 Pair) Pair {
	return XY(p1.X-p2.X, p1.Y-p2.Y)
}
