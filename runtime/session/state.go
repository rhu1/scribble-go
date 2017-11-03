package session

import (
	"fmt"
	"github.com/nickng/scribble-go/runtime/transport"
)

type PreState interface {
	Ept() *Endpoint
}

func Accept(ini PreState, rolename string, id int, conn transport.Transport) error {
	cn, ok := ini.Ept().Conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", id, rolename)
	}
	go func(i int, conn transport.Transport) {
		ini.Ept().Conn[rolename][i-1] = conn.Accept()
	}(id, conn)
	return nil
}

func Connect(ini PreState, rolename string, id int, conn transport.Transport) error {
	cn, ok := ini.Ept().Conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", id, rolename)
	}
	// Probably a good idea to use tcp.NewConnectionWithRetry
	ini.Ept().Conn[rolename][id-1] = conn.Connect()
	return nil
}
