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

func (ep *Endpoint) SendString(rolename string, i int, msg string) error {
	return ep.Fmts[rolename][i].EncodeString(msg)
}

func (ep *Endpoint) RecvString(rolename string, i int, msg *string) error {
	tmp, err := ep.Fmts[rolename][i].DecodeString()
	if err == nil {
		*msg = tmp
	}
	return err
}

func (ep *Endpoint) SendInt(rolename string, i int, msg int) error {
	return ep.Fmts[rolename][i].EncodeInt(msg)
}

func (ep *Endpoint) RecvInt(rolename string, i int, msg *int) error {
	tmp, err := ep.Fmts[rolename][i].DecodeInt()
	if err == nil {
		*msg = tmp
	}
	return err
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
