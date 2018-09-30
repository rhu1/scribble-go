package messages

type Foo struct {
	X int 
}

func (Foo) GetOp() string {
	return "Foo"	
}

type Bar struct {
	Y string 
}

func (Bar) GetOp() string {
	return "Bar"	
}
