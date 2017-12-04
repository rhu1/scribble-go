package dot

import (
	"fmt"
	"github.com/nickng/scribble-go-runtime/runtime/session"
	"github.com/nickng/scribble-go-runtime/runtime/transport"
	"github.com/nickng/scribble-go-runtime/runtime/transport/tcp"
	"log"
)

const RoleA = "A"
const RoleB = "B"

type RoleA_1To1_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func NewRoleA(id, nA, nB int) (*RoleA_1To1_Init, error) {
	if id > nA || id < 1 {
		return nil, fmt.Errorf("'A' id not in range [1, %d]", nA)
	}
	if nB < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'B': %d", nB)
	}
	conn := make(map[string][]transport.Channel)
	conn[RoleB] = make([]transport.Channel, nB)

	return &RoleA_1To1_Init{ept: session.NewEndpoint(id, nA, conn)}, nil
}

func (ini *RoleA_1To1_Init) Accept(rolename string, id int, addr, port string) error {
	cn, ok := ini.ept.Conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", id, rolename)
	}
	go func(i int, addr, port string) {
		ini.ept.Conn[rolename][i-1] = tcp.NewConnection(addr, port).Accept()
	}(id, addr, port)
	return nil
}

type RoleA_1To1_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
// For the A, wait until a connection for each participant is available.
// FIXME: inefficient spinlock.
// TODO: Assumption, rolename and
func (ini *RoleA_1To1_Init) Init() (*RoleA_1To1_1, error) {
	ini.Use()
	// FIXME
	for ini.ept.Conn[RoleB][ini.ept.Id-1] == nil {
	}

	return &RoleA_1To1_1{session.LinearResource{}, ini.ept}, nil
}

type RoleA_1To1_End struct {
}

// Session has started, so if an error occurs, then a runtime error is produced
// and the program exits
func (st1 *RoleA_1To1_1) DotSend(v int) *RoleA_1To1_End {
	st1.Use()

	//s1.ept.Id -- self id -- TODO: "-1" is hardcoded left-index 1
	err := st1.ept.Conn[RoleB][st1.ept.Id-1].Send(v)
	if err != nil {
		log.Fatalf("wrong value from A at %d: %s", st1.ept.Id, err)
	}
	return &RoleA_1To1_End{}
}

// Convenience to check that user implements the full protocol
func (ini *RoleA_1To1_Init) Run(f func(*RoleA_1To1_1) *RoleA_1To1_End) {

	st1, err := ini.Init()

	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}

	f(st1)
}

type RoleB_1Ton_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func NewRoleB(id, nB, nA int) (*RoleB_1Ton_Init, error) {
	if id > nB || id < 1 {
		return nil, fmt.Errorf("'B' id not in range [1, %d]", nB)
	}
	if nA < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'A': %d", nA)
	}
	conn := make(map[string][]transport.Channel)
	conn[RoleA] = make([]transport.Channel, nA)

	return &RoleB_1Ton_Init{session.LinearResource{}, session.NewEndpoint(id, nB, conn)}, nil
}

func (ini *RoleB_1Ton_Init) Connect(rolename string, id int, addr, port string) error {
	cn, ok := ini.ept.Conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", id, rolename)
	}
	// Probably a good idea to use tcp.NewConnectionWithRetry
	ini.ept.Conn[rolename][id-1] = tcp.NewConnection(addr, port).Connect()
	return nil
}

type RoleB_1Ton_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
func (ini *RoleB_1Ton_Init) Init() (*RoleB_1Ton_1, error) {
	if ini.ept.Conn[RoleA][ini.ept.Id-1] == nil {
		return nil, fmt.Errorf("invalid connection from 'B[%d]' to 'A[%d]'", ini.ept.Id, ini.ept.Id)
	}
	return &RoleB_1Ton_1{session.LinearResource{}, ini.ept}, nil
}

type RoleB_1Ton_End struct {
}

func (st1 *RoleB_1Ton_1) DotRecv() (int, *RoleB_1Ton_End) {
	st1.Use()
	var tmp int
	st1.ept.Conn[RoleA][st1.ept.Id-1].Recv(&tmp)
	return tmp, &RoleB_1Ton_End{}
}

func (ini *RoleB_1Ton_Init) Run(f func(*RoleB_1Ton_1) *RoleB_1Ton_End) {
	st1, err := ini.Init()
	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}
	f(st1)
}
