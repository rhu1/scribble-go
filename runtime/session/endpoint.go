package session

import (
	"github.com/nickng/scribble-go/runtime/transport"
)

type Endpoint struct {
	Id       int
	numRoles int
	Conn     map[string][]*transport.Channel
}

func NewEndpoint(roleId, numRoles int, conn map[string][]*transport.Channel) *Endpoint {
	return &Endpoint{roleId, numRoles, conn}
}
