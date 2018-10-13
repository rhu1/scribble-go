package session2

// A Pair is a X-Y integer coordinate.
type Pair struct {
	X, Y int
}

// XY creates a pair from an (x,y) coordinate.
func XY(x, y int) Pair {
	return Pair{X: x, Y: y}
}
