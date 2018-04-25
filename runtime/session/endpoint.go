package session

import (
	"sync"

	"github.com/rhu1/scribble-go-runtime/runtime/transport"
)

/* // FIXME: deprecate -- use Endpoint type directly in state chans again
type ParamEndpoint interface {
	Ept() *Endpoint
	Params() map[string]int
	//IsParamEndpoint()
}*/


type Endpoint struct {
	Id       int
	//NumRoles int

	connWg sync.WaitGroup // Counts initiated connections from this Endpoint.

	// guards Conn
	ConnMu sync.RWMutex
	Conn   map[string](map[int]transport.Channel)
	Fmt    map[string](map[int]ScribMessageFormatter)
}

//func NewEndpoint(roleId int, conn map[string]map[int]transport.Channel) *Endpoint {
func NewEndpoint(self int, rolenames []string) *Endpoint {
	conns := make(map[string]map[int]transport.Channel)
	fmts := make(map[string]map[int]ScribMessageFormatter)
	for _, r := range rolenames {
		conns[r] = make(map[int]transport.Channel)
		fmts[r] = make(map[int]ScribMessageFormatter)
	}
	return &Endpoint{
		Id:   self,
		Conn: conns,
		Fmt:  fmts,
	}
}

func (ep *Endpoint) Send(rolename string, i int, msg *ScribMessage) error {
	bs, err := ep.Fmt[rolename][i].ToBytes(msg)
	if err != nil {
		return err	
	}
	err = ep.Conn[rolename][i].ScribWrite(bs)
	return err
}

func (ep *Endpoint) Recv(rolename string, i int, msg *ScribMessage) error {
	var bs []byte
	var tmp *ScribMessage 
	var err error
	for tmp == nil && err == nil {
		ep.Conn[rolename][i].ScribRead(&bs)
		tmp, err = ep.Fmt[rolename][i].FromBytes(bs)
	}
	*msg = *tmp
	return err
}

// CheckConnection ensures connections initiated (by Accept)
// in Endpoint e are fully established.
func (e *Endpoint) CheckConnection() {
	e.connWg.Wait()
}
