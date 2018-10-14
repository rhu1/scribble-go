package session2

import (
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

func init() {
	gob.Register(wrapper{})
}

type wrapper struct {
	Msg interface{}
}

func (wrapper) GetOp() string {
	return "_wrapper"
}

// MPChan represents a multiparty channel, it contains
// metadata and the connections to all connected parties.
type MPChan struct {
	// ConnWg tracks initiated but un-established connections.
	// Connections functions (Dial/Accept) must manually call
	// ConnWg.Add(1) to add to the count when initiating a new
	// connection to ensure connections are established before
	// being used by Run(), e.g.
	//
	//     c.MPChan.ConnWg.Add(1)
	//     conn, err := socket.Accept()
	//     // store conn ...
	//     c.MPChan.ConnWg.Done()
	ConnWg sync.WaitGroup

	Fmts  map[string](map[int]ScribMessageFormatter)
	Conns map[string](map[int]transport2.BinChannel)
}

// NewMPChan returns a new initialised multiparty channel instance.
func NewMPChan(self int, rolenames []string) *MPChan {
	fmts := make(map[string]map[int]ScribMessageFormatter)
	conns := make(map[string]map[int]transport2.BinChannel)
	for _, r := range rolenames {
		conns[r] = make(map[int]transport2.BinChannel)
		fmts[r] = make(map[int]ScribMessageFormatter)
	}
	return &MPChan{Fmts: fmts, Conns: conns}
}

// ISend sends a message msg to role r[i].
// The parameter msg should be a pointer type, for example,
//
//     var i int = 42
//     c.ISend("Foo", 1, &i) // sends 42 to Foo[1]
//
func (c *MPChan) ISend(r string, i int, msg interface{}) error {
	return c.MSend(r, i, wrapper{Msg: msg})
}

// IRecv receives a message msg from role r[i].
// The parameter msg should be a pointer type and should be
// allocated, for example,
//
//    var val T
//    c.IRecv("Foo", 2, &val) // receives from Foo[2] into v
//
//    var ptr *T = new(T) // allocate for memory ptr points to
//    c.IRecv("Foo", 2, ptr) // receives from Foo[2] into *ptr
//
func (c *MPChan) IRecv(r string, i int, msg interface{}) error {
	// IRecv uses the underlying MRecv to receive messages
	// since MRecv expects a ScribMessage, a wrapper w of
	// that type is created as a container to temporarily
	// store the msg pointer to cross the function boundary.
	// The wrapper is ignored after receiving the value.
	var w ScribMessage = wrapper{msg}
	return c.MRecv(r, i, &w)
}

// MSend sends a Scribble message msg to role r[i].
func (c *MPChan) MSend(r string, i int, msg ScribMessage) error {
	rmap, hasRole := c.Fmts[r]
	if !hasRole {
		return fmt.Errorf("cannot send: role %s does not exist (typo?)", r)
	}
	fmtrmap, hasFmtr := rmap[i]
	if !hasFmtr {
		return fmt.Errorf("cannot send: role %s[%d] is not connected", r, i)
	}
	return fmtrmap.Serialize(msg)
}

// MRecv receives a Scribble message msg from role r[i].
//
// The Scribble message msg is a pointer to a pre-allocated ScribMessage.
func (c *MPChan) MRecv(r string, i int, msg *ScribMessage) error {
	rmap, hasRole := c.Fmts[r]
	if !hasRole {
		return fmt.Errorf("cannot recv: role %s does not exist (typo?)", r)
	}
	fmtrmap, hasFmtr := rmap[i]
	if !hasFmtr {
		return fmt.Errorf("cannot recv: role %s[%d] is not connected", r, i)
	}
	return fmtrmap.Deserialize(msg)
}

// Close closes all connected channels.
func (c *MPChan) Close() error {
	var err error
	for _, cs := range c.Conns {
		for _, c := range cs {
			if e := c.Close(); err == nil && e != nil {
				err = e
			}
		}
	}
	return err
}

// CheckConnection waits for initiated connection to be established.
func (c *MPChan) CheckConnection() {
	c.ConnWg.Wait()
}
