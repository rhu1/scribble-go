package session

import (
	"fmt"
	"github.com/nickng/scribble-go/runtime/transport/tcp"
)

type Endpoint struct {
	roleId   int
	numRoles int
	conn     map[string][]*tcp.Conn
}

func NewEndpoint(roleId, numRoles int, conn map[string][]*tcp.Conn) *Endpoint {
	return &Endpoint{roleId, numRoles, conn}
}

func (ept *Endpoint) Accept(rolename string, id int, addr, port string) error {
	cn, ok := ept.conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", i, rolename)
	}
	go func(i int, addr, port string) {
		ept.conn[rolename][i-1] = tcp.NewConnection(addr, port).Accept().(*tcp.Conn)
	}(id, addr, port)
	return nil
}

func (ept *Endpoint) Connect(rolename string, id int, addr, port string) error {
	cn, ok := ept.conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", id, rolename)
	}
	// Probably a good idea to use tcp.NewConnectionWithRetry
	ept.conn[rolename][id-1] = tcp.NewConnection(addr, port).Connect().(*tcp.Conn)
	return nil
}
