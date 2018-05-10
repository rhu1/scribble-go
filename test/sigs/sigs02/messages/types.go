package messages

// Data is the default payload type.
type Data int  // N.B. testing primitive 

func (Data) GetOp() string {
	return "Data"
}
