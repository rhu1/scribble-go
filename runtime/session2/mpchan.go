package session2

import (
	"fmt"
	//"sync"
	"strconv"

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

// Or could make ScribMessage wrappers...
func (ep *MPChan) SendString(rolename string, i int, msg string) error {
	//return ep.Fmts[rolename][i].EncodeString(msg)
	return ep.SendBytes(rolename, i, []byte(msg))
}

func (ep *MPChan) RecvString(rolename string, i int, msg *string) error {
	/*tmp, err := ep.Fmts[rolename][i].DecodeString()
	if err == nil {
		*msg = tmp
	}
	return err*/
	var bs []byte
	err := ep.RecvBytes(rolename, i, &bs)
	if err == nil {
		*msg = string(bs)
	}
	return err
}

func (ep *MPChan) SendInt(rolename string, i int, msg int) error {
	return ep.SendString(rolename, i, strconv.Itoa(msg))
}

func (ep *MPChan) RecvInt(rolename string, i int, msg *int) error {
	var tmp string
	err := ep.RecvString(rolename, i, &tmp)
	if err == nil {
		*msg, _ = strconv.Atoi(tmp)
	}
	return err
}

func (ep *MPChan) SendBytes(rolename string, i int, bs []byte) error {
	return ep.Fmts[rolename][i].EncodeBytes(bs)
}

func (ep *MPChan) RecvBytes(rolename string, i int, bs *[]byte) error {
	tmp, err := ep.Fmts[rolename][i].DecodeBytes()
	if err == nil {
		*bs = tmp
	}
	return err
}

func (ep *MPChan) Send(rolename string, i int, msg ScribMessage) error {
	return ep.Fmts[rolename][i].Serialize(msg)
}

func (ep *MPChan) Recv(rolename string, i int, msg *ScribMessage) error {
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
