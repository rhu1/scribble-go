package session2

import (
	"encoding/gob"
	"fmt"
	//"sync"
	//"strconv"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

var _ = fmt.Print

func init() {
	gob.Register(wrapper{})	
}

type wrapper struct {
	Msg interface{}
}

//func (*wrapper) GetOp() string {
func (wrapper) GetOp() string {
	return "_wrapper"	
}

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

// Pre: msg is a pointer value
func (ep *MPChan) ISend(rolename string, i int, msg interface{}) error {
	//fmt.Printf("ISend %v %T\n", msg, msg)
	return ep.MSend(rolename, i, wrapper{Msg:msg})  // CHECKME: &wrapper?
}

// Could just use interface{}?  but specify *interface{} as typing info?
// N.B. the "interface{}" part is itself a pointer, cf. ISend
func (ep *MPChan) IRecv(rolename string, i int, msg *interface{}) error {
	var w ScribMessage
	err := ep.MRecv(rolename, i, &w)
	//fmt.Printf("IRecv %v %T \n", w.(wrapper).Msg, w.(wrapper).Msg)
	if err == nil {
		*msg = w.(wrapper).Msg
	}
	return err
}

// *Foo is coming here
func (ep *MPChan) MSend(rolename string, i int, msg ScribMessage) error {
	//fmt.Printf("MSend %v %T %v %T\n", msg, msg, msg.(wrapper).Msg, msg.(wrapper).Msg)
	return ep.Fmts[rolename][i].Serialize(msg)
}

func (ep *MPChan) MRecv(rolename string, i int, msg *ScribMessage) error {
	err := ep.Fmts[rolename][i].Deserialize(msg)

	//fmt.Printf("MRecv %v %T\n", *msg, *msg)//, (*msg).(wrapper).Msg, (*msg).(wrapper).Msg)

	return err
//MSend {0xc0420ea080} session2.wrapper 0xc0420ea080 *string
//MRecv {A} session2.wrapper A string  // XXX gob converts Msg pointer field to non-pointer?
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

/*// Or could make ScribMessage wrappers...
func (ep *MPChan) SendString(rolename string, i int, msg string) error {
	return ep.SendBytes(rolename, i, []byte(msg))
}

func (ep *MPChan) RecvString(rolename string, i int, msg *string) error {
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
}*/
