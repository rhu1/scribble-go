package messages

type Foo struct {
	X int 
}

func (Foo) GetOp() string {
	return "Foo"	
}
