package messages


type Bar struct {
	Y string 
}

func (Bar) GetOp() string {
	return "Bar"	
}
