package session2

import (
	"fmt"
	//"sync"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

var _ = fmt.Print

type Endpoint struct {
	Self    int  // N.B. generated Endpoint also has a Self field
	Fmts    map[string](map[int]ScribMessageFormatter)
	Conns   map[string](map[int]transport2.BinChannel)
}

func NewEndpoint(self int, rolenames []string) *Endpoint {
	fmts := make(map[string]map[int]ScribMessageFormatter)
	conns := make(map[string]map[int]transport2.BinChannel)
	for _, r := range rolenames {
		conns[r] = make(map[int]transport2.BinChannel)
		fmts[r] = make(map[int]ScribMessageFormatter)
	}
	return &Endpoint{
		Self:   self,
		Fmts:  fmts,
		Conns: conns,
	}
}

func (ep *Endpoint) Send(rolename string, i int, msg ScribMessage) error {
	/*bs, err := ep.Fmts[rolename][i].ToBytes(msg)
	if err != nil {
		return err	
	}
	err = ep.Conns[rolename][i].Write(bs)
	return err*/
	return ep.Fmts[rolename][i].Serialize(msg)
}

func (ep *Endpoint) Recv(rolename string, i int, msg *ScribMessage) error {
	/*var bs []byte
	var tmp *ScribMessage 
	var err error
	for tmp == nil && err == nil {
		ep.Conns[rolename][i].Read(bs)
		tmp, err = ep.Fmts[rolename][i].FromBytes(bs)
	}
	*msg = *tmp
	return err*/
	tmp, err := ep.Fmts[rolename][i].Deserialize()
	if err == nil {
		*msg = tmp
	}
	return err
}

func (e *Endpoint) CheckConnection() {
	//...TODO
}
