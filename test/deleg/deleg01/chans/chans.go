package chans


import A "github.com/rhu1/scribble-go-runtime/test/deleg/deleg01/Deleg1/Proto2/A_1to1"


// Deprecated -- delegation is currently implicitly as MessageSig payload
type Foo struct {
	X *A.Init
}

func (Foo) GetOp() string {
	return "Foo"	
}


