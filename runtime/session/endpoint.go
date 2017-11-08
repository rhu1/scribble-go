package session

import (
	"sync"

	"github.com/nickng/scribble-go/runtime/transport"
)

type Endpoint struct {
	Id       int
	NumRoles int

	// guards Conn
	ConnMu sync.RWMutex
	Conn   map[string][]transport.Channel
}

func NewEndpoint(roleId, numRoles int, conn map[string][]transport.Channel) *Endpoint {
	return &Endpoint{
		Id:       roleId,
		NumRoles: numRoles,
		Conn:     conn,
	}
}
