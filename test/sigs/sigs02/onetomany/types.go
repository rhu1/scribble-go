package onetomany

// Data is the default payload type.
type Data int

func (Data) GetOp() string {
	return "Data"
}
