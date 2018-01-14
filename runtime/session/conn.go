package session

import (
	"fmt"
	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"log"
)

type ParamRole struct {
	Name  string
	Param int
}

//func NewConn(roles []ParamRole) (map[string][]transport.Channel, error) {
func NewConn(roles []ParamRole) (map[string]map[int]transport.Channel, error) {
	//conn := make(map[string][]transport.Channel)
	conn := make(map[string]map[int]transport.Channel)
	for _, r := range roles {
		if r.Param < 1 {
			return nil, fmt.Errorf("Invalid parameter of role '%s': '%d'", r.Name, r.Param)
		}
		//conn[r.Name] = make([]transport.Channel, r.Param)
		conn[r.Name] = make(map[int]transport.Channel)
	}
	return conn, nil
}

func RoleRange(id, nw int) {
	if id < 1 || id > nw {
		log.Panicf("Role id '%d' not in range [1,%d]", id, nw)
	}
}
