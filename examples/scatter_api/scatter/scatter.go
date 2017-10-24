package scatter

import (
	"fmt"
	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/tcp"
	"log"
)

const Server = "server"
const Worker = "worker"

type Server_1To1_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func NewServer(id, nserver, nworker int) (*Server_1To1_Init, error) {
	if id > nserver || id < 1 {
		return nil, fmt.Errorf("'server' id not in range [1, %d]", nserver)
	}
	if nworker < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'worker': %d", nworker)
	}
	conn := make(map[string][]transport.Channel)
	conn[Worker] = make([]transport.Channel, nworker)

	return &Server_1To1_Init{session.LinearResource{}, &session.Endpoint{id, nserver, conn}}, nil
}

func (ini *Server_1To1_Init) Accept(rolename string, id int, addr, port string) error {
	cn, ok := ini.ept.Conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", id, rolename)
	}
	go func(i int, addr, port string) {
		ini.ept.Conn[rolename][i-1] = tcp.NewConnection(addr, port).Accept().(*tcp.Conn)
	}(id, addr, port)
	return nil
}

type Server_1To1_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
// For the server, wait until a connection for each participant is available.
// FIXME: inefficient spinlock.
// TODO: Assumption, rolename and
func (ini *Server_1To1_Init) Init() (*Server_1To1_1, error) {
	ini.Use()
	conn := ini.ept.Conn[Worker]
	n_worker := len(conn)

	// FIXME
	for i := 0; i < n_worker; i++ {
		for ini.ept.Conn[Worker][i] == nil {
		}
	}

	return &Server_1To1_1{session.LinearResource{}, ini.ept}, nil
}

type Server_1To1_End struct {
}

// Session has started, so if an error occurs, then a runtime error is produced
// and the program exits
func (st1 *Server_1To1_1) SendAll(pl []int) *Server_1To1_End {
	if len(pl) != len(st1.ept.Conn[Worker]) {
		log.Fatalf("sending wrong number of arguments in 'st1': %d != %d", len(st1.ept.Conn[Worker]), len(pl))
	}
	st1.Use()

	for i, v := range pl {
		st1.ept.Conn[Worker][i].Send(v)
	}
	return &Server_1To1_End{}
}

// Convenience to check that user implements the full protocol
func (ini *Server_1To1_Init) Run(f func(*Server_1To1_1) *Server_1To1_End) {

	st1, err := ini.Init()

	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}

	f(st1)
}

type Worker_1Ton_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func NewWorker(id, nworker, nserver int) (*Worker_1Ton_Init, error) {
	if id > nworker || id < 1 {
		return nil, fmt.Errorf("'worker' id not in range [1, %d]", nworker)
	}
	if nserver < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'server': %d", nserver)
	}
	conn := make(map[string][]transport.Channel)
	conn[Server] = make([]transport.Channel, nserver)

	return &Worker_1Ton_Init{session.LinearResource{}, session.NewEndpoint(id, nworker, conn)}, nil
}

func (ini *Worker_1Ton_Init) Connect(rolename string, id int, addr, port string) error {
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

type Worker_1Ton_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
func (ini *Worker_1Ton_Init) Init() (*Worker_1Ton_1, error) {
	n_server := len(ini.ept.Conn[Server])
	for i := 0; i < n_server; i++ {
		if ini.ept.Conn[Server][i] == nil {
			return nil, fmt.Errorf("invalid connection from 'worker[%d]' to 'server[%d]'", ini.ept.Id, i)
		}
	}
	return &Worker_1Ton_1{session.LinearResource{}, ini.ept}, nil
}

type Worker_1Ton_End struct {
}

func (st1 *Worker_1Ton_1) RecvAll() ([]int, *Worker_1Ton_End) {
	var tmp int
	st1.Use()

	res := make([]int, len(st1.ept.Conn[Server]))
	for i, conn := range st1.ept.Conn[Server] {
		err := conn.Recv(&tmp)
		if err != nil {
			log.Fatalf("wrong value from server at %d: %s", st1.ept.Id, err)
		}
		res[i] = tmp
	}
	return res, &Worker_1Ton_End{}
}

func (ini *Worker_1Ton_Init) Run(f func(*Worker_1Ton_1) *Worker_1Ton_End) {
	st1, err := ini.Init()
	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}
	f(st1)
}
