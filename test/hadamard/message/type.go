package message

// Val is Val(int) signature.
type Val struct {
	R string
	I int
}

// GetOp returns the label.
func (*Val) GetOp() string { return "Val" }
