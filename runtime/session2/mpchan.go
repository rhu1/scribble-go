package session2

import (
	"fmt"
	//"sync"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

var _ = fmt.Print

type MPChan struct {
	Fmts    map[string](map[int]ScribMessageFormatter)
	Conns   map[string](map[int]transport2.BinChannel)  // Don't need to export, wrapped by Fmts
}

func NewMPChan(self int, rolenames []string) *MPChan {
	fmts := make(map[string]map[int]ScribMessageFormatter)
	conns := make(map[string]map[int]transport2.BinChannel)
	for _, r := range rolenames {
		conns[r] = make(map[int]transport2.BinChannel)
		fmts[r] = make(map[int]ScribMessageFormatter)
	}
	return &MPChan{
		Fmts:  fmts,
		Conns: conns,
	}
}

func (ep *MPChan) SendString(rolename string, i int, msg string) error {
	return ep.Fmts[rolename][i].EncodeString(msg)
}

func (ep *MPChan) RecvString(rolename string, i int, msg *string) error {
	tmp, err := ep.Fmts[rolename][i].DecodeString()
	if err == nil {
		*msg = tmp
	}
	return err
}

func (ep *MPChan) SendInt(rolename string, i int, msg int) error {
	return ep.Fmts[rolename][i].EncodeInt(msg)
}

func (ep *MPChan) RecvInt(rolename string, i int, msg *int) error {
	tmp, err := ep.Fmts[rolename][i].DecodeInt()
	if err == nil {
		*msg = tmp
	}
	return err
}

func (ep *MPChan) Send(rolename string, i int, msg ScribMessage) error {
	/*bs, err := ep.Fmts[rolename][i].ToBytes(msg)
	if err != nil {
		return err	
	}
	err = ep.Conns[rolename][i].Write(bs)
	return err*/
	return ep.Fmts[rolename][i].Serialize(msg)
}

func (ep *MPChan) Recv(rolename string, i int, msg *ScribMessage) error {
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

func (e *MPChan) Close() error {
	var err error
	for _, cs := range e.Conns {
		for _, c := range cs {
			if e := c.Close(); err == nil && e != nil {
				err = e	
			}
		}
	}
	return err
}

func (e *MPChan) CheckConnection() {
	//...TODO
}
